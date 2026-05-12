package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	entCrud "github.com/tx7do/go-crud/entgo"

	"go-wind-cms/app/game/service/internal/data/ent"
	"go-wind-cms/app/game/service/internal/data/ent/predicate"
	"go-wind-cms/app/game/service/internal/data/ent/tradeorder"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	gameV1 "go-wind-cms/api/gen/go/game/service/v1"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
)

type OrderRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	repository *entCrud.Repository[
		ent.TradeOrderQuery, ent.TradeOrderSelect,
		ent.TradeOrderCreate, ent.TradeOrderCreateBulk,
		ent.TradeOrderUpdate, ent.TradeOrderUpdateOne,
		ent.TradeOrderDelete,
		predicate.TradeOrder,
		gameV1.Order, ent.TradeOrder,
	]
}

func NewOrderRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *OrderRepo {
	r := &OrderRepo{
		log:       ctx.NewLoggerHelper("order/repo/game-service"),
		entClient: entClient,
	}

	r.init()

	return r
}

func (r *OrderRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.TradeOrderQuery, ent.TradeOrderSelect,
		ent.TradeOrderCreate, ent.TradeOrderCreateBulk,
		ent.TradeOrderUpdate, ent.TradeOrderUpdateOne,
		ent.TradeOrderDelete,
		predicate.TradeOrder,
		gameV1.Order, ent.TradeOrder,
	](r.mapper())
}

func (r *OrderRepo) mapper() *mapper.CopierMapper[gameV1.Order, ent.TradeOrder] {
	m := mapper.NewCopierMapper[gameV1.Order, ent.TradeOrder]()

	m.AppendConverters(copierutil.NewTimeStringConverterPair())
	m.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	return m
}

func (r *OrderRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().TradeOrder.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query order count failed: %s", err.Error())
		return 0, gameV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *OrderRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*gameV1.ListOrderResponse, error) {
	if req == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TradeOrder.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &gameV1.ListOrderResponse{Total: 0, Items: nil}, nil
	}

	return &gameV1.ListOrderResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *OrderRepo) Get(ctx context.Context, req *gameV1.GetOrderRequest) (*gameV1.Order, error) {
	if req == nil {
		return nil, gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TradeOrder.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *gameV1.GetOrderRequest_Id:
		whereCond = append(whereCond, tradeorder.IDEQ(int64(req.GetId())))

	case *gameV1.GetOrderRequest_OrderNo:
		whereCond = append(whereCond, tradeorder.OrderNoEQ(req.GetOrderNo()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *OrderRepo) Create(ctx context.Context, data *gameV1.Order) error {
	if data == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TradeOrder.Create().
		SetOrderNo(data.GetOrderNo()).
		SetUserID(data.GetUserId()).
		SetNillableOrderType(enumInt32Ptr(data.OrderType)).
		SetNillableStatus(enumInt32Ptr(data.Status)).
		SetNillableGameType(data.GameType).
		SetNillableDays(data.Days).
		SetPayAmount(data.GetPayAmount()).
		SetNillablePayType(data.PayType)

	if data.PayTime != nil {
		builder.SetPayTime(data.PayTime.AsTime())
	}
	if data.ExpireTime != nil {
		builder.SetExpireTime(data.ExpireTime.AsTime())
	}

	builder.
		SetNillableCancelReason(data.CancelReason).
		SetNillableCancelTime(func() *time.Time {
			if t := data.GetCancelTime(); t != nil {
				v := t.AsTime()
				return &v
			}
			return nil
		}()).
		SetNillableFreightFee(data.FreightFee).
		SetNillableRemark(data.Remark).
		SetNillableCreatedBy(data.CreatedBy).
		SetCreatedAt(time.Now())

	if _, err := builder.Save(ctx); err != nil {
		r.log.Errorf("insert order failed: %s", err.Error())
		return gameV1.ErrorInternalServerError("insert order failed")
	}

	return nil
}

func (r *OrderRepo) Update(ctx context.Context, req *gameV1.UpdateOrderRequest) error {
	if req == nil || req.Data == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TradeOrder.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *gameV1.Order) {
			builder.
				SetNillableOrderNo(req.Data.OrderNo).
				SetNillableUserID(req.Data.UserId).
				SetNillableOrderType(enumInt32Ptr(req.Data.OrderType)).
				SetNillableStatus(enumInt32Ptr(req.Data.Status)).
				SetNillableGameType(req.Data.GameType).
				SetNillableDays(req.Data.Days).
				SetNillablePayAmount(req.Data.PayAmount).
				SetNillableCancelReason(req.Data.CancelReason).
				SetNillableRefundChannel(req.Data.RefundChannel).
				SetNillableRemark(req.Data.Remark).
				SetNillableFreightFee(req.Data.FreightFee).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(tradeorder.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *OrderRepo) Delete(ctx context.Context, req *gameV1.DeleteOrderRequest) error {
	if req == nil {
		return gameV1.ErrorBadRequest("invalid parameter")
	}

	if err := r.entClient.Client().TradeOrder.DeleteOneID(int64(req.GetId())).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return gameV1.ErrorNotFound("order not found")
		}

		r.log.Errorf("delete order failed: %s", err.Error())
		return gameV1.ErrorInternalServerError("delete failed")
	}

	return nil
}

func enumInt32Ptr[T ~int32](v *T) *int32 {
	if v == nil {
		return nil
	}
	val := int32(*v)
	return &val
}
