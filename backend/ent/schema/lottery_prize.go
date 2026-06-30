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

// LotteryPrize holds the schema definition for the LotteryPrize entity.
// 奖品配置
type LotteryPrize struct {
	ent.Schema
}

func (LotteryPrize) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "lottery_prizes"},
	}
}

func (LotteryPrize) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (LotteryPrize) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Comment("奖品名称"),
		field.String("prize_type").
			MaxLen(50).
			NotEmpty().
			Comment("奖品类型: balance(余额), concurrency_boost(并发加成), points(积分返还)"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("奖品数值(余额金额/并发数/积分数)"),
		field.Int("weight").
			Default(10).
			Comment("概率权重，越大越容易中奖"),
		field.Int("total_stock").
			Default(-1).
			Comment("总库存，-1表示无限"),
		field.Int("remaining_stock").
			Default(-1).
			Comment("剩余库存"),
		field.String("icon").
			MaxLen(255).
			Default("").
			Comment("奖品图标(emoji或图标名)"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("状态: active, disabled"),
		field.Int("sort_order").
			Default(0).
			Comment("排序"),
	}
}

func (LotteryPrize) Edges() []ent.Edge {
	return nil
}

func (LotteryPrize) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("sort_order"),
	}
}
