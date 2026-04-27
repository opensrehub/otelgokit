package tracing_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/opensrehub/otelgokit/pkg/tracing"
)

func TestInitTracer(t *testing.T) {
	ctx := context.Background()
	tr, tp, err := tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")
	if err != nil {
		t.Fatalf("InitTracer: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil tracer")
	}
	if tp == nil {
		t.Fatal("expected non-nil tracer provider")
	}
	defer tp.Shutdown(ctx)
}

func TestInitTracerIdempotent(t *testing.T) {
	ctx := context.Background()
	tr1, tp1, err1 := tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")
	tr2, tp2, err2 := tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")

	if err1 != nil || err2 != nil {
		t.Fatalf("InitTracer errors: %v, %v", err1, err2)
	}
	if tr1 == nil || tr2 == nil {
		t.Fatal("expected non-nil tracers")
	}
	if tp1 != tp2 {
		t.Error("expected the same provider instance on repeated calls")
	}
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")

	ctx, span := tracing.StartSpan(ctx, "test-span")
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	defer span.End()

	if !span.SpanContext().IsValid() {
		t.Error("expected valid span context")
	}
}

func TestRecordError(t *testing.T) {
	ctx := context.Background()
	tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")

	_, span := tracing.StartSpan(ctx, "test-span")
	defer span.End()

	tracing.RecordError(span, errors.New("test error"))
	tracing.RecordError(span, nil) // nil must be a no-op
}

func TestInjectExtractContext(t *testing.T) {
	ctx := context.Background()
	tracing.InitTracer(ctx, "test-service", "test", "localhost:4317")

	ctx, span := tracing.StartSpan(ctx, "parent-span")
	defer span.End()

	carrier := make(mapCarrier)
	tracing.InjectContext(ctx, carrier)

	if len(carrier) == 0 {
		t.Error("expected trace context to be injected into carrier")
	}

	extracted := tracing.ExtractContext(context.Background(), carrier)
	if extracted == nil {
		t.Fatal("expected non-nil extracted context")
	}
}

func TestNewTransport(t *testing.T) {
	transport := tracing.NewTransport(http.DefaultTransport)
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestNewHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := tracing.NewHandler(handler, "test-op")
	if wrapped == nil {
		t.Fatal("expected non-nil handler")
	}
}

// mapCarrier implements propagation.TextMapCarrier for testing.
type mapCarrier map[string]string

func (m mapCarrier) Get(key string) string  { return m[key] }
func (m mapCarrier) Set(key, val string)    { m[key] = val }
func (m mapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
