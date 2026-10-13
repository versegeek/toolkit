---
name: Bug report
about: Report a bug in an existing provider
labels: bug
---

## Provider

<!-- Which provider is affected? e.g. db/gorm, cache/redis, trace/otel -->

**Package:** `<category>/<name>`

## Environment

| Field | Value |
|-------|-------|
| Go version | `go version` output |
| OS / Arch | e.g. macOS arm64, Linux amd64 |
| toolkit version / commit | e.g. v0.1.0 or commit SHA |

## Bug Description

<!-- A clear and concise description of what the bug is. -->

## Steps to Reproduce

1. 
2. 
3. 

## Expected Behaviour

<!-- What you expected to happen. -->

## Actual Behaviour

<!-- What actually happened. Include any error messages or stack traces. -->

```
<error output here>
```

## Minimal Reproduction

<!-- A minimal Go snippet that reproduces the problem. -->

```go
package main

import (
    // …
)

func main() {
    // reproduces the bug
}
```

## Additional Context

<!-- Any other context (config values, env vars, related issues). Do NOT include passwords or secrets. -->
