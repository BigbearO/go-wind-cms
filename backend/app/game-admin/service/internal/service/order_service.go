package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"

	gameV1 "go-wind-cms/api/gen/go/game/service/v1"
	gameadminV1 "go-wind-cms/api/gen/go/game-admin/service/v1"
)

type OrderService struct {
	gameadminV1.OrderServiceHTTPServer

	log *log.Helper

	orderServiceClient   gameV1.OrderServiceClient
	accountServiceClient gameV1.AccountServiceClient
}

func NewOrderService(
	ctx *bootstrap.Context,
	orderServiceClient gameV1.OrderServiceClient,
	accountServiceClient gameV1.AccountServiceClient,
) *OrderService {
	return &OrderService{
		log:                  ctx.NewLoggerHelper("order/service/game-admin-service"),
		orderServiceClient:   orderServiceClient,
		accountServiceClient: accountServiceClient,
	}
}

func (s *OrderService) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListOrderResponse, error) {
	return s.orderServiceClient.List(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, req *gameV1.GetOrderRequest) (*gameV1.Order, error) {
	return s.orderServiceClient.Get(ctx, req)
}

func (s *OrderService) Create(ctx context.Context, req *gameV1.CreateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	return s.orderServiceClient.Create(ctx, req)
}

func (s *OrderService) Update(ctx context.Context, req *gameV1.UpdateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	return s.orderServiceClient.Update(ctx, req)
}

func (s *OrderService) Delete(ctx context.Context, req *gameV1.DeleteOrderRequest) (*emptypb.Empty, error) {
	return s.orderServiceClient.Delete(ctx, req)
}

func (s *OrderService) CreateOrderAndRenew(ctx context.Context, req *gameadminV1.CreateOrderAndRenewRequest) (*emptypb.Empty, error) {
	if req == nil || req.Order == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.orderServiceClient.Create(ctx, &gameV1.CreateOrderRequest{Data: req.Order}); err != nil {
		return nil, err
	}

	if _, err := s.accountServiceClient.RenewMembership(ctx, &gameV1.RenewMembershipRequest{
		UserAccount: req.Order.GetUserId(),
		Type:        req.Order.GetGameType(),
		Days:        req.Order.GetDays(),
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
