package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// NewCounter creates a named Int64Counter.
//
// Use for values that only go up: request counts, processed items, bytes sent.
func NewCounter(meter metric.Meter, name, description string) (metric.Int64Counter, error) {
	c, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("create counter %q: %w", name, err)
	}
	return c, nil
}

// NewHistogram creates a named Float64Histogram with the given unit (e.g. "s", "By", "ms").
//
// Use for distributions: latency, payload sizes, queue wait times.
func NewHistogram(meter metric.Meter, name, description, unit string) (metric.Float64Histogram, error) {
	h, err := meter.Float64Histogram(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("create histogram %q: %w", name, err)
	}
	return h, nil
}

// NewUpDownCounter creates a named Int64UpDownCounter.
//
// Use for values that can increase and decrease: active connections, queue depth, cache size.
func NewUpDownCounter(meter metric.Meter, name, description string) (metric.Int64UpDownCounter, error) {
	c, err := meter.Int64UpDownCounter(name, metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("create updown counter %q: %w", name, err)
	}
	return c, nil
}

// NewObservableGauge registers a Float64ObservableGauge with a callback that is invoked
// on each collection cycle. The callback must be non-blocking.
//
// Use for values that are expensive or impractical to track inline: goroutine count,
// memory stats, external resource states.
func NewObservableGauge(meter metric.Meter, name, description string, cb func(context.Context, metric.Float64Observer) error) error {
	_, err := meter.Float64ObservableGauge(name,
		metric.WithDescription(description),
		metric.WithFloat64Callback(cb),
	)
	if err != nil {
		return fmt.Errorf("create observable gauge %q: %w", name, err)
	}
	return nil
}
