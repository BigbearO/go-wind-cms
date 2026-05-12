package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/tx7do/go-crud/entgo/mixin"
)

// GameAccount holds the schema definition for the GameAccount entity.
type GameAccount struct {
	ent.Schema
}

func (GameAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "game_account",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_bin",
		},
		entsql.WithComments(true),
		schema.Comment("游戏账号表"),
	}
}

// Fields of the GameAccount.
func (GameAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("账号ID").
			Immutable(),

		field.String("user_account").
			Comment("用户账号").
			NotEmpty(),

		field.String("credential").
			Comment("用户登录凭证").
			Optional().
			Nillable(),

		field.Time("start_time").
			Comment("会员开始时间").
			Optional().
			Nillable(),

		field.Time("expire_time").
			Comment("会员过期时间").
			Optional().
			Nillable(),

		field.Int32("type").
			Comment("游戏类型").
			Default(0),

		field.Bool("enabled").
			Comment("是否启用").
			Default(true),

		field.Int64("emulator_id").
			Comment("分配的模拟器ID").
			Optional().
			Nillable(),
	}
}

// Mixin of the GameAccount.
func (GameAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeAt{},
		mixin.OperatorID{},
		mixin.TenantID[uint32]{},
	}
}

// Indexes of the GameAccount.
func (GameAccount) Indexes() []ent.Index {
	return []ent.Index{
		// 用户账号 + 游戏类型 唯一
		index.Fields("user_account", "type").
			Unique().
			StorageKey("uix_game_account_user_type"),

		// 按用户账号查询
		index.Fields("user_account").
			StorageKey("idx_game_account_user_account"),

		// 按游戏类型查询
		index.Fields("type").
			StorageKey("idx_game_account_type"),

		// 按启用状态查询
		index.Fields("enabled").
			StorageKey("idx_game_account_enabled"),

		// 过期时间范围查询
		index.Fields("expire_time").
			StorageKey("idx_game_account_expire_time"),

		// 模拟器ID查询
		index.Fields("emulator_id").
			StorageKey("idx_game_account_emulator_id"),

		// 租户 + 创建时间（用于分页）
		index.Fields("tenant_id", "created_at").
			StorageKey("idx_game_account_tenant_created_at"),

		// 创建时间索引
		index.Fields("created_at").
			StorageKey("idx_game_account_created_at"),
	}
}
