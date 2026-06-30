package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CheckInRecord holds the schema definition for the CheckInRecord entity.
// 用户签到记录
type CheckInRecord struct {
	ent.Schema
}

func (CheckInRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "check_in_records"},
	}
}

func (CheckInRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (CheckInRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Comment("用户ID"),
		field.Time("check_in_date").
			SchemaType(map[string]string{dialect.Postgres: "date"}).
			Comment("签到日期"),
		field.Int("points_earned").
			Default(0).
			Comment("本次签到获得积分"),
		field.Int("consecutive_days").
			Default(1).
			Comment("连续签到天数"),
		field.Int("total_points").
			Default(0).
			Comment("当前总积分"),
	}
}

func (CheckInRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("check_in_records").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (CheckInRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 每个用户每天只能签到一次
		index.Fields("user_id", "check_in_date").
			Unique(),
		index.Fields("user_id"),
		index.Fields("check_in_date"),
	}
}
