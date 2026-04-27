package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/opensrehub/otelgokit/internal/resource"
)

var (
	meterProvider *sdkmetric.MeterProvider
	initOnce      sync.Once
)

// InitMeter initializes the global MeterProvider (with Go runtime metrics) and returns
// a Meter scoped to serviceName. It is safe to call multiple times; initialization runs only once.
func InitMeter(ctx context.Context, serviceName, serviceEnv, collectorEndpoint string) (metric.Meter, *sdkmetric.MeterProvider, error) {
	var err error
	initOnce.Do(func() {
		exporter, expErr := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(collectorEndpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if expErr != nil {
			err = fmt.Errorf("new otlp metric grpc exporter: %w", expErr)
			return
		}

		res, resErr := resource.Init(ctx, serviceName, serviceEnv)
		if resErr != nil {
			err = fmt.Errorf("init resource: %w", resErr)
			return
		}

		periodicReader := sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithProducer(runtime.NewProducer()),
		)
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(periodicReader),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(meterProvider)

		err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	})
	if err != nil {
		return nil, nil, err
	}
	return meterProvider.Meter(serviceName), meterProvider, nil
}

// ShutdownMeterProvider flushes and shuts down the global MeterProvider.
func ShutdownMeterProvider(ctx context.Context) error {
	if meterProvider != nil {
		return meterProvider.Shutdown(ctx)
	}
	return nil
}
