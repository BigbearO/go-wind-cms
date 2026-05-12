//go:build wireinject
// +build wireinject

package providers

import (
	"github.com/google/wire"

	"go-wind-cms/app/game/service/internal/data"
)

var ProviderSet = wire.NewSet(
	data.NewEntClient,
	data.NewDiscovery,
	data.NewOrderRepo,
	data.NewAccountRepo,
)
