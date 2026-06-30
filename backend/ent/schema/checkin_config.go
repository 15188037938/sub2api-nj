package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CheckInConfig holds the schema definition for the CheckInConfig entity.
// 签到配置（全局单例或按分组配置）
type CheckInConfig struct {
	ent.Schema
}

func (CheckInConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "check_in_configs"},
	}
}

func (CheckInConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (CheckInConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int("daily_min_points").
			Default(1).
			Comment("每日签到最低积分"),
		field.Int("daily_max_points").
			Default(10).
			Comment("每日签到最高积分"),
		field.Int("lottery_cost").
			Default(10).
			Comment("单次抽奖消耗积分"),
		field.Int("daily_max_draws").
			Default(5).
			Comment("每人每天最多抽奖次数"),
		field.String("consecutive_bonus_json").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default("[]").
			Comment("连续签到加成配置JSON: [{days:3, bonus:2}, {days:7, bonus:5}, {days:30, bonus:20}]"),
		field.Bool("enabled").
			Default(true).
			Comment("是否启用签到抽奖功能"),
	}
}

func (CheckInConfig) Edges() []ent.Edge {
	return nil
}

func (CheckInConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
