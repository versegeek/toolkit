# Testing Guide

## Required Test Cases

Every provider package MUST have a `<name>_test.go` with at least these four tests:

### 1. `TestNewConfigFromEnv_Defaults`
Verify every `SetDefault` call produces the expected value.
```go
func TestNewConfigFromEnv_Defaults(t *testing.T) {
    cfg := mypkg.NewConfigFromEnv()
    assert.Equal(t, "localhost:1234", cfg.Address)
    assert.Equal(t, 5*time.Second, cfg.Timeout)
}
```

### 2. `TestNew_NilConfig`
`New(nil)` must not panic and must return a non-nil `Config`.
```go
func TestNew_NilConfig(t *testing.T) {
    c := mypkg.New(nil)
    require.NotNil(t, c)
    assert.NotNil(t, c.Config)
}
```

### 3. `TestClient_Close_WithoutInit`
`Close()` before `Init()` must return `nil` and must not panic.
```go
func TestClient_Close_WithoutInit(t *testing.T) {
    c := mypkg.New(nil)
    assert.NoError(t, c.Close())
}
```

### 4. `TestNewClient_WireCompatible`
Wire constructor returns `(non-nil, cleanup, nil)`. Cleanup must not panic.
```go
func TestNewClient_WireCompatible(t *testing.T) {
    c, cleanup, err := mypkg.NewClient(&mypkg.Config{...})
    require.NoError(t, err)
    require.NotNil(t, c)
    defer cleanup()
}
```

## Race Detector is Mandatory

All tests must pass with `-race`. No exceptions.

```bash
go test -race ./...               # full suite
go test -race ./db/gorm/...       # single package
go test -race -run TestName ./mq/nats/  # single test
```

## No External Services in CI

Tests that require a live Redis/MySQL/NATS etc. must use `t.Skip` when the service is unavailable:

```go
func TestClient_Init_Success(t *testing.T) {
    c, cleanup, err := mypkg.NewClient(cfg)
    if err != nil {
        t.Skipf("service not reachable: %v", err)
    }
    defer cleanup()
    // exercise the client
}
```

For databases, prefer in-memory alternatives:
- SQLite: `DSN: ":memory:"` — no external service needed
- ClickHouse: test with `Init()` error path only
- MongoDB: test error paths with `127.0.0.1:<unused-port>`

## Test Package Convention

Use the external test package (`package <name>_test`), not the internal one:

```go
// ✅ correct
package gorm_test

// ❌ avoid — gives access to unexported fields, which can hide real API issues
package gorm
```

## Testing Error Paths

Always test the failure path of `Init()`, not just the happy path. This catches goroutine/resource leaks:

```go
func TestClient_Init_Unreachable(t *testing.T) {
    cfg := &mypkg.Config{Address: "127.0.0.1:19999"} // unreachable
    c := mypkg.New(cfg)
    err := c.Init()
    require.Error(t, err)
    assert.NoError(t, c.Close()) // must be safe after failed Init
}
```

## Tooling

```bash
# testify is the only allowed test helper library
import "github.com/stretchr/testify/assert"
import "github.com/stretchr/testify/require"

# No Ginkgo, no gomock, no other frameworks
```

## Packages Currently Missing Tests

The following packages have logic but no test files. New PRs for these packages must add tests:

| Package | Minimum tests needed |
|---------|---------------------|
| `cache/redis` | NewConfigFromEnv, Close-without-Init, cluster nil guard |
| `container` | MustInit, MustClose LIFO order, idempotency |
| `server/grpc` | NewConfigFromEnv, Close-without-Init |
| `server/pprof` | NewConfigFromEnv, Close-without-Init |
| `server/prometheus` | NewConfigFromEnv, Close-without-Init |
| `log/logrus` | NewConfigFromEnv level/formatter/output selection |
