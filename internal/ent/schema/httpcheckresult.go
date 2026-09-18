package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// HTTPCheckResult holds the schema definition for the HTTPCheckResult entity.
type HTTPCheckResult struct {
	ent.Schema
}

// Fields of the HTTPCheckResult.
func (HTTPCheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("check_id", uuid.UUID{}).
			Unique(),
		field.Int("status_code"),
		field.Int64("response_time_ms"),
		field.Int64("response_size_bytes"),
		field.String("final_url"),
	}
}

// Edges of the HTTPCheckResult.
func (HTTPCheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("check", MonitorCheck.Type).
			Ref("http_result").
			Field("check_id").
			Unique().
			Required(),
	}
}
