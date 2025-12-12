# Chaos CLI

Chaos CLI is a developer tool for running HTTP chaos experiments, discovering API endpoints, and analyzing the impact of injected failures and delays.

It provides:
- A reverse proxy that can inject delay and failure rates on specific routes.
- Passive endpoint discovery by observing live traffic through a proxy.
- An analyzer that compares baseline vs experiment metrics to highlight direct and hidden impacts.

## Installation

Prerequisites:
- `Go` 1.20+ installed and available in `PATH`.

Clone and use directly without installation:
- Run commands via `go run .` to avoid building example binaries.

Build the CLI binary (platform-specific):
- Windows (PowerShell):
  - `go build -o chaos-cli.exe .`
  - Run local executable with `.\/chaos-cli.exe ...` (PowerShell requires the `.` prefix for local files)
- macOS/Linux:
  - `go build -o chaos-cli .`
  - Run with `./chaos-cli ...`

Note: Avoid `go build ./...` as it compiles all packages, including `examples` with multiple `main` functions, which will fail.

## Quick Start

1) Start a demo backend (port `3000`):
- `go run examples/demo-server.go`

2) Run a chaos proxy for 5 seconds, injecting 25ms delay and 10% failures on `POST /login`:
- `go run . http proxy --target http://localhost:3000 --port 8080 --delay 25ms --failure-rate 0.1 --path /login --method POST --duration 5s --output experiment.ndjson`

3) Exercise the proxy during the run (in a separate terminal):
- `curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"user":"demo"}'`
- `curl http://localhost:8080/products`

4) Capture a baseline (no chaos):
- `go run . http proxy --target http://localhost:3000 --port 8080 --duration 5s --output baseline.ndjson`

5) Analyze impact (prints text and saves JSON):
- `go run . http analyze --baseline baseline.ndjson --experiment experiment.ndjson --format both --output impact.report.json`

## Commands

### HTTP Proxy
Start a reverse proxy and apply chaos rules.

Usage:
- `go run . http proxy [flags]`

Flags:
- `--target` string: Backend target URL (default `http://localhost:3000`).
- `--port` int: Proxy listen port (default `8080`).
- `--delay` duration: Delay to inject (e.g., `25ms`, `1s`).
- `--failure-rate` float: Failure rate `0.0`–`1.0` (e.g., `0.1`).
- `--path` string: API path to match (e.g., `/login`).
- `--method` string: HTTP method to match (`GET`, `POST`, etc.). Empty means any.
- `--duration` duration: Runtime (e.g., `60s`). `0` means run until Ctrl+C.
- `--output` string: File to write NDJSON metrics on shutdown.

Example:
- `go run . http proxy --target http://localhost:3000 --port 8080 --delay 100ms --failure-rate 0.2 --path /orders --method GET --duration 10s --output experiment.ndjson`

### Discover
Discover API endpoints by observing traffic through a reverse proxy.

Usage:
- `go run . discover [flags]`

Flags:
- `--target` string: Backend target URL (required).
- `--port` string: Proxy listen port (default `8080`).
- `--duration` int: Auto-stop after N seconds (`0` = manual via Ctrl+C).
- `--output` string: Output file for discovered endpoints (default `endpoints.json`).

Example:
- `go run . discover --target http://localhost:3000 --port 8081 --duration 6 --output endpoints.json`

Send traffic through the discovery proxy (port `8081`) while it runs to capture endpoints.

### Analyze
Compare baseline vs experiment metrics and generate an impact report.

Usage:
- `go run . http analyze --baseline baseline.ndjson --experiment experiment.ndjson --format [text|json|both] --output impact.report.json`

Output:
- Text report printed to console.
- JSON report saved when `--format json` or `--format both`.

## Demo Server

The `examples/demo-server.go` provides a simple backend with endpoints like `/login`, `/orders`, `/products`.
- Start with: `go run examples/demo-server.go`
- Use the proxy to route traffic through `http://localhost:8080` or discovery via `http://localhost:8081`.

## Metrics Format (NDJSON)

Each line in metrics files is a JSON object with:
- `timestamp`: time of request.
- `method`: HTTP method.
- `path`: request path.
- `status_code`: response status.
- `latency_ms`: measured latency in ms.
- `chaos_applied`: whether chaos was applied.
- `chaos_type`: `delay`, `failure`, or `none`.
- `backend_error`: whether the backend returned an error or proxy detected it.

## Troubleshooting

- Build errors when running `go build ./...`:
  - Use `go run .` or `go build -o chaos-cli .` to avoid compiling `examples` with multiple `main` functions.
- No metrics written:
  - Ensure `--output` is set and traffic is sent through the proxy during the run.
- Discovery reports 0 endpoints:
  - Send requests through the discovery proxy (not directly to backend) while it runs.
- Windows tips:
  - Use PowerShell for commands; curl is available as `curl`.

## Development

- Code is organized under `pkg/` with packages for `proxy`, `metrics`, `discover`, and `analyze`.
- Main CLI commands live under `cmd/`.
- Recommended workflow:
  - Run demo server.
  - Use `http proxy` to generate baseline and experiment NDJSON.
  - Run `analyze` to compare and inspect the impact.
  - Use `discover` to capture endpoint inventory.

## License

This project is licensed under the terms specified in `LICENSE`.