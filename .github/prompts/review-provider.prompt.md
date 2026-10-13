---
description: "Review a provider package for compliance with toolkit conventions and fix all issues"
mode: agent
tools: ["codebase", "editFiles", "runCommands"]
---

You are reviewing a provider package in `github.com/versegeek/toolkit` for correctness, robustness, and convention compliance.

## What to review

The user will specify a package path (e.g. `db/gorm`, `mq/rabbitmq`). Read all `.go` files in that package.

## Checklist

Work through every item below. For each failing item, **fix it immediately** before moving on.

### Structure & Naming

- [ ] Package lives in the correct category directory (`db/`, `cache/`, `mq/`, `server/`, `log/`, `trace/`, `auth/`)
- [ ] Main struct is named `Client`
- [ ] Has `New(cfg *Config) *Client` constructor (does **not** connect)
- [ ] Has `NewClient(cfg *Config) (*Client, func(), error)` Wire-compatible constructor
- [ ] `Config` struct has all fields exported
- [ ] `NewConfigFromEnv()` creates its own `viper.New()` and calls `internal/config.LoadFromFile(v)`
- [ ] `wire.go` exists with `//go:build wireinject` tag and `ProviderSet`

### Interface & Lifecycle

- [ ] `var _ provider.Provider = (*Client)(nil)` compile-time check present
- [ ] Embeds `provider.AbstractProvider` (or `provider.AbstractRunProvider` for servers)
- [ ] `New()` does **not** open any connections
- [ ] `Init()` opens connections, wraps errors with `fmt.Errorf("<name>: ...: %w", err)`
- [ ] `Close()` has nil guard: `if c.client == nil { return nil }`
- [ ] `Close()` does **not** panic if `Init()` was never called or failed partway

### Robustness

- [ ] No goroutine leak: if `Init()` fails midway, all partially-opened resources are closed before returning the error
- [ ] No data race: any field written in one goroutine and read in another is protected (mutex or atomic)
- [ ] `Init()` is idempotent or returns an error on second call (whichever fits the design)

### Tests

- [ ] `TestNewConfigFromEnv_Defaults` exists
- [ ] `TestNew_NilConfig` exists
- [ ] `TestClient_Close_WithoutInit` exists
- [ ] `TestNewClient_WireCompatible` exists
- [ ] All tests pass: `go test -race ./<package>/...`

## Validation

After all fixes, run:

```bash
go build ./...
go test -race ./<package>/...
go vet ./<package>/...
```

Report a summary of every issue found and fixed. If no issues were found, say so explicitly.
