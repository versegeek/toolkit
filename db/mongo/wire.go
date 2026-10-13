//go:build wireinject
// +build wireinject

package mongo

import "github.com/google/wire"

// ProviderSet wires Config + MongoDB client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
