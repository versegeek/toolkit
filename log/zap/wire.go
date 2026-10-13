//go:build wireinject

package zaplog

import (
	"github.com/google/wire"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet is the Wire provider set for the zap logging provider.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

var _ = provider.InitProvider[*Client]
