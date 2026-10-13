# Lifecycle Contract

## Four-Phase Lifecycle

Every provider passes through four phases in order:

```
New()   →   Init()   →   Run()   →   Close()
 ↑            ↑            ↑           ↑
struct     connect      goroutine   disconnect
created    opens        starts      releases
```

| Phase | Who calls it | What it must do |
|-------|-------------|-----------------|
| `New(cfg)` | User code | Create struct, store config. NO connections. |
| `Init()` | `Container.MustInit` or `provider.InitProvider` | Open connections. Clean up partial state on error. |
| `Run()` | `Container.MustRun` (in a goroutine) | Block until stopped. RunProvider only. |
| `Close()` | `Container.MustClose` (LIFO order) | Release all resources. Must be safe to call even if Init() never ran. |

## LIFO Shutdown Order

Providers are closed in **reverse** registration order:

```go
c.MustInit(log)    // registered 1st → closed LAST
c.MustInit(db)     // registered 2nd → closed 2nd-to-last
c.MustInit(redis)  // registered 3rd → closed 1st
```

This ensures downstream providers close before upstream ones (e.g. HTTP server closes before DB).

## Signal Handling

`Container.MustRun()` blocks until `SIGINT`, `SIGTERM`, or `SIGQUIT`, then calls `MustClose()` automatically.

```go
c := container.New()
c.MustInit(redis.New(nil))
defer c.MustClose()   // safety net if MustRun is not used
c.MustRun()           // blocks; MustClose called on signal
```

For Pattern B (Wire), use `container.HandleSignals(cleanup)`:

```go
app, cleanup, err := InitApp()
defer cleanup()
container.HandleSignals(cleanup)  // blocks until signal, then calls cleanup
```

## provider.InitProvider

`InitProvider[T]` is the helper that makes a provider Wire-compatible:

```go
func InitProvider[T Provider](p T) (T, func(), error) {
    if err := p.Init(); err != nil {
        return p, nil, fmt.Errorf("%T init failed: %w", p, err)
    }
    return p, func() { p.Close() }, nil
}
```

It calls `Init()` and returns a cleanup function. It does **not** start goroutines — that is exclusively `Container.MustRun`'s responsibility.

## RunProvider Steady State

`AbstractRunProvider` tracks whether a server is ready via an `atomic.Bool`:

```go
// Call this only AFTER the port is bound / the service is accepting requests.
s.SetRunning(true)

// Check from another goroutine (e.g. a health probe):
if !server.IsRunning() { /* not ready yet */ }
```

**Wrong pattern** (calling SetRunning before the port is bound):
```go
func (s *Server) Run() error {
    s.SetRunning(true)           // ❌ port not bound yet
    return s.srv.ListenAndServe()
}
```

**Correct pattern** (same as gRPC server):
```go
func (s *Server) Run() error {
    ln, err := net.Listen("tcp", addr)
    if err != nil { return err }
    s.SetRunning(true)           // ✅ port is bound
    return s.srv.Serve(ln)
}
```

## Error Handling Rules

- Wrap errors: `fmt.Errorf("<package>: <action> failed: %w", err)`
- On partial Init failure: close any already-opened resources before returning
- `Close()` must never panic — use nil guard at the top

```go
func (c *Client) Close() error {
    if c.client == nil {   // ← always guard
        return nil
    }
    return c.client.Close()
}
```

## Idempotency

- `MustRun` and `MustClose` are idempotent per Container instance (backed by `sync.Once`)
- Multiple `Container` instances can coexist in the same process safely
- `Init()` on individual providers is **not** idempotent — calling twice may leak resources
