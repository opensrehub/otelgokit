package resource

import (
	"context"
	"sync"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

var (
	once     sync.Once
	instance *sdkresource.Resource
)

// Init returns the shared SDK resource, initializing it on the first call.
// Subsequent calls return the cached instance regardless of the arguments passed.
func Init(ctx context.Context, serviceName, serviceEnv string) (*sdkresource.Resource, error) {
	var err error
	once.Do(func() {
		extra, resErr := sdkresource.New(
			ctx,
			sdkresource.WithOS(),
			sdkresource.WithProcess(),
			sdkresource.WithContainer(),
			sdkresource.WithHost(),
			sdkresource.WithAttributes(
				semconv.ServiceName(serviceName),
				semconv.DeploymentEnvironmentName(serviceEnv),
			),
		)
		if resErr != nil {
			err = resErr
			return
		}
		instance, err = sdkresource.Merge(sdkresource.Default(), extra)
	})
	return instance, err
}
