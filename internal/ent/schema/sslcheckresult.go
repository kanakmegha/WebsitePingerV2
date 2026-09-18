package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// SSLCheckResult holds the schema definition for the SSLCheckResult entity.
type SSLCheckResult struct {
	ent.Schema
}

// Fields of the SSLCheckResult.
func (SSLCheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("check_id", uuid.UUID{}).
			Unique(),
		field.Time("expiry_date"),
		field.String("issuer"),
		field.Bool("valid"),
		field.Int("days_remaining"),
	}
}

// Edges of the SSLCheckResult.
func (SSLCheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("check", MonitorCheck.Type).
			Ref("ssl_result").
			Field("check_id").
			Unique().
			Required(),
	}
}
