//go:build wireinject
// +build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package providers

import (
	"github.com/google/wire"

	"go-wind-cms/app/game-admin/service/internal/server"
)

// ProviderSet is the Wire provider set for server layer.
var ProviderSet = wire.NewSet(
	server.NewRestServer,
	server.NewRestMiddleware,
)
