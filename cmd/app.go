package main

import (
	"github.com/versegeek/toolkit/cache/redis"
	zaplog "github.com/versegeek/toolkit/log/zap"
)

// App holds references to all managed providers.
// Add new provider fields here and update NewApp accordingly.
type App struct {
	Log   *zaplog.Client
	Redis *redis.Client
}

// NewApp assembles the App from Wire-injected providers.
// Wire calls this after all providers have been initialised.
func NewApp(l *zaplog.Client, r *redis.Client) *App {
	return &App{Log: l, Redis: r}
}
