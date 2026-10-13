# Security & Quality Conventions

## Credential Masking

**Never log credentials in plaintext.** URLs, DSNs, or connection strings that may contain passwords must be masked before logging.

### Rule
```go
// ❌ Wrong — logs password in plaintext
logrus.WithField("url", cfg.URL).Debug("connected")

// ✅ Correct — mask before logging
logrus.WithField("url", maskURL(cfg.URL)).Debug("connected")
```

### Pattern: maskURL / maskURI
Use `strings.LastIndex` to find `@` — passwords may themselves contain `@`:

```go
func maskURL(rawURL string) string {
    schemeEnd := strings.Index(rawURL, "://")
    if schemeEnd < 0 { return rawURL }
    authority := rawURL[schemeEnd+3:]
    atIdx := strings.LastIndex(authority, "@")  // LastIndex, not Index
    if atIdx < 0 { return rawURL }
    userinfo := authority[:atIdx]
    colonIdx := strings.Index(userinfo, ":")
    if colonIdx < 0 { return rawURL }
    return rawURL[:schemeEnd+3] + userinfo[:colonIdx+1] + "***" + rawURL[schemeEnd+3+atIdx:]
}
```

Providers that implement this: `mq/rabbitmq` (`maskURL`), `db/mongo` (`maskURI`).

## Resource Leak Prevention

### Rule: clean up on partial Init failure

If `Init()` opens resource A, then fails at resource B, resource A must be closed before returning:

```go
func (c *Client) Init() error {
    conn, err := openConnection(...)
    if err != nil { return err }

    if err := conn.Ping(); err != nil {
        conn.Close()  // ← clean up A before returning
        return fmt.Errorf("...: %w", err)
    }

    c.conn = conn  // assign only after all checks pass
    return nil
}
```

### Rule: assign to struct field LAST

Always assign `c.client = ...` **after** all validation succeeds. This ensures `Close()` only runs on a fully-initialized client.

```go
// ✅ Correct
conn, err := connect()
if err != nil { return err }
if err := validate(conn); err != nil {
    conn.Close()
    return err
}
c.client = conn  // assigned last

// ❌ Wrong — Close() would double-close on error
c.client = conn
if err := validate(conn); err != nil {
    return err
}
```

## Goroutine Safety

### signal.Notify cleanup
Always call `signal.Stop(ch)` after `signal.Notify`:

```go
signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
defer signal.Stop(ch)  // ← prevents channel/registration leak
```

### shared registries (maps)
Registry maps (like GORM's `dialectors`, ent's `drivers`) must be protected:

```go
var (
    dialectorsMu sync.RWMutex
    dialectors   = map[string]func(string) gorm.Dialector{...}
)

func RegisterDialector(name string, fn func(string) gorm.Dialector) {
    dialectorsMu.Lock()
    defer dialectorsMu.Unlock()
    dialectors[name] = fn
}
```

## Error Wrapping Convention

All errors from `Init()`, `Run()`, and `Close()` must include the package name prefix:

```go
// ✅ Correct
return fmt.Errorf("redis: ping failed: %w", err)
return fmt.Errorf("nats: connect failed: %w", err)
return fmt.Errorf("otel: OTLP exporter error: %w", err)

// ❌ Wrong — no context, hard to trace in logs
return err
return fmt.Errorf("connect failed: %w", err)
```

## Viper.GetStringSlice Gotcha

`viper.GetStringSlice("KEY")` does **not** split comma-separated strings read from environment variables. Use a `splitCSV()` helper instead:

```go
// ❌ Wrong — won't split "a,b,c" from env var
addresses := v.GetStringSlice("ADDRS")

// ✅ Correct
func splitCSV(s string) []string {
    if s == "" { return nil }
    parts := strings.Split(s, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        if t := strings.TrimSpace(p); t != "" {
            out = append(out, t)
        }
    }
    return out
}
addresses := splitCSV(v.GetString("ADDRS"))
```

Providers using this pattern: `db/clickhouse`, `mq/nats`, `mq/nsq`.
