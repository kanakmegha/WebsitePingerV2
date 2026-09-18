package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// DomainCheckResult holds the schema definition for the DomainCheckResult entity.
type DomainCheckResult struct {
	ent.Schema
}

// Fields of the DomainCheckResult.
func (DomainCheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("check_id", uuid.UUID{}).
			Unique(),
		field.Time("expiry_date"),
		field.String("registrar"),
		field.Int("days_remaining"),
	}
}

// Edges of the DomainCheckResult.
func (DomainCheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("check", MonitorCheck.Type).
			Ref("domain_result").
			Field("check_id").
			Unique().
			Required(),
	}
}
