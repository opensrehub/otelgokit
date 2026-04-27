package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	goruntime "runtime"
	"time"

	"github.com/opensrehub/otelgokit/pkg/logging"
	"github.com/opensrehub/otelgokit/pkg/metrics"
	"github.com/opensrehub/otelgokit/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Handlers holds all metric instruments shared across handler methods.
// Four OTel instrument types are demonstrated:
//
//	Counter        — http.server.requests.total, http.server.errors.total
//	Histogram      — http.server.request.duration, http.server.response.size
//	UpDownCounter  — http.server.active_requests
//	ObservableGauge — process.goroutines (registered in New, collected asynchronously)
type Handlers struct {
	requestTotal    metric.Int64Counter
	requestDuration metric.Float64Histogram
	responseSize    metric.Float64Histogram
	activeRequests  metric.Int64UpDownCounter
	errorTotal      metric.Int64Counter
}

// New creates a Handlers instance, registers all metric instruments, and sets up
// the goroutine observable gauge.
func New(meter metric.Meter) (*Handlers, error) {
	requestTotal, err := metrics.NewCounter(meter,
		"http.server.requests.total",
		"Total number of HTTP requests received",
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := metrics.NewHistogram(meter,
		"http.server.request.duration",
		"HTTP request duration",
		"s",
	)
	if err != nil {
		return nil, err
	}

	responseSize, err := metrics.NewHistogram(meter,
		"http.server.response.size",
		"HTTP response body size",
		"By",
	)
	if err != nil {
		return nil, err
	}

	activeRequests, err := metrics.NewUpDownCounter(meter,
		"http.server.active_requests",
		"Number of in-flight HTTP requests currently being processed",
	)
	if err != nil {
		return nil, err
	}

	errorTotal, err := metrics.NewCounter(meter,
		"http.server.errors.total",
		"Total number of HTTP requests that resulted in a server error",
	)
	if err != nil {
		return nil, err
	}

	// ObservableGauge: collected asynchronously on each scrape cycle.
	// The callback must never block — read from memory, not network or disk.
	if err := metrics.NewObservableGauge(meter,
		"process.goroutines",
		"Current number of goroutines in the process",
		func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(float64(goruntime.NumGoroutine()))
			return nil
		},
	); err != nil {
		return nil, err
	}

	return &Handlers{
		requestTotal:    requestTotal,
		requestDuration: requestDuration,
		responseSize:    responseSize,
		activeRequests:  activeRequests,
		errorTotal:      errorTotal,
	}, nil
}

// Hello demonstrates basic span creation, logging, and metric recording.
// Trace context is already extracted by the tracing.NewHandler middleware in main.
func (h *Handlers) Hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := "/"
		attrs := metric.WithAttributes(attribute.String("http.route", route))

		h.activeRequests.Add(r.Context(), 1, attrs)
		defer h.activeRequests.Add(r.Context(), -1, attrs)

		start := time.Now()
		ctx, span := tracing.StartSpan(r.Context(), "hello")
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", route),
		)

		logging.LogInfo(ctx, "handling request", "method", r.Method, "path", r.URL.Path)

		n := writeJSON(w, http.StatusOK, map[string]string{"message": "Hello, OpenTelemetry!"})

		h.requestTotal.Add(ctx, 1, attrs)
		h.requestDuration.Record(ctx, time.Since(start).Seconds(), attrs)
		h.responseSize.Record(ctx, float64(n), attrs)
	}
}

// Slow simulates variable-latency work to demonstrate histogram distributions and warnings.
func (h *Handlers) Slow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := "/slow"
		attrs := metric.WithAttributes(attribute.String("http.route", route))

		h.activeRequests.Add(r.Context(), 1, attrs)
		defer h.activeRequests.Add(r.Context(), -1, attrs)

		start := time.Now()
		ctx, span := tracing.StartSpan(r.Context(), "slow-operation")
		defer span.End()

		delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
		span.SetAttributes(attribute.Int64("simulated.delay_ms", delay.Milliseconds()))

		logging.LogDebug(ctx, "starting slow operation", "delay_ms", delay.Milliseconds())
		time.Sleep(delay)

		if delay > 300*time.Millisecond {
			logging.LogWarn(ctx, "slow response detected", "duration_ms", delay.Milliseconds())
		}

		n := writeJSON(w, http.StatusOK, map[string]any{"delay_ms": delay.Milliseconds()})

		h.requestTotal.Add(ctx, 1, attrs)
		h.requestDuration.Record(ctx, time.Since(start).Seconds(), attrs)
		h.responseSize.Record(ctx, float64(n), attrs)
	}
}

// Fail demonstrates error recording on spans using tracing.RecordError.
func (h *Handlers) Fail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := "/fail"
		attrs := metric.WithAttributes(
			attribute.String("http.route", route),
			attribute.Bool("error", true),
		)

		h.activeRequests.Add(r.Context(), 1, attrs)
		defer h.activeRequests.Add(r.Context(), -1, attrs)

		start := time.Now()
		ctx, span := tracing.StartSpan(r.Context(), "failing-operation")
		defer span.End()

		err := fmt.Errorf("simulated failure")
		tracing.RecordError(span, err)
		logging.LogError(ctx, "operation failed", "error", err)

		n := writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})

		h.requestTotal.Add(ctx, 1, attrs)
		h.requestDuration.Record(ctx, time.Since(start).Seconds(), attrs)
		h.responseSize.Record(ctx, float64(n), attrs)
		h.errorTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("http.route", route)))
	}
}

// Health is a lightweight liveness probe that intentionally skips telemetry.
func (h *Handlers) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// writeJSON encodes v as JSON and writes it to w, returning the number of bytes written.
func writeJSON(w http.ResponseWriter, status int, v any) int {
	b, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	n, _ := w.Write(b)
	return n
}
