//go:build wireinject
// +build wireinject

package nsq

import "github.com/google/wire"

// ProviderSet wires Config + NSQ producer client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
