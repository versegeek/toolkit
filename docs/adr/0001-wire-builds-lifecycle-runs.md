# Wire builds providers; the lifecycle runs them

Wire only constructs and initialises providers (through `provider.InitProvider`) and returns their combined cleanup; it never calls `Run`. Run providers are started by `lifecycle.Run` or `Group.Run` after the application has registered its routes and services, and shutdown invokes the cleanup exactly once.

## Considered Options

- Start `Run` inside `InitProvider`, as `lib-core-go` does. Rejected: servers begin serving before routes are registered, a failing `Run` cannot trigger the cleanup, and an injector cannot be built in a test without binding ports.

## Consequences

- Shutdown stops run providers first (in reverse order) and then calls the cleanup, which closes them again, so a run provider's `Close` must be safe to call more than once.
- `InitProvider`'s cleanup is idempotent, so calling it from both `defer` and the lifecycle is harmless.
