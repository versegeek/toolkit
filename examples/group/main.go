// Command group shows manual composition: a lifecycle.Group initialises zap,
// Redis and an HTTP server in order, runs the server until SIGINT or SIGTERM,
// then closes everything once in reverse order.
//
//	REDIS_ADDRESS=127.0.0.1:6379 go run ./examples/group
//	curl localhost:8080/ping
package main

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"github.com/versegeek/toolkit/cache/redis"
	"github.com/versegeek/toolkit/lifecycle"
	zaplog "github.com/versegeek/toolkit/log/zap"
	"github.com/versegeek/toolkit/server/http"
)

func main() {
	g := lifecycle.New()
	// Releases whatever was initialised if a later MustInit panics.
	defer g.MustClose()

	// The log provider goes first so that the other providers log through it.
	g.MustInit(zaplog.New(nil))
	rdb := redis.New(nil)
	g.MustInit(rdb)
	srv := http.New(nil)
	g.MustInit(srv)

	srv.Router.GET("/ping", func(c *gin.Context) {
		n, err := rdb.Client().Incr(c, "pings").Result()
		if err != nil {
			c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(nethttp.StatusOK, gin.H{"pings": n})
	})

	g.MustRun()
}
