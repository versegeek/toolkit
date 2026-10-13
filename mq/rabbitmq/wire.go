//go:build wireinject
// +build wireinject

package rabbitmq

import "github.com/google/wire"

// ProviderSet wires Config + RabbitMQ client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
