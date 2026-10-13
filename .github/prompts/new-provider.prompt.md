---
description: "Scaffold a complete, lifecycle-managed provider package for github.com/versegeek/toolkit"
mode: agent
tools: ["codebase", "editFiles", "runCommands"]
---

You are scaffolding a new provider for the `github.com/versegeek/toolkit` library.

## Your task

Create a fully working provider package. Ask the user for:
1. **What service/SDK** does this provider wrap? (e.g. Apache Kafka, Elasticsearch)
2. **Which category directory** should it live in?
   - `db/` — relational, document, or analytical databases
   - `cache/` — cache / KV stores
   - `mq/` — message queues and brokers
   - `server/` — servers with a blocking `Run()` loop
   - `log/` — logging backends
   - `trace/` — distributed tracing
   - `auth/` — authentication / authorisation

Then generate three files in `<category>/<name>/`:

### `<name>.go`

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

// Config holds <Name> configuration.
type Config struct {
    Address string
    // … add all fields, all exported
}

// NewConfigFromEnv reads <Name> configuration from the environment.
// Keys: <NAME>_ADDRESS, …
func NewConfigFromEnv() *Config {
    v := viper.New()
    config.LoadFromFile(v)

    v.SetDefault("<NAME>_ADDRESS", "localhost:XXXX")

    return &Config{
        Address: v.GetString("<NAME>_ADDRESS"),
    }
}

// Client wraps *underlying.Client as a lifecycle-managed Provider.
type Client struct {
    provider.AbstractProvider   // embed for no-op Init/Close defaults

    Config *Config
    client *underlying.Client  // unexported
}

// compile-time interface check — required
var _ provider.Provider = (*Client)(nil)

// New creates a Client without connecting.
// If config is nil, NewConfigFromEnv() is used.
func New(cfg *Config) *Client {
    if cfg == nil {
        cfg = NewConfigFromEnv()
    }
    return &Client{Config: cfg}
}

// NewClient is the Wire-compatible constructor.
func NewClient(cfg *Config) (*Client, func(), error) {
    return provider.InitProvider(New(cfg))
}

// Client returns the underlying client. Only valid after Init().
func (c *Client) Client() *underlying.Client { return c.client }

// Init opens the connection.
func (c *Client) Init() error {
    conn, err := underlying.Connect(c.Config.Address)
    if err != nil {
        return fmt.Errorf("<name>: connect failed: %w", err)
    }
    c.client = conn
    return nil
}

// Close releases the connection. Safe to call if Init() was never called.
func (c *Client) Close() error {
    if c.client == nil {
        return nil
    }
    return c.client.Close()
}
```

### `wire.go`

```go
//go:build wireinject

package <name>

import (
    "github.com/google/wire"
    "github.com/versegeek/toolkit/provider"
)

// ProviderSet is the Wire provider set for the <Name> provider.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

var _ = provider.InitProvider[*Client]
```

### `<name>_test.go`

Implement all four required test cases:

1. `TestNewConfigFromEnv_Defaults` — verify default config values are sensible
2. `TestNew_NilConfig` — `New(nil)` must not panic, must return non-nil `Config`
3. `TestClient_Close_WithoutInit` — `Close()` before `Init()` must return `nil`, must not panic
4. `TestNewClient_WireCompatible` — `NewClient(cfg)` returns `(client, cleanup, nil)`; `cleanup()` must not panic

If the provider can be tested without a real external service (e.g. it connects to `localhost` with an in-process server, or uses an in-memory mode), also add:

5. `TestClient_Init_Success` — full round-trip: `Init()`, exercise the client, `Close()`

## Validation

After generating the files run:

```bash
go build ./...
go test -race ./<category>/<name>/...
go vet ./<category>/<name>/...
```

Fix any errors before finishing.
