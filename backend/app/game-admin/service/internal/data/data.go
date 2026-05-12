package data

import (
	"github.com/redis/go-redis/v9"

	authzEngine "github.com/tx7do/kratos-authz/engine"
	"github.com/tx7do/kratos-authz/engine/noop"

	"github.com/go-kratos/kratos/v2/registry"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	redisClient "github.com/tx7do/kratos-bootstrap/cache/redis"
	bRegistry "github.com/tx7do/kratos-bootstrap/registry"
	"github.com/tx7do/kratos-bootstrap/rpc"

	authenticationV1 "go-wind-cms/api/gen/go/authentication/service/v1"
	gameV1 "go-wind-cms/api/gen/go/game/service/v1"

	"go-wind-cms/pkg/serviceid"
)

func NewClientType() authenticationV1.ClientType {
	return authenticationV1.ClientType_admin
}

func NewAuthorizer() authzEngine.Engine {
	return noop.State{}
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(ctx *bootstrap.Context) (*redis.Client, func(), error) {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil, func() {}, nil
	}

	l := ctx.NewLoggerHelper("redis/data/game-admin-service")

	cli := redisClient.NewClient(cfg.Data, l)

	return cli, func() {
		if err := cli.Close(); err != nil {
			l.Error(err)
		}
	}, nil
}

// NewDiscovery 创建服务发现客户端
func NewDiscovery(ctx *bootstrap.Context) registry.Discovery {
	cfg := ctx.GetConfig()
	if cfg == nil {
		return nil
	}

	discovery, err := bRegistry.NewDiscovery(cfg.Registry)
	if err != nil {
		return nil
	}

	return discovery
}

func NewAuthenticationServiceClient(ctx *bootstrap.Context, r registry.Discovery) authenticationV1.AuthenticationServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewDiscoveryName(serviceid.CoreService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return authenticationV1.NewAuthenticationServiceClient(cli)
}

func NewOrderServiceClient(ctx *bootstrap.Context, r registry.Discovery) gameV1.OrderServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewGrpcDiscoveryName(serviceid.GameService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return gameV1.NewOrderServiceClient(cli)
}

func NewAccountServiceClient(ctx *bootstrap.Context, r registry.Discovery) gameV1.AccountServiceClient {
	cli, err := rpc.CreateGrpcClient(ctx.Context(), r, serviceid.NewGrpcDiscoveryName(serviceid.GameService), ctx.GetConfig())
	if err != nil {
		return nil
	}

	return gameV1.NewAccountServiceClient(cli)
}
