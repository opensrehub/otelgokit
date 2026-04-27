package tracing

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/opensrehub/otelgokit/internal/resource"
)

var (
	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	initOnce       sync.Once
)

// InitTracer initializes the global TracerProvider and returns the tracer and provider.
// It is safe to call multiple times; initialization runs only once.
func InitTracer(ctx context.Context, serviceName, serviceEnv, collectorEndpoint string) (trace.Tracer, *sdktrace.TracerProvider, error) {
	var err error
	initOnce.Do(func() {
		exporter, expErr := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(collectorEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if expErr != nil {
			err = fmt.Errorf("new otlp trace grpc exporter: %w", expErr)
			return
		}

		res, resErr := resource.Init(ctx, serviceName, serviceEnv)
		if resErr != nil {
			err = fmt.Errorf("init resource: %w", resErr)
			return
		}

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		tracer = tracerProvider.Tracer(serviceName)
	})
	if err != nil {
		return nil, nil, err
	}
	return tracer, tracerProvider, nil
}

// StartSpan starts a new span using the global tracer.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// InjectContext injects the trace context from ctx into the outgoing carrier.
func InjectContext(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// ExtractContext extracts a trace context from the incoming carrier into a new context.
func ExtractContext(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// RecordError records err on span and sets the span status to Error.
// It is a no-op when err is nil.
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// ShutdownTracerProvider flushes and shuts down the global TracerProvider.
func ShutdownTracerProvider(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}

// NewTransport returns an HTTP RoundTripper that injects trace context into outgoing requests.
func NewTransport(base http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(base)
}

// NewHandler wraps handler with OTel middleware that extracts trace context and records spans.
func NewHandler(handler http.Handler, operation string, opts ...otelhttp.Option) http.Handler {
	return otelhttp.NewHandler(handler, operation, opts...)
}
