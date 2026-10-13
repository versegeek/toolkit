//go:build wireinject
// +build wireinject

package oauth2

import "github.com/google/wire"

// ProviderSet wires Config + OAuth2 client together.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)
