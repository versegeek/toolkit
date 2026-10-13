# SQL providers share DB_* keys and one SQLite driver

`db/gorm` and `db/ent` read the same `DB_*` configuration keys, `DB_DRIVER` has no default, SQLite always means the pure-Go `github.com/glebarez/go-sqlite` driver, and an in-memory SQLite database is limited to a single pooled connection. PostgreSQL goes through pgx for both ORMs.

## Considered Options

- Per-backend prefixes (`MYSQL_*`, `POSTGRES_*`, `CLICKHOUSE_*`) with a `sqlite` + `:memory:` default. Rejected: a service talks to one SQL database, and a silent in-memory default loses data in production.
- Keeping `modernc.org/sqlite` for ent. Rejected: it registers the same `sqlite` driver name as the gorm driver, so importing both ORMs panicked at start-up.

## Consequences

- The native `db/clickhouse` provider keeps its own `CLICKHOUSE_*` keys; it is a separate client, not one of the SQL ORMs.
- Every pooled connection to `:memory:` would otherwise open its own empty database.
