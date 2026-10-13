## Description

<!-- A brief description of what this PR does and why. -->

## Type of Change

- [ ] 🆕 New provider
- [ ] 🐛 Bug fix
- [ ] ♻️ Refactor (no behaviour change)
- [ ] 📝 Documentation
- [ ] 🧪 Tests only
- [ ] ⚙️ CI / tooling

## Pre-flight Checklist

- [ ] `make build` passes (`wire` codegen + `go build ./...`)
- [ ] `go test -race ./...` passes with zero failures
- [ ] `make lint` produces no errors (`go fmt` / `go vet` / `golangci-lint`)
- [ ] New provider: `var _ provider.Provider = (*Client)(nil)` compile-time check added
- [ ] New provider: `Close()` handles the case where `Init()` was never called (nil guard)
- [ ] New provider: `wire.go` with `ProviderSet` added

## New Provider Details

<!-- Fill in if this PR adds a new provider, otherwise delete this section. -->

| Field | Value |
|-------|-------|
| Package path | e.g. `mq/kafka` |
| Env var prefix | e.g. `KAFKA_` |
| Tested without real service? | yes / no |

## Breaking Changes

<!-- Does this PR change any public API, env var names, or default behaviour? -->

- [ ] No breaking changes
- [ ] Yes — describe below:

<!-- Breaking change description here -->

## Related Issues

Closes #
