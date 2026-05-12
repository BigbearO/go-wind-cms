package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	entCrud "github.com/tx7do/go-crud/entgo"

	"go-wind-cms/app/game/service/internal/data/ent"
	"go-wind-cms/app/game/service/internal/data/ent/gameaccount"
	"go-wind-cms/app/game/service/internal/data/ent/predicate"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	gameV1 "go-wind-cms/api/gen/go/game/service/v1"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
)

type AccountRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	repository *entCrud.Repository[
		ent.GameAccountQuery, ent.GameAccountSelect,
		ent.GameAccountCreate, ent.GameAccountCreateBulk,
		ent.GameAccountUpdate, ent.GameAccountUpdateOne,
		ent.GameAccountDelete,
		predicate.GameAccount,
		gameV1.Account, ent.GameAccount,
	]
}

func NewAccountRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *AccountRepo {
	r := &AccountRepo{
		log:       ctx.NewLoggerHelper("account/repo/game-service"),
		entClient: entClient,
	}

	r.init()

	return r
}

func (r *AccountRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.GameAccountQuery, ent.GameAccountSelect,
		ent.GameAccountCreate, ent.GameAccountCreateBulk,
		ent.GameAccountUpdate, ent.GameAccountUpdateOne,
		ent.GameAccountDelete,
		predicate.GameAccount,
		gameV1.Account, ent.GameAccount,
	](r.mapper())
}

func (r *AccountRepo) mapper() *mapper.CopierMapper[gameV1.Account, ent.GameAccount] {
	m := mapper.NewCopierMapper[gameV1.Account, ent.GameAccount]()

	m.AppendConverters(copierutil.NewTimeStringConverterPair())
	m.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	return m
}

func (r *AccountRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().GameAccount.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query account count failed: %s", err.Error())
		return 0, gameV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *AccountRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListAccountResponse, error) {
	if req == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GameAccount.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &gameV1.ListAccountResponse{Total: 0, Items: nil}, nil
	}

	return &gameV1.ListAccountResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *AccountRepo) Get(ctx context.Context, req *gameV1.GetAccountRequest) (*gameV1.Account, error) {
	if req == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GameAccount.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *gameV1.GetAccountRequest_Id:
		whereCond = append(whereCond, gameaccount.IDEQ(int64(req.GetId())))

	case *gameV1.GetAccountRequest_UserAccount:
		whereCond = append(whereCond, gameaccount.UserAccountEQ(req.GetUserAccount()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *AccountRepo) Create(ctx context.Context, data *gameV1.Account) error {
	if data == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GameAccount.Create().
		SetUserAccount(data.GetUserAccount()).
		SetNillableCredential(data.Credential).
		SetType(data.GetType()).
		SetEnabled(data.GetEnabled()).
		SetNillableEmulatorID(data.EmulatorId).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(time.Now())

	if data.StartTime != nil {
		builder.SetStartTime(data.StartTime.AsTime())
	}
	if data.ExpireTime != nil {
		builder.SetExpireTime(data.ExpireTime.AsTime())
	}

	if _, err := builder.Save(ctx); err != nil {
		r.log.Errorf("insert account failed: %s", err.Error())
		return gameV1.ErrorInternalServerError("insert account failed")
	}

	return nil
}

func (r *AccountRepo) Update(ctx context.Context, req *gameV1.UpdateAccountRequest) error {
	if req == nil || req.Data == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GameAccount.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *gameV1.Account) {
			builder.
				SetNillableUserAccount(req.Data.UserAccount).
				SetNillableCredential(req.Data.Credential).
				SetNillableType(req.Data.Type).
				SetNillableEnabled(req.Data.Enabled).
				SetNillableEmulatorID(req.Data.EmulatorId).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(gameaccount.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *AccountRepo) Delete(ctx context.Context, req *gameV1.DeleteAccountRequest) error {
	if req == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().GameAccount.DeleteOneID(int64(req.GetId())).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return gameV1.ErrorNotFound("account not found")
		}

		r.log.Errorf("delete account failed: %s", err.Error())
		return gameV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

// GetByUserAccountAndType finds an account by user_account and type.
func (r *AccountRepo) GetByUserAccountAndType(ctx context.Context, userAccount string, accountType int32) (*gameV1.Account, error) {
	entity, err := r.entClient.Client().GameAccount.Query().
		Where(
			gameaccount.UserAccountEQ(userAccount),
			gameaccount.TypeEQ(accountType),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("query account by user_account and type failed: %s", err.Error())
		return nil, gameV1.ErrorInternalServerError("query account failed")
	}

	dto := r.mapper().ToDTO(entity)
	return dto, nil
}

// RenewMembership renews the membership expire_time for an account.
func (r *AccountRepo) RenewMembership(ctx context.Context, req *gameV1.RenewMembershipRequest) error {
	if req == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	if req.GetDays() <= 0 {
		return gameV1.ErrorInvalidDays("days must be greater than 0")
	}

	account, err := r.GetByUserAccountAndType(ctx, req.GetUserAccount(), req.GetType())
	if err != nil {
		return err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if account == nil {
		// Account does not exist - create new with start_time=today, expire_time=today+days
		expire := today.AddDate(0, 0, int(req.GetDays()))
		builder := r.entClient.Client().GameAccount.Create().
			SetUserAccount(req.GetUserAccount()).
			SetType(req.GetType()).
			SetEnabled(true).
			SetStartTime(today).
			SetExpireTime(expire).
			SetCreatedAt(time.Now())

		if _, saveErr := builder.Save(ctx); saveErr != nil {
			r.log.Errorf("create account for renewal failed: %s", saveErr.Error())
			return gameV1.ErrorInternalServerError("create account failed")
		}

		return nil
	}

	// Account exists - calculate new expire_time
	var newExpire time.Time
	if account.ExpireTime != nil && account.ExpireTime.AsTime().After(today) {
		// Not expired - add days to current expire_time
		newExpire = account.ExpireTime.AsTime().AddDate(0, 0, int(req.GetDays()))
	} else {
		// Expired or no expire_time - start from today
		newExpire = today.AddDate(0, 0, int(req.GetDays()))
	}

	_, updateErr := r.entClient.Client().GameAccount.Update().
		Where(gameaccount.IDEQ(int64(account.GetId()))).
		SetExpireTime(newExpire).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if updateErr != nil {
		r.log.Errorf("update account expire_time failed: %s", updateErr.Error())
		return gameV1.ErrorInternalServerError("update account failed")
	}

	return nil
}
