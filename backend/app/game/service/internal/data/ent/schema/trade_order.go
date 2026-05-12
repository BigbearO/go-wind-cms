package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// TradeOrder holds the schema definition for the TradeOrder entity.
type TradeOrder struct {
	ent.Schema
}

func (TradeOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "trade_order",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("订单表"),
	}
}

// Fields of the TradeOrder.
func (TradeOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("订单ID").
			Immutable(),

		field.String("order_no").
			Comment("三方订单流水号").
			NotEmpty(),

		field.String("user_id").
			Comment("用户ID").
			NotEmpty(),

		field.Int32("order_type").
			Comment("订单类型").
			Default(0),

		field.Int32("status").
			Comment("订单状态：1-待支付 2-已支付 3-已取消 4-已退款").
			Default(1),

		field.Int32("game_type").
			Comment("游戏类型").
			Default(0),

		field.Int32("days").
			Comment("购买天数").
			Default(0),

		field.String("pay_amount").
			Comment("支付金额，单位分").
			Default("0"),

		field.Int32("pay_type").
			Comment("支付渠道").
			Optional().
			Nillable(),

		field.Time("pay_time").
			Comment("支付时间").
			Optional().
			Nillable(),

		field.Time("expire_time").
			Comment("订单过期时间").
			Optional().
			Nillable(),

		field.String("cancel_reason").
			Comment("取消理由").
			Optional().
			Nillable(),

		field.Time("cancel_time").
			Comment("取消时间").
			Optional().
			Nillable(),

		field.Int32("refund_status").
			Comment("退款状态").
			Optional().
			Nillable(),

		field.Time("refund_time").
			Comment("退款时间").
			Optional().
			Nillable(),

		field.String("refund_channel").
			Comment("退款渠道").
			Optional().
			Nillable(),

		field.String("freight_fee").
			Comment("运费").
			Optional().
			Nillable().
			Default("0"),
	}
}

// Mixin of the TradeOrder.
func (TradeOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
		mixin.Remark{},
	}
}

// Indexes of the TradeOrder.
func (TradeOrder) Indexes() []ent.Index {
	return []ent.Index{
		// 订单编号唯一
		index.Fields("order_no").
			Unique().
			StorageKey("uix_trade_order_order_no"),

		// 按用户ID查询订单
		index.Fields("user_id").
			StorageKey("idx_trade_order_user_id"),

		// 按订单状态查询
		index.Fields("status").
			StorageKey("idx_trade_order_status"),

		// 按游戏类型查询
		index.Fields("game_type").
			StorageKey("idx_trade_order_game_type"),

		// 支付时间范围查询
		index.Fields("pay_time").
			StorageKey("idx_trade_order_pay_time"),

		// 过期时间范围查询
		index.Fields("expire_time").
			StorageKey("idx_trade_order_expire_time"),

		// 租户 + 创建时间（用于分页）
		index.Fields("tenant_id", "created_at").
			StorageKey("idx_trade_order_tenant_created_at"),

		// 创建时间索引
		index.Fields("created_at").
			StorageKey("idx_trade_order_created_at"),
	}
}
