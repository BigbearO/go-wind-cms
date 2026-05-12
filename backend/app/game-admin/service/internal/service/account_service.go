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

type AccountService struct {
	gameadminV1.AccountServiceHTTPServer

	log *log.Helper

	accountServiceClient gameV1.AccountServiceClient
}

func NewAccountService(ctx *bootstrap.Context, accountServiceClient gameV1.AccountServiceClient) *AccountService {
	return &AccountService{
		log:                  ctx.NewLoggerHelper("account/service/game-admin-service"),
		accountServiceClient: accountServiceClient,
	}
}

func (s *AccountService) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListAccountResponse, error) {
	return s.accountServiceClient.List(ctx, req)
}

func (s *AccountService) Get(ctx context.Context, req *gameV1.GetAccountRequest) (*gameV1.Account, error) {
	return s.accountServiceClient.Get(ctx, req)
}

func (s *AccountService) Create(ctx context.Context, req *gameV1.CreateAccountRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	return s.accountServiceClient.Create(ctx, req)
}

func (s *AccountService) Update(ctx context.Context, req *gameV1.UpdateAccountRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	return s.accountServiceClient.Update(ctx, req)
}

func (s *AccountService) Delete(ctx context.Context, req *gameV1.DeleteAccountRequest) (*emptypb.Empty, error) {
	return s.accountServiceClient.Delete(ctx, req)
}

func (s *AccountService) RenewMembership(ctx context.Context, req *gameV1.RenewMembershipRequest) (*emptypb.Empty, error) {
	return s.accountServiceClient.RenewMembership(ctx, req)
}
