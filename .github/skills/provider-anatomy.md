# Provider Anatomy

Every provider in this repository follows the exact same structural shape. Deviating from this shape is a bug.

## Standard Provider Shape

```go
// Package <name> provides a <Description> Provider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(<name>.New(nil))
//	c.MustRun()
//
// Wire usage (Pattern B) – add <name>.ProviderSet to wire.Build().
package <name>

import (
    "fmt"
    "github.com/spf13/viper"
    "github.com/versegeek/toolkit/internal/config"
    "github.com/versegeek/toolkit/provider"
)

// Config holds all connection/behaviour parameters.
// Every field MUST be exported.
type Config struct {
    Address string
    Timeout time.Duration
    // ...
}

// NewConfigFromEnv reads configuration from env vars or config.yaml.
// MUST create its own viper.New() — never share a Viper instance.
func NewConfigFromEnv() *Config {
    v := viper.New()
    config.LoadFromFile(v)            // searches ., ./config, ./configs for config.yaml
    v.SetDefault("<NAME>_ADDRESS", "localhost:1234")
    return &Config{Address: v.GetString("<NAME>_ADDRESS")}
}

// Client is the lifecycle-managed wrapper. Name is always "Client".
type Client struct {
    provider.AbstractProvider   // no-op Init/Close defaults
    Config *Config
    client *underlying.Client  // unexported
}

// Compile-time interface check — REQUIRED, never omit.
var _ provider.Provider = (*Client)(nil)

// New creates the struct. MUST NOT open any connections.
func New(cfg *Config) *Client {
    if cfg == nil { cfg = NewConfigFromEnv() }
    return &Client{Config: cfg}
}

// NewClient is the Wire-compatible constructor.
func NewClient(cfg *Config) (*Client, func(), error) {
    return provider.InitProvider(New(cfg))
}

// Client exposes the underlying client. Only valid after Init().
func (c *Client) Client() *underlying.Client { return c.client }

// Init opens the connection. Wrap all errors with fmt.Errorf("<name>: ...: %w", err).
// If Init fails partway, clean up any partially-opened resources before returning.
func (c *Client) Init() error {
    conn, err := underlying.Connect(c.Config.Address)
    if err != nil {
        return fmt.Errorf("<name>: connect failed: %w", err)
    }
    c.client = conn
    return nil
}

// Close releases the connection.
// MUST have nil guard — Close() can be called even if Init() was never called.
func (c *Client) Close() error {
    if c.client == nil {
        return nil
    }
    return c.client.Close()
}
```

## RunProvider Shape (servers with a blocking loop)

Servers embed `provider.AbstractRunProvider` instead of `AbstractProvider` and implement `Run()`.

```go
type Server struct {
    provider.AbstractRunProvider   // not AbstractProvider
    Config *Config
    Router *gin.Engine
    srv    *nethttp.Server
}

var _ provider.RunProvider = (*Server)(nil)

// Init prepares the server but does NOT bind any ports.
func (s *Server) Init() error {
    // Create engine, assign s.srv — no Listen() here
    return nil
}

// Run binds the port and blocks. Call SetRunning(true) AFTER the port is bound.
func (s *Server) Run() error {
    ln, err := net.Listen("tcp", s.srv.Addr)
    if err != nil { return err }
    s.SetRunning(true)              // only after Listen succeeds
    return s.srv.Serve(ln)
}

// Close MUST guard s.srv nil — Close() can be called before Init().
func (s *Server) Close() error {
    if s.srv == nil {
        return s.AbstractRunProvider.Close()
    }
    // graceful shutdown ...
    return s.AbstractRunProvider.Close()
}
```

## wire.go Template

```go
//go:build wireinject

package <name>

import (
    "github.com/google/wire"
    "github.com/versegeek/toolkit/provider"
)

var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

var _ = provider.InitProvider[*Client]
```

## Naming Rules

| Item | Rule | Example |
|------|------|---------|
| Struct | Always `Client` | `type Client struct` |
| No-Wire constructor | `New(cfg)` | `func New(cfg *Config) *Client` |
| Wire constructor | `NewClient(cfg)` | `func NewClient(cfg *Config) (*Client, func(), error)` |
| Wire set | `ProviderSet` | `var ProviderSet = wire.NewSet(...)` |
| Cluster variant | `ClusterProviderSet` | only when backend has native cluster mode |
| Env var prefix | `UPPER_SNAKE_CASE` service name | `REDIS_ADDRESS`, `NATS_URL` |

## Directory Placement

| Provider type | Directory |
|--------------|-----------|
| Relational / document / analytical DB | `db/<name>/` |
| Cache / KV store | `cache/<name>/` |
| Message queue / broker | `mq/<name>/` |
| HTTP/gRPC/debug server | `server/<name>/` |
| Logging backend | `log/<name>/` |
| Distributed tracing | `trace/<name>/` |
| Auth / OAuth | `auth/<name>/` |

## Known Exceptions

- `db/ent`: exposes `Driver() *entsql.Driver` and `Dialect() string` instead of `Client()` because the ent ORM wraps a driver, not a client. Usage: `ent.NewClient(ent.Driver(client.Driver()))`.
- `cache/redis`: has two structs — `Client` (single-node) and `ClusterClient` (cluster), each with their own `ProviderSet` / `ClusterProviderSet`.
