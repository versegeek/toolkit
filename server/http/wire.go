//go:build wireinject
// +build wireinject

package http

import "github.com/google/wire"

// ProviderSet wires Config + HTTP server together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)
