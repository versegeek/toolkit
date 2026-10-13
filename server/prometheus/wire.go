//go:build wireinject
// +build wireinject

package prometheus

import "github.com/google/wire"

// ProviderSet wires Config + Prometheus server together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)
