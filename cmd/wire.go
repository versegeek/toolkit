//go:build wireinject
// +build wireinject

// cmd/wire.go — Wire dependency injection entry point.
//
// Wire reads this file and generates cmd/wire_gen.go containing InitApp().
// Never edit wire_gen.go manually.
//
// To regenerate:
//
//	make wire
package main

import (
	"github.com/google/wire"
	"github.com/versegeek/toolkit/cache/redis"
	zaplog "github.com/versegeek/toolkit/log/zap"
)

// InitApp wires all providers together and returns the assembled App.
// The returned cleanup function closes all providers in reverse order.
func InitApp() (*App, func(), error) {
	wire.Build(
		zaplog.ProviderSet,
		redis.ProviderSet,
		NewApp,
	)
	return nil, nil, nil
}
