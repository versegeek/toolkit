# Configuration loads in Init, not in New

`New(nil)` performs no I/O and leaves `Config` nil; `Init` then calls `NewConfigFromEnv`, which returns an error for an unreadable or malformed config file instead of silently falling back to defaults (a missing file is fine). Constructors therefore never panic or swallow configuration errors, and Wire composition passes the config explicitly through the provider set.
