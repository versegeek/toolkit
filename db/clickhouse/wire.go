//go:build wireinject
// +build wireinject

package clickhouse

import "github.com/google/wire"

// ProviderSet wires Config + native ClickHouse client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
