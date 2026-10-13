package main

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"github.com/versegeek/toolkit/cache/redis"
	zaplog "github.com/versegeek/toolkit/log/zap"
	"github.com/versegeek/toolkit/server/http"
)

// App holds the providers the example uses.
type App struct {
	Log   *zaplog.Client
	Redis *redis.Client
	HTTP  *http.Server
}

// NewApp registers the routes. Wire calls it once every provider is
// initialised and before anything runs, so the routes exist before the
// server accepts requests.
func NewApp(logger *zaplog.Client, rdb *redis.Client, srv *http.Server) *App {
	srv.Router.GET("/ping", func(c *gin.Context) {
		n, err := rdb.Client().Incr(c, "pings").Result()
		if err != nil {
			c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(nethttp.StatusOK, gin.H{"pings": n})
	})
	return &App{Log: logger, Redis: rdb, HTTP: srv}
}
