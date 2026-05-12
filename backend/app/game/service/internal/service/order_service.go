package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"

	gameV1 "go-wind-cms/api/gen/go/game/service/v1"

	"go-wind-cms/app/game/service/internal/data"
)

type OrderService struct {
	gameV1.UnimplementedOrderServiceServer

	log *log.Helper

	orderRepo *data.OrderRepo
}

func NewOrderService(ctx *bootstrap.Context, orderRepo *data.OrderRepo) *OrderService {
	return &OrderService{
		log:       ctx.NewLoggerHelper("order/service/game-service"),
		orderRepo: orderRepo,
	}
}

func (s *OrderService) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListOrderResponse, error) {
	return s.orderRepo.List(ctx, req)
}

func (s *OrderService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.CountOrderResponse, error) {
	count, err := s.orderRepo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &gameV1.CountOrderResponse{Count: uint64(count)}, nil
}

func (s *OrderService) Get(ctx context.Context, req *gameV1.GetOrderRequest) (*gameV1.Order, error) {
	return s.orderRepo.Get(ctx, req)
}

func (s *OrderService) Create(ctx context.Context, req *gameV1.CreateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.orderRepo.Create(ctx, req.Data); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderService) Update(ctx context.Context, req *gameV1.UpdateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.orderRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderService) Delete(ctx context.Context, req *gameV1.DeleteOrderRequest) (*emptypb.Empty, error) {
	if err := s.orderRepo.Delete(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
