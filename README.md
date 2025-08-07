# Agent and Server for collecting runtime & system metrics via REST/gRPC

## Agent

![agent](./images/agent.png)

The Agent collects runtime/system metrics using Go’s `runtime` and system stats, then sends them to a server using either **REST** or **gRPC** at configurable intervals.

### Features
- runtime + system metrics
- REST/gRPC support
- worker pool for limited concurrent reporting
- HMAC Hash signing
- gzip compression
- configurable polling/report intervals
- swagger docs

### Usage

```bash
go run ./cmd/agent --config=configs/agent.yaml
```

```bash
go run ./cmd/agent -a="localhost:8080" -p=2 -r=10 \
-k="aliffka" -p="rest" -l=5
```

### Flags
| Flag                    | Description                       |
|-------------------------|-----------------------------------|
| `--address`, `-a`       | Server endpoint address           |
| `--pollinterval`, `-p`  | Interval to collect metrics (sec) |
| `--reportinterval`, `-r`| Interval to send metrics (sec)    |
| `--key`, `-k`           | Secret key for HMAC signing       |
| `--limit`, `-l`         | Max concurrent requests           |
| `--protocol`, `-t`      | `rest` or `grpc`                  |
| `--config`, `-c`        | Path to config file               |


## Server

![server](./images/server.png)

HTTP/gRPC server for collecting metrics(gauge and counter). It supports in-memory storage, file persistence, and PostgreSQL for durability. The server is built with modular architecture(repository-usecase-service) and supports configuration via flags or YAML file.

### Features
- REST/gRPC support
- reposiroty: memory, file, PostgreSQL
- HMAC signature validation
- IP subnet validation
- configurable periodic backup
- swagger docs

### Usage

```bash
go run ./cmd/server/server --config=configs/server.yaml
```

```bash
go run ./cmd/server/server -a="localhost:8080" -i=300 -r=true \
-f=/tmp/metrics.json -d=postgres://user:pass@localhost:5432/metrics \
-k="aliffka" -p="grpc"
```

### Flags

| Flag               | Description                                        |
|--------------------|----------------------------------------------------|
| `--a`              | HTTP/gRPC server address                           |
| `--i`              | Interval for storing data (in seconds)             |
| `--r`              | Restore data from file on startup (`true/false`)   |
| `--f`              | File path for metrics persistence                  |
| `--d`              | PostgreSQL connection string (DSN)                 |
| `--key`, `-k`      | Secret key for verifying HMAC signatures           |
| `--t`              | Trusted subnet for gRPC client IP validation       |
| `--protocol`, `-p` | Protocol to use: `rest` or `grpc`                  |
| `--config`, `-c`   | Path to configuration file                         |

## Build

This project uses a `Makefile` to manage builds, dependencies, testing, and more.

| Target             | Description                                                     |
|--------------------|-----------------------------------------------------------------|
|`make all`          | Run everything: deps, server build, agent build, tests, and lint|
|`make server`       | Build the server binary                                         |
|`make agent`        | Build the agent binary                                          |
|`make deps`         | Download Go module dependencies and vendor them                 |
|`make test`         | Run all unit tests with coverage report                         |
|`make lint`         | Run `golangci-lint` to check for code issues                    |
|`make clean`        | Remove compiled binaries and coverage files                     |
|`make swagger`      | Generate Swagger docs using `swag`                              |
|`make proto`        | Compile `.proto` definitions into Go files (`protoc` required)  |
|`make test_results` | View HTML coverage report and delete temp files after           |
