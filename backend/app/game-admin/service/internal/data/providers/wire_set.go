//go:build wireinject
// +build wireinject

//go:generate go run github.com/google/wire/cmd/wire

package providers

import (
	"github.com/google/wire"

	"go-wind-cms/app/game-admin/service/internal/data"
)

// ProviderSet is the Wire provider set for data layer.
var ProviderSet = wire.NewSet(
	data.NewRedisClient,
	data.NewDiscovery,
	data.NewClientType,
	data.NewAuthorizer,
	data.NewAuthenticationServiceClient,
	data.NewOrderServiceClient,
	data.NewAccountServiceClient,
)
