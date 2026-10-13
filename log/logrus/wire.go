//go:build wireinject
// +build wireinject

package logprovider

import "github.com/google/wire"

// ProviderSet wires Config + logrus client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
