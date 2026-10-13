//go:build wireinject

package otel

import (
	"github.com/google/wire"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet is the Wire provider set for the OpenTelemetry tracing provider.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

var _ = provider.InitProvider[*Client]
