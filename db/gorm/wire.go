//go:build wireinject
// +build wireinject

package gorm

import "github.com/google/wire"

// ProviderSet wires Config + GORM MySQL client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
