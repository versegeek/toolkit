//go:build wireinject
// +build wireinject

package pprof

import "github.com/google/wire"

// ProviderSet wires Config + PProf server together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)
