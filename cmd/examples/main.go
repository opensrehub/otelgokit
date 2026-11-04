package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opensrehub/otelgokit/cmd/examples/config"
	"github.com/opensrehub/otelgokit/cmd/examples/handlers"
	"github.com/opensrehub/otelgokit/pkg/logging"
	"github.com/opensrehub/otelgokit/pkg/metrics"
	"github.com/opensrehub/otelgokit/pkg/tracing"
)

const shutdownTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.LoadConfig()

	// Logging — initialize first so subsequent errors are also exported
	_, err := logging.InitLogger(ctx, cfg.ServiceName, cfg.Environment, cfg.CollectorURL)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	// Tracing
	_, tp, err := tracing.InitTracer(ctx, cfg.ServiceName, cfg.Environment, cfg.CollectorURL)
	if err != nil {
		return fmt.Errorf("initialize tracer: %w", err)
	}

	// Metrics
	meter, mp, err := metrics.InitMeter(ctx, cfg.ServiceName, cfg.Environment, cfg.CollectorURL)
	if err != nil {
		return fmt.Errorf("initialize meter: %w", err)
	}

	// Use a fresh, bounded context because ctx is canceled by the shutdown signal.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		mustShutdown("meter", func() error { return mp.Shutdown(shutdownCtx) })
		mustShutdown("tracer", func() error { return tp.Shutdown(shutdownCtx) })
		mustShutdown("logger", func() error { return logging.ShutdownLogger(shutdownCtx) })
	}()

	// Handlers
	h, err := handlers.New(meter)
	if err != nil {
		return fmt.Errorf("initialize handlers: %w", err)
	}

	// Routes:
	//   GET /        — basic hello: tracing, logging, metrics
	//   GET /slow    — variable latency: demonstrates histograms and LogWarn
	//   GET /fail    — always errors: demonstrates RecordError
	//   GET /health  — liveness probe (no telemetry overhead)
	mux := http.NewServeMux()
	mux.Handle("GET /", tracing.NewHandler(h.Hello(), "hello"))
	mux.Handle("GET /slow", tracing.NewHandler(h.Slow(), "slow"))
	mux.Handle("GET /fail", tracing.NewHandler(h.Fail(), "fail"))
	mux.Handle("GET /health", h.Health())

	addr := fmt.Sprintf(":%d", cfg.Port)
	logging.LogInfo(ctx, "server starting",
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Environment),
		slog.String("addr", addr),
	)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		logging.LogInfo(context.Background(), "server shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}

// mustShutdown calls fn on defer and logs any error rather than panicking,
// ensuring all providers attempt a clean flush even if one fails.
func mustShutdown(name string, fn func() error) {
	if err := fn(); err != nil {
		log.Printf("shutdown %s: %v", name, err)
	}
}
