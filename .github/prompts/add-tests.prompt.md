---
description: "Add the four required unit tests to a provider package that is missing test coverage"
mode: agent
tools: ["codebase", "editFiles", "runCommands"]
---

You are adding unit tests to an existing provider package in `github.com/versegeek/toolkit`.

## Your task

The user will specify a package path (e.g. `cache/redis`, `mq/nats`). Read the provider's `.go` files to understand its `Config`, `Client`, and constructor signatures, then generate `<name>_test.go`.

## Required test cases (all four are mandatory)

### 1. `TestNewConfigFromEnv_Defaults`
Verify that `NewConfigFromEnv()` returns a non-nil `*Config` with sensible default values.
Assert each field that has a `v.SetDefault(...)` in `NewConfigFromEnv()`.

```go
func TestNewConfigFromEnv_Defaults(t *testing.T) {
    cfg := <name>.NewConfigFromEnv()
    assert.Equal(t, "expected-default", cfg.SomeField)
    // … one assertion per default
}
```

### 2. `TestNew_NilConfig`
Verify that `New(nil)` does not panic and returns a `*Client` with a non-nil `Config`.

```go
func TestNew_NilConfig(t *testing.T) {
    c := <name>.New(nil)
    require.NotNil(t, c)
    assert.NotNil(t, c.Config)
}
```

### 3. `TestClient_Close_WithoutInit`
Verify that `Close()` before `Init()` returns `nil` and does not panic.

```go
func TestClient_Close_WithoutInit(t *testing.T) {
    c := <name>.New(nil)
    assert.NoError(t, c.Close())
}
```

### 4. `TestNewClient_WireCompatible`
Verify the Wire constructor: must return `(non-nil client, non-nil cleanup, nil error)`.
Call `cleanup()` and verify it does not panic.

```go
func TestNewClient_WireCompatible(t *testing.T) {
    cfg := &<name>.Config{ /* minimal valid config */ }
    c, cleanup, err := <name>.NewClient(cfg)
    require.NoError(t, err)
    require.NotNil(t, c)
    defer cleanup()
    // optionally assert internal state
}
```

## Optional: integration test

If the provider can be tested without a real external service (e.g. SQLite in-memory, stdout exporter, mock server on localhost), add:

```go
func TestClient_Init_<Scenario>(t *testing.T) {
    c := <name>.New(&<name>.Config{ /* in-memory / local config */ })
    require.NoError(t, c.Init())
    defer func() { require.NoError(t, c.Close()) }()
    // exercise the client
}
```

## Rules

- Use `github.com/stretchr/testify/assert` and `require`; no Ginkgo
- Package declaration: `package <name>_test` (external test package)
- All tests must pass with the race detector: `go test -race ./<package>/...`
- Tests must not require real external services to pass in CI

## Validation

After generating the file, run:

```bash
go test -race -v ./<package>/...
```

Fix any failures before finishing.
