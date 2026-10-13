# toolkit

[![Go Reference](https://pkg.go.dev/badge/github.com/versegeek/toolkit.svg)](https://pkg.go.dev/github.com/versegeek/toolkit)
[![CI](https://github.com/versegeek/toolkit/actions/workflows/ci.yaml/badge.svg)](https://github.com/versegeek/toolkit/actions/workflows/ci.yaml)

Lifecycle-managed providers for Go services (databases, caches, message queues, servers, logging, tracing and auth) that you compose by hand or with [Google Wire](https://github.com/google/wire).

> :construction: Under development: APIs may change until v1.

## Features

- **One shape for every provider**: `New` does no I/O, `Init` connects, run providers `Run` in the foreground, `Close` releases. A failed `Init` leaves nothing open.
- **Ordered lifecycle**: run providers start only after everything is initialised; SIGINT, SIGTERM or a failing run provider shuts everything down once, in reverse order.
- **With or without Wire**: every provider package exports a `ProviderSet`; `lifecycle.Group` covers the hand-written case.
- **Configuration that fails loudly**: environment variables or `config.yaml`; a malformed file is an error, not silently ignored.
- **One log stream**: the toolkit logs through `log/slog`, and the zap and logrus providers route it into their backend.
- **Pure-Go SQLite**: no cgo.

## Providers

| Package | Provides | Config keys |
|---------|----------|-------------|
| `cache/redis` | go-redis client and cluster client | `REDIS_*` |
| `db/gorm` | GORM for MySQL, PostgreSQL, SQLite and ClickHouse | `DB_*` |
| `db/ent` | ent SQL driver for MySQL, PostgreSQL and SQLite | `DB_*` |
| `db/clickhouse` | Native ClickHouse connection | `CLICKHOUSE_*` |
| `db/mongo` | MongoDB client (driver v2) | `MONGODB_*` |
| `mq/rabbitmq` | RabbitMQ (AMQP 0-9-1) connection | `RABBITMQ_*` |
| `mq/nats` | NATS connection, optional JetStream | `NATS_*` |
| `mq/nsq` | NSQ producer | `NSQ_*` |
| `server/http` | Gin HTTP server | `HTTP_*` |
| `server/grpc` | gRPC server with health and reflection | `GRPC_*` |
| `server/pprof` | pprof endpoints, on 127.0.0.1 by default | `PPROF_*` |
| `server/prometheus` | Prometheus metrics endpoint | `PROMETHEUS_*` |
| `log/zap` | zap logger, installed as `zap.L` and the slog default | `ZAP_*` |
| `log/logrus` | logrus standard logger, installed as the slog default | `LOGRUS_*` |
| `trace/otel` | OpenTelemetry tracing over OTLP/HTTP or stdout | `TRACE_*` |
| `auth/oauth2` | OAuth2 authorization code and client credentials flows | `OAUTH2_*` |

Each package documents its keys and defaults on its `NewConfigFromEnv`. The supporting packages are `provider` (the interfaces), `lifecycle` (start-up and shutdown), `config` (Viper loading) and `utils`.

## Installation

```bash
go get github.com/versegeek/toolkit@latest
```

Requires Go 1.25 or later. Your binary only contains the providers you import.

## Usage

### Manual composition

```go
package main

import (
	"github.com/gin-gonic/gin"

	"github.com/versegeek/toolkit/cache/redis"
	"github.com/versegeek/toolkit/lifecycle"
	zaplog "github.com/versegeek/toolkit/log/zap"
	"github.com/versegeek/toolkit/server/http"
)

func main() {
	g := lifecycle.New()
	defer g.MustClose() // releases what was initialised if a later MustInit panics

	g.MustInit(zaplog.New(nil)) // nil config: read from the environment in Init
	rdb := redis.New(nil)
	g.MustInit(rdb)
	srv := http.New(nil)
	g.MustInit(srv)

	// Register routes after Init and before Run.
	srv.Router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"pings": rdb.Client().Incr(c, "pings").Val()})
	})

	g.MustRun() // serves until SIGINT/SIGTERM, then closes http, redis, zap
}
```

Full example: [`examples/group`](examples/group).

### Wire composition

Wire builds and initialises the providers; `lifecycle.Run` runs the servers and calls Wire's cleanup exactly once.

```go
//go:build wireinject

package main

func InitApp() (*App, func(), error) {
	panic(wire.Build(zaplog.ProviderSet, redis.ProviderSet, http.ProviderSet, NewApp))
}
```

```go
// NewApp receives initialised providers; register routes here.
func NewApp(logger *zaplog.Client, rdb *redis.Client, srv *http.Server) *App {
	srv.Router.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
	return &App{Log: logger, Redis: rdb, HTTP: srv}
}

func main() {
	app, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	if err := lifecycle.Run(context.Background(), cleanup, app.HTTP); err != nil {
		log.Fatal(err)
	}
}
```

Generate `wire_gen.go` with `make wire`. Full example: [`examples/wire`](examples/wire).

## Configuration

- Each provider reads environment variables and an optional `config.yaml` from `.`, `./config` or `./configs`; environment variables win.
- Pass a `Config` to skip the environment: `redis.New(&redis.Config{Addrs: []string{"localhost:6379"}})`.
- With `New(nil)` the configuration is loaded in `Init`, so configuration errors surface there.
- List values such as `REDIS_ADDRESS` (for a cluster) or `ZAP_OUTPUT` are comma-separated.
- `db/gorm` and `db/ent` share the `DB_*` keys and require `DB_DRIVER` (`mysql`, `postgres`, `sqlite` or `clickhouse`).

```yaml
# config.yaml
REDIS_ADDRESS: "localhost:6379"
DB_DRIVER: "postgres"
DB_HOST: "localhost"
DB_DATABASE: "app"
HTTP_PORT: 8080
TRACE_SERVICE_NAME: "orders"
ZAP_LEVEL: "info"
```

## Writing a provider

A provider implements `provider.Provider`, and a run provider also implements `provider.RunProvider`:

- `New(cfg)` stores the config and does no I/O.
- `Init` loads the config when it is nil, connects, and closes whatever it opened before returning an error.
- `Close` is safe before or after a failed `Init`. A run provider's `Close` makes `Run` return and may be called twice.
- Implement `provider.Enabler` if configuration can switch the provider off.
- Export `NewClient` (or `NewServer`) as `provider.InitProvider(New(cfg))` and a `ProviderSet`.

Terms are defined in [GLOSSARY.md](GLOSSARY.md) and the design decisions in [docs/adr](docs/adr).

## Migrating from the `pkg/` layout

| Before (`36c39b3`) | Now |
|--------------------|-----|
| `pkg/stack.New()` | `lifecycle.New()`; `MustInit`, `MustRun` and `MustClose` are unchanged |
| `pkg/provider` | `provider`; `Enabled()` is now the optional `provider.Enabler` interface |
| `pkg/common.LoadFromFile(v)` | `config.LoadFromFile(v)`, which also reads environment variables and returns an error |
| `pkg/provider/redis` | `cache/redis`; `Config.Address` became `Config.Addrs` |
| `pkg/provider/gorm` (MySQL) | `db/gorm` with `DB_DRIVER=mysql`; `MYSQL_*` keys became `DB_*` |

## Development

```bash
make build        # go build ./...
make test         # go test -race ./...
make lint         # go fmt, go vet, golangci-lint
make wire         # regenerate wire_gen.go in examples/wire
make wire-check   # fail if wire_gen.go is stale
make tidy         # go mod tidy
```
