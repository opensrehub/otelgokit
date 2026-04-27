package metrics_test

import (
	"context"
	"testing"

	"github.com/opensrehub/otelgokit/pkg/metrics"
)

func TestInitMeter(t *testing.T) {
	ctx := context.Background()
	meter, mp, err := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")
	if err != nil {
		t.Fatalf("InitMeter: %v", err)
	}
	if meter == nil {
		t.Fatal("expected non-nil meter")
	}
	if mp == nil {
		t.Fatal("expected non-nil meter provider")
	}
	defer mp.Shutdown(ctx)
}

func TestInitMeterIdempotent(t *testing.T) {
	ctx := context.Background()
	meter1, mp1, err1 := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")
	meter2, mp2, err2 := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")

	if err1 != nil || err2 != nil {
		t.Fatalf("InitMeter errors: %v, %v", err1, err2)
	}
	if meter1 == nil || meter2 == nil {
		t.Fatal("expected non-nil meters")
	}
	if mp1 != mp2 {
		t.Error("expected the same provider instance on repeated calls")
	}
}

func TestMeterInstruments(t *testing.T) {
	ctx := context.Background()
	meter, _, err := metrics.InitMeter(ctx, "test-service", "test", "localhost:4317")
	if err != nil {
		t.Fatalf("InitMeter: %v", err)
	}

	counter, err := meter.Int64Counter("test.counter")
	if err != nil {
		t.Fatalf("Int64Counter: %v", err)
	}
	counter.Add(ctx, 1)

	hist, err := meter.Float64Histogram("test.histogram")
	if err != nil {
		t.Fatalf("Float64Histogram: %v", err)
	}
	hist.Record(ctx, 0.5)
}
