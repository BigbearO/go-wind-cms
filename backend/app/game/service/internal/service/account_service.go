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

type AccountService struct {
	gameV1.UnimplementedAccountServiceServer

	log *log.Helper

	accountRepo *data.AccountRepo
}

func NewAccountService(ctx *bootstrap.Context, accountRepo *data.AccountRepo) *AccountService {
	return &AccountService{
		log:         ctx.NewLoggerHelper("account/service/game-service"),
		accountRepo: accountRepo,
	}
}

func (s *AccountService) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListAccountResponse, error) {
	return s.accountRepo.List(ctx, req)
}

func (s *AccountService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.CountAccountResponse, error) {
	count, err := s.accountRepo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &gameV1.CountAccountResponse{Count: uint64(count)}, nil
}

func (s *AccountService) Get(ctx context.Context, req *gameV1.GetAccountRequest) (*gameV1.Account, error) {
	return s.accountRepo.Get(ctx, req)
}

func (s *AccountService) Create(ctx context.Context, req *gameV1.CreateAccountRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.accountRepo.Create(ctx, req.Data); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *AccountService) Update(ctx context.Context, req *gameV1.UpdateAccountRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.accountRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *AccountService) Delete(ctx context.Context, req *gameV1.DeleteAccountRequest) (*emptypb.Empty, error) {
	if err := s.accountRepo.Delete(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *AccountService) RenewMembership(ctx context.Context, req *gameV1.RenewMembershipRequest) (*emptypb.Empty, error) {
	if err := s.accountRepo.RenewMembership(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
