//go:build wireinject
// +build wireinject

package nats

import "github.com/google/wire"

// ProviderSet wires Config + NATS client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
