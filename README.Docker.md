### Building and running your application

When you're ready, start your application by running:
`docker compose up --build`.

Your application will be available at http://localhost:3000.

#### Accessing OTEL TUI

The OTEL TUI runs as a service in Docker Compose and provides an interactive terminal interface for observing telemetry data.

**To access the OTEL TUI:**

1. **Start the services:**
   ```bash
   docker compose up -d
   ```

2. **Attach to the OTEL TUI container:**
   ```bash
   docker compose attach otel-tui
   ```

#### What OTEL TUI Shows

The OTEL TUI provides real-time monitoring of:

- **📊 Metrics**: OpenTelemetry metrics from the application (request counter, latency histogram, Go runtime)
- **🔍 Traces**: Distributed traces showing request flow through the system
- **📝 Logs**: Structured logs with OpenTelemetry correlation
- **⚡ Spans**: Individual operations and their timing/metadata


### Deploying your application to the cloud

First, build your image, e.g.: `docker build -t myapp .`.
If your cloud uses a different CPU architecture than your development
machine (e.g., you are on a Mac M1 and your cloud provider is amd64),
you'll want to build the image for that platform, e.g.:
`docker build --platform=linux/amd64 -t myapp .`.

Then, push it to your registry, e.g. `docker push myregistry.com/myapp`.

Consult Docker's [getting started](https://docs.docker.com/go/get-started-sharing/)
docs for more detail on building and pushing.

### References
* [Docker's Go guide](https://docs.docker.com/language/golang/)