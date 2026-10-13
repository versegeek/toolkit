// Command cmd is a temporary example that wires zap and Redis together.
// It moves to examples/wire once the providers are reworked.
package main

import (
	"context"
	"log"

	"github.com/versegeek/toolkit/lifecycle"
)

func main() {
	_, cleanup, err := InitApp()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	if err := lifecycle.Run(context.Background(), cleanup); err != nil {
		log.Fatal(err)
	}
}
