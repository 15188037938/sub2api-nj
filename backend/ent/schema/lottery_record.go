package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LotteryRecord holds the schema definition for the LotteryRecord entity.
// 抽奖记录
type LotteryRecord struct {
	ent.Schema
}

func (LotteryRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "lottery_records"},
	}
}

func (LotteryRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Comment("用户ID"),
		field.Int64("prize_id").
			Comment("奖品ID"),
		field.String("prize_name").
			MaxLen(100).
			Comment("奖品名称(冗余，防奖品删除)"),
		field.String("prize_type").
			MaxLen(50).
			Comment("奖品类型(冗余)"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("奖品数值(冗余)"),
		field.Int("cost_points").
			Default(0).
			Comment("消耗积分"),
		field.Bool("claimed").
			Default(false).
			Comment("是否已领取"),
		field.Time("claimed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("领取时间"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (LotteryRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("lottery_records").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (LotteryRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("prize_id"),
		index.Fields("created_at"),
		index.Fields("claimed"),
	}
}
