//go:build wireinject
// +build wireinject

package ent

import "github.com/google/wire"

// ProviderSet wires Config + ent MySQL driver together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
