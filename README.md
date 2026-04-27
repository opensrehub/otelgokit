# otelgokit

A minimal Go library for bootstrapping OpenTelemetry **tracing**, **metrics**, and **logging** with a single function call per signal.

[![Go Reference](https://pkg.go.dev/badge/github.com/opensrehub/otelgokit.svg)](https://pkg.go.dev/github.com/opensrehub/otelgokit)
[![Go Report Card](https://goreportcard.com/badge/github.com/opensrehub/otelgokit)](https://goreportcard.com/report/github.com/opensrehub/otelgokit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

---

## What it does

`otelgokit` wraps the official [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/) to eliminate the boilerplate of wiring up exporters, resources, and providers. You get a fully configured OTLP gRPC pipeline for all three signals with a few lines of code.

| Signal  | Provider | Exporter |
|---------|----------|----------|
| Traces  | `TracerProvider` | OTLP gRPC |
| Metrics | `MeterProvider` + Go runtime metrics | OTLP gRPC |
| Logs    | `LoggerProvider` via `slog` bridge | OTLP gRPC |

---

## Requirements

- Go 1.23 or later
- A running [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) (or any OTLP-compatible backend such as Jaeger, Tempo, or Datadog Agent)

---

## Installation

> **Note:** Before publishing, update the module path in `go.mod` from the current value to your public module path (e.g. `github.com/opensrehub/otelgokit`). Replace `github.com/opensrehub/otelgokit` in the commands below with your actual module path.

```sh
go get github.com/opensrehub/otelgokit
go mod tidy
```

---

## Quick Start

### Tracing

```go
package main

import (
    "context"
    "log"

    "github.com/opensrehub/otelgokit/pkg/tracing"
)

func main() {
    ctx := context.Background()

    _, tp, err := tracing.InitTracer(ctx, "my-service", "production", "localhost:4317")
    if err != nil {
        log.Fatalf("tracing: %v", err)
    }
    defer tp.Shutdown(ctx)

    ctx, span := tracing.StartSpan(ctx, "my-operation")
    defer span.End()
}
```

### Metrics

```go
package main

import (
    "context"
    "log"

    "github.com/opensrehub/otelgokit/pkg/metrics"
)

func main() {
    ctx := context.Background()

    meter, mp, err := metrics.InitMeter(ctx, "my-service", "production", "localhost:4317")
    if err != nil {
        log.Fatalf("metrics: %v", err)
    }
    defer mp.Shutdown(ctx)

    counter, _ := meter.Int64Counter("requests.total")
    counter.Add(ctx, 1)
}
```

### Logging

```go
package main

import (
    "context"
    "log"

    "github.com/opensrehub/otelgokit/pkg/logging"
)

func main() {
    ctx := context.Background()

    _, err := logging.InitLogger(ctx, "my-service", "production", "localhost:4317")
    if err != nil {
        log.Fatalf("logging: %v", err)
    }
    defer logging.ShutdownLogger(ctx)

    logging.LogInfo(ctx, "service started")
    logging.LogWarn(ctx, "disk space low", "free_gb", 2)
    logging.LogError(ctx, "request failed", "error", err)
}
```

---

## HTTP Middleware

Wrap an HTTP handler to automatically propagate and record trace context:

```go
http.Handle("/", tracing.NewHandler(myHandler, "my-operation"))
```

For outgoing HTTP requests, inject trace context into headers:

```go
import "go.opentelemetry.io/otel/propagation"

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
tracing.InjectContext(ctx, propagation.HeaderCarrier(req.Header))
resp, err := http.DefaultClient.Do(req)
```

For incoming requests, extract trace context from headers:

```go
ctx = tracing.ExtractContext(req.Context(), propagation.HeaderCarrier(req.Header))
ctx, span := tracing.StartSpan(ctx, "handle-request")
defer span.End()
```

---

## Error Recording

Use `RecordError` to attach an error to a span and set its status in one call:

```go
ctx, span := tracing.StartSpan(ctx, "process-order")
defer span.End()

if err := processOrder(ctx, order); err != nil {
    tracing.RecordError(span, err)
    return err
}
```

---

## Running the Example Locally

The repo ships a complete example service with an OpenTelemetry Collector and [otel-tui](https://github.com/ymtdzzz/otel-tui) for live in-terminal telemetry visualization.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

### Start the stack

```sh
docker compose up --build
```

This starts three services:

| Service | Port | Description |
|---------|------|-------------|
| `app` | `3000` | Example HTTP service |
| `otel-collector` | `4317` (gRPC), `4318` (HTTP) | OpenTelemetry Collector |
| `otel-tui` | — | Terminal UI for live telemetry |

The example service exposes four routes:

| Route | What it demonstrates |
|-------|----------------------|
| `GET /` | Basic tracing, logging at INFO, request counter + duration histogram |
| `GET /slow` | Variable latency (100–500 ms), LogWarn on slow responses, histogram distribution |
| `GET /fail` | `tracing.RecordError`, error span status, LogError |
| `GET /health` | Liveness probe with no telemetry overhead |

### Attach to the telemetry UI

```sh
docker compose attach otel-tui
```

### Generate traffic

```sh
bash traffic.sh
```

> Requires [`hey`](https://github.com/rakyll/hey): `go install github.com/rakyll/hey@latest`

---

## Configuration

All three `Init*` functions accept the collector gRPC endpoint as the last argument. You can drive them from environment variables (see `cmd/examples/config/config.go`):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `otel-service` | OTLP `service.name` resource attribute |
| `ENVIRONMENT` | `development` | OTLP `deployment.environment` resource attribute |
| `OTEL_COLLECTOR_URL` | `localhost:4317` | gRPC endpoint of the OTel Collector |
| `PORT` | `3000` | HTTP port the example service listens on |

---

## API Reference

### `pkg/tracing`

| Function | Description |
|----------|-------------|
| `InitTracer(ctx, service, env, endpoint)` | Initialize the global `TracerProvider` |
| `StartSpan(ctx, name, opts...)` | Start a new span |
| `RecordError(span, err)` | Record an error on a span and set status to Error |
| `InjectContext(ctx, carrier)` | Inject trace context into an outgoing carrier |
| `ExtractContext(ctx, carrier)` | Extract trace context from an incoming carrier |
| `NewHandler(handler, op, opts...)` | OTel HTTP server middleware |
| `NewTransport(base)` | OTel HTTP client transport |
| `ShutdownTracerProvider(ctx)` | Flush and shut down the provider |

### `pkg/metrics`

| Function | Description |
|----------|-------------|
| `InitMeter(ctx, service, env, endpoint)` | Initialize the global `MeterProvider` with Go runtime metrics |
| `ShutdownMeterProvider(ctx)` | Flush and shut down the provider |

### `pkg/logging`

| Function | Description |
|----------|-------------|
| `InitLogger(ctx, service, env, endpoint)` | Initialize the global `slog`-backed logger |
| `LogDebug(ctx, msg, args...)` | Log at DEBUG level |
| `LogInfo(ctx, msg, args...)` | Log at INFO level |
| `LogWarn(ctx, msg, args...)` | Log at WARN level |
| `LogError(ctx, msg, args...)` | Log at ERROR level |
| `ShutdownLogger(ctx)` | Flush and shut down the provider |

---

## Architecture

```
Your Service
    │
    ├─ pkg/tracing  ──► TracerProvider  ──► OTLP gRPC ─┐
    ├─ pkg/metrics  ──► MeterProvider   ──► OTLP gRPC ─┼─► OTel Collector ──► Backend
    └─ pkg/logging  ──► LoggerProvider  ──► OTLP gRPC ─┘
```

The library sets global OTel providers, so all instrumented libraries (e.g. `otelhttp`, `otelgrpc`) pick them up automatically without any additional wiring.

---

## Upgrade Guide

1. Update the version in your `go.mod`:

```sh
go get github.com/opensrehub/otelgokit@vX.Y.Z
```

2. Tidy dependencies:

```sh
go mod tidy
```

3. Review the [releases](https://github.com/opensrehub/otelgokit/releases) for breaking changes.

4. Test your application and commit:

```sh
git add go.mod go.sum
git commit -m "chore: upgrade otelgokit to vX.Y.Z"
```

---

## Contributing

Contributions are welcome! Please open an issue before submitting a pull request for significant changes.

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Start the example to verify end-to-end: `docker compose up --build`
4. Open a pull request

---

## License

[Apache 2.0](LICENSE)
