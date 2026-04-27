package logging_test

import (
	"context"
	"testing"

	"github.com/opensrehub/otelgokit/pkg/logging"
)

func TestInitLogger(t *testing.T) {
	ctx := context.Background()
	logger, err := logging.InitLogger(ctx, "test-service", "test", "localhost:4317")
	if err != nil {
		t.Fatalf("InitLogger: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestInitLoggerIdempotent(t *testing.T) {
	ctx := context.Background()
	l1, err1 := logging.InitLogger(ctx, "test-service", "test", "localhost:4317")
	l2, err2 := logging.InitLogger(ctx, "test-service", "test", "localhost:4317")

	if err1 != nil || err2 != nil {
		t.Fatalf("InitLogger errors: %v, %v", err1, err2)
	}
	if l1 != l2 {
		t.Error("expected the same logger instance on repeated calls")
	}
}

func TestLogFunctions(t *testing.T) {
	ctx := context.Background()
	logging.InitLogger(ctx, "test-service", "test", "localhost:4317")

	// None of these should panic
	logging.LogDebug(ctx, "debug message", "key", "value")
	logging.LogInfo(ctx, "info message", "key", "value")
	logging.LogWarn(ctx, "warn message", "key", "value")
	logging.LogError(ctx, "error message", "key", "value")
}

func TestLogFunctionsBeforeInit(t *testing.T) {
	ctx := context.Background()
	// Logging functions must be safe to call before InitLogger
	logging.LogInfo(ctx, "message without init")
}

func TestShutdownLogger(t *testing.T) {
	ctx := context.Background()
	logging.InitLogger(ctx, "test-service", "test", "localhost:4317")

	if err := logging.ShutdownLogger(ctx); err != nil {
		t.Errorf("ShutdownLogger: %v", err)
	}
}
