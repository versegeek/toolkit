# One module with a Go 1.25 floor

The toolkit stays a single Go module declaring `go 1.25.0`: the highest version any dependency requires and the version its consumers use. Consumers compile only the provider packages they import, at the cost of a larger shared `go.sum`.

## Considered Options

- One module per provider. Rejected for now: a single maintainer would have to tag and release up to twenty modules in lockstep.
