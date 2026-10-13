//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/versegeek/toolkit/cache/redis"
	zaplog "github.com/versegeek/toolkit/log/zap"
	"github.com/versegeek/toolkit/server/http"
)

// InitApp builds and initialises the providers in dependency order and
// returns a cleanup that closes them in reverse order.
func InitApp() (*App, func(), error) {
	panic(wire.Build(
		zaplog.ProviderSet,
		redis.ProviderSet,
		http.ProviderSet,
		NewApp,
	))
}
