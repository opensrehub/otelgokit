package logging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/go-logr/stdr"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/opensrehub/otelgokit/internal/resource"
)

var (
	loggerInstance *slog.Logger
	loggerProvider *log.LoggerProvider
	initOnce       sync.Once
)

// InitLogger initializes the global slog-backed OTel logger.
// It is safe to call multiple times; initialization runs only once.
func InitLogger(ctx context.Context, serviceName, serviceEnv, collectorEndpoint string) (*slog.Logger, error) {
	var err error
	initOnce.Do(func() {
		exporter, expErr := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(collectorEndpoint),
			otlploggrpc.WithInsecure(),
		)
		if expErr != nil {
			err = fmt.Errorf("new otlp log grpc exporter: %w", expErr)
			return
		}

		res, resErr := resource.Init(ctx, serviceName, serviceEnv)
		if resErr != nil {
			err = fmt.Errorf("init resource: %w", resErr)
			return
		}

		loggerProvider = log.NewLoggerProvider(
			log.WithResource(res),
			log.WithProcessor(log.NewBatchProcessor(exporter)),
		)

		handler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider))
		loggerInstance = slog.New(handler)

		otel.SetLogger(stdr.NewWithOptions(
			NewLoggerAdapter(loggerInstance),
			stdr.Options{LogCaller: stdr.All},
		))
	})
	return loggerInstance, err
}

// LogDebug logs at DEBUG level.
func LogDebug(ctx context.Context, msg string, args ...any) {
	if loggerInstance != nil {
		loggerInstance.DebugContext(ctx, msg, args...)
	}
}

// LogInfo logs at INFO level.
func LogInfo(ctx context.Context, msg string, args ...any) {
	if loggerInstance != nil {
		loggerInstance.InfoContext(ctx, msg, args...)
	}
}

// LogWarn logs at WARN level.
func LogWarn(ctx context.Context, msg string, args ...any) {
	if loggerInstance != nil {
		loggerInstance.WarnContext(ctx, msg, args...)
	}
}

// LogError logs at ERROR level.
func LogError(ctx context.Context, msg string, args ...any) {
	if loggerInstance != nil {
		loggerInstance.ErrorContext(ctx, msg, args...)
	}
}

// ShutdownLogger flushes and shuts down the global LoggerProvider.
func ShutdownLogger(ctx context.Context) error {
	if loggerProvider != nil {
		return loggerProvider.Shutdown(ctx)
	}
	return nil
}
