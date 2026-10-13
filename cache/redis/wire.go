//go:build wireinject
// +build wireinject

package redis

import "github.com/google/wire"

// ProviderSet wires Config + single-node Redis together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// ClusterProviderSet wires Config + Redis Cluster together.
var ClusterProviderSet = wire.NewSet(NewConfigFromEnv, NewClusterClient)
