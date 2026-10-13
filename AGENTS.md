# toolkit

Go library (`github.com/versegeek/toolkit`) of lifecycle-managed providers for databases, caches, message queues, servers, logging, tracing and auth, composable by hand (`lifecycle.Group`) or with Google Wire (`lifecycle.Run`).

Read `GLOSSARY.md` for the vocabulary and `docs/adr/` before changing how providers are built, run or closed.

## Commands

```bash
make build        # go build ./...
make test         # go test -race ./...
make lint         # go fmt + go vet + golangci-lint (v2 config in .golangci.yml)
make wire         # regenerate examples/wire/wire_gen.go (wire v0.7.0); never edit it by hand
make wire-check   # CI check that wire_gen.go is current
make tidy         # go mod tidy (GOPROXY=https://goproxy.cn,direct by default)
```

## Layout

- `provider/`: `Provider`, `RunProvider`, `Enabler`, `InitProvider`. `lifecycle/`: `Run` and `Group`. `config/`: Viper loading.
- Providers live in existing category directories only: `cache/`, `db/`, `mq/`, `server/`, `log/`, `trace/`, `auth/`. The package name is the backend name (`redis`, `gorm`, `http`, `zap`…); callers alias it when it clashes with the upstream SDK.
- `internal/` is shared plumbing: `httpserver` (bind-then-ready, graceful shutdown), `sqlconn` (`DB_*` keys, DSNs, pools), `redact` (passwords in URLs), `logbridge` (slog default swap).
- `examples/group` and `examples/wire` must keep compiling; `go build ./...` covers them.

## Provider conventions

Each provider package has `config.go` (`Config`, `NewConfigFromEnv() (*Config, error)` documenting every key and default) and `<name>.go` (package doc, `ProviderSet`, the `Client` or `Server` type, `var _ provider.Provider = …`, `New`, `NewClient`/`NewServer`, `Init`, `Close`, `Run` for servers). Split out more files only for separate concerns.

- `New(cfg)` does no I/O. `Init` loads the config when `Config` is nil, assigns fields only after every check passes, and closes anything it opened before returning an error.
- `Close` is safe before `Init` and after a failed `Init`. A run provider's `Close` makes `Run` return and may be called twice; set `SetRunning(true)` only once the listener is bound.
- Wrap errors as `fmt.Errorf("<package>: <action> failed: %w", err)`.
- Log with `log/slog` only, never a specific backend, and never log credentials (`internal/redact`).
- Bound network calls in `Init`/`Close` with timeouts; zero config values keep the upstream default.
- Provider packages never import each other.

## Tests

Tests are being added after the refactor. Use external test packages (`package <name>_test`) with testify, run with `-race`, and skip, rather than wait, when an external service is unavailable.

## Agent skills

### Issue tracker

Issues and specs live as local markdown files under `.scratch/`. See `docs/agents/issue-tracker.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
