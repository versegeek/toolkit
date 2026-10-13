---
name: Feature request
about: Propose a new provider or enhancement to an existing one
labels: enhancement
---

## Feature Type

- [ ] New provider (new package under `db/`, `cache/`, `mq/`, `server/`, `log/`, `trace/`, `auth/`)
- [ ] Enhancement to an existing provider
- [ ] New utility / helper

## Proposed Package Path

<!-- e.g. mq/kafka, db/cassandra, cache/valkey -->

`<category>/<name>`

## Problem / Motivation

<!-- What problem does this solve? What use case does it enable? -->

## Proposed API

<!-- Sketch the Config struct and key methods. -->

```go
type Config struct {
    // …
}

func NewConfigFromEnv() *Config { … }

type Client struct { … }

func (c *Client) Init() error  { … }
func (c *Client) Close() error { … }
func (c *Client) Client() *underlying.Client { … }
```

## Reference Implementations

<!-- Links to the underlying SDK or similar providers in other libraries. -->

- SDK: 
- Similar provider example: 

## Alternatives Considered

<!-- Any alternative approaches you considered. -->

## Willing to Implement

- [ ] Yes, I can submit a PR for this
- [ ] No, requesting someone else implement it
