# toolkit

[![Go Reference](https://pkg.go.dev/badge/github.com/versegeek/toolkit.svg)](https://pkg.go.dev/github.com/versegeek/toolkit)
[![CI](https://github.com/versegeek/toolkit/actions/workflows/ci.yaml/badge.svg)](https://github.com/versegeek/toolkit/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/versegeek/toolkit)](https://goreportcard.com/report/github.com/versegeek/toolkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A **lifecycle-managed provider library** for Go. Every integration (database, cache, message queue, tracing …) is an independent, zero-dependency provider that you wire together on demand — no monolithic init, no hidden globals.

## Features

- **On-demand loading** — only the providers you register are initialised
- **LIFO shutdown** — `MustClose()` tears down providers in reverse-init order automatically
- **Two composition patterns** — lightweight Container (no codegen) or Google Wire (compile-time DI)
- **Race-free** — all providers are tested with `-race`
- **Pure-Go SQLite** — no CGo required for SQLite drivers

## Provider Catalog

| Category | Package | Description |
|----------|---------|-------------|
| **Database** | `db/gorm` | GORM ORM — SQLite / MySQL / PostgreSQL / ClickHouse |
| | `db/ent` | ent ORM — SQLite / MySQL / PostgreSQL |
| | `db/clickhouse` | Native ClickHouse client (`clickhouse-go/v2`) |
| | `db/mongo` | MongoDB client |
| **Cache** | `cache/redis` | Redis single-node + Cluster |
| **Message Queue** | `mq/rabbitmq` | RabbitMQ (AMQP 0-9-1) |
| | `mq/nats` | NATS core + JetStream |
| | `mq/nsq` | NSQ producer |
| **Server** | `server/http` | Gin HTTP server with graceful shutdown |
| | `server/grpc` | gRPC server with health + reflection |
| | `server/pprof` | pprof debug server |
| | `server/prometheus` | Prometheus metrics server |
| **Logging** | `log/logrus` | logrus — configures global logger |
| | `log/zap` | zap — high-performance structured logger |
| **Tracing** | `trace/otel` | OpenTelemetry — OTLP / Zipkin / stdout |
| **Auth** | `auth/oauth2` | OAuth2 Authorization Code + Client Credentials |
| **Utils** | `utils` | UUID helpers |

## Installation

```bash
go get github.com/versegeek/toolkit
```

Import only the providers you need:

```bash
go get github.com/versegeek/toolkit/cache/redis
go get github.com/versegeek/toolkit/db/gorm
```

## Quick Start

### Pattern A — Container (no code generation)

Suitable for simple services or scripts where you want explicit control.

```go
package main

import (
    "github.com/versegeek/toolkit/container"
    "github.com/versegeek/toolkit/cache/redis"
    "github.com/versegeek/toolkit/db/gorm"
    "github.com/versegeek/toolkit/server/http"
)

func main() {
    c := container.New()

    // Providers are initialised in registration order.
    c.MustInit(redis.New(nil))   // nil → reads from env / config.yaml
    c.MustInit(gorm.New(nil))

    srv := http.New(nil)
    c.MustInit(srv)

    // Register routes after Init, before Run.
    srv.Router.GET("/ping", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"message": "pong"})
    })

    // MustRun starts RunProviders and blocks until SIGINT/SIGTERM.
    // MustClose is called automatically in LIFO order on shutdown.
    defer c.MustClose()
    c.MustRun()
}
```

### Pattern B — Google Wire (compile-time DI)

Suitable for larger services where you want compile-time verification.

```go
// cmd/wire.go  (build tag: wireinject)
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/versegeek/toolkit/cache/redis"
    "github.com/versegeek/toolkit/db/gorm"
)

func InitApp() (*App, func(), error) {
    wire.Build(
        redis.ProviderSet,
        gorm.ProviderSet,
        NewApp,
    )
    return nil, nil, nil
}
```

```go
// main.go
func main() {
    app, cleanup, err := InitApp()
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()
    container.HandleSignals(cleanup) // blocks until signal
}
```

Generate the wiring: `make wire`

## Configuration

Every provider reads its configuration from environment variables (or a `config.yaml` / `configs/config.yaml` file). Environment variables always take precedence.

```yaml
# config.yaml
REDIS_ADDRESS: "localhost:6379"

DB_DRIVER: "sqlite"          # sqlite | mysql | postgres | clickhouse
MYSQL_DSN: ""                # leave empty to use individual host/port/db fields

TRACE_SERVICE_NAME: "my-service"
TRACE_EXPORTER: "otlp"       # otlp | zipkin | stdout
TRACE_OTLP_ENDPOINT: "http://localhost:4318"

ZAP_LEVEL: "info"
ZAP_ENCODING: "json"
```

See each provider package's `NewConfigFromEnv()` function for the full list of supported keys.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to add new providers, coding conventions, and the PR checklist.
