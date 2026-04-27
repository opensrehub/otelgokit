package metrics_test

import (
	"context"
	"testing"

	"github.com/opensrehub/otelgokit/pkg/metrics"
	"go.opentelemetry.io/otel/metric"
)

func TestNewCounter(t *testing.T) {
	ctx := context.Background()
	meter, _, _ := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")

	c, err := metrics.NewCounter(meter, "test.counter", "a test counter")
	if err != nil {
		t.Fatalf("NewCounter: %v", err)
	}
	c.Add(ctx, 1)
	c.Add(ctx, 5)
}

func TestNewHistogram(t *testing.T) {
	ctx := context.Background()
	meter, _, _ := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")

	h, err := metrics.NewHistogram(meter, "test.histogram", "a test histogram", "s")
	if err != nil {
		t.Fatalf("NewHistogram: %v", err)
	}
	h.Record(ctx, 0.123)
	h.Record(ctx, 1.5)
}

func TestNewUpDownCounter(t *testing.T) {
	ctx := context.Background()
	meter, _, _ := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")

	c, err := metrics.NewUpDownCounter(meter, "test.updown", "a test updown counter")
	if err != nil {
		t.Fatalf("NewUpDownCounter: %v", err)
	}
	c.Add(ctx, 3)
	c.Add(ctx, -1)
}

func TestNewObservableGauge(t *testing.T) {
	ctx := context.Background()
	meter, _, _ := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")

	err := metrics.NewObservableGauge(meter,
		"test.gauge",
		"a test observable gauge",
		func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(42.0)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("NewObservableGauge: %v", err)
	}
}
