// Command wire shows Wire composition: a generated injector builds and
// initialises zap, Redis and an HTTP server, and lifecycle.Run serves until
// SIGINT or SIGTERM, then closes everything once in reverse order.
//
//	REDIS_ADDRESS=127.0.0.1:6379 go run ./examples/wire
//	curl localhost:8080/ping
//
// After changing wire.go, regenerate wire_gen.go with make wire.
package main

import (
	"context"
	"log"

	"github.com/versegeek/toolkit/lifecycle"
)

func main() {
	app, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	if err := lifecycle.Run(context.Background(), cleanup, app.HTTP); err != nil {
		log.Fatal(err)
	}
}
