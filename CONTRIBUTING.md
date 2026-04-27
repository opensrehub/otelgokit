# Contributing to otelgokit

Thank you for your interest in contributing!

## Getting started

1. Fork the repository and clone it locally.
2. Install Go (see `go.mod` for the minimum version).
3. Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) to run the full example stack.

## Making changes

1. Create a feature branch:
   ```sh
   git checkout -b feat/my-feature
   ```

2. Make your changes, then run the full check suite:
   ```sh
   go build ./...
   go vet ./...
   go test ./...
   ```

3. Verify end-to-end with the example stack:
   ```sh
   docker compose up --build
   ```

4. Open a pull request against `main`.

## Pull request guidelines

- Keep PRs focused — one concern per PR.
- Add or update tests for any changed behaviour.
- Update the README if the public API changes.
- All CI checks must pass before merging.

## Reporting issues

Open a GitHub issue with:
- What you expected to happen
- What actually happened
- Go version (`go version`) and OS

## License

By contributing, you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).
