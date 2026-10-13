//go:build wireinject
// +build wireinject

package grpc

import "github.com/google/wire"

// ProviderSet wires Config + gRPC server together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)
