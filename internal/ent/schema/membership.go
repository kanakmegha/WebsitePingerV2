package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Membership holds the schema definition for the Membership entity.
type Membership struct {
	ent.Schema
}

// Fields of the Membership.
func (Membership) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("tenant_id", uuid.UUID{}),
		field.Enum("role").
			Values("owner", "admin", "member").
			Default("member"),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Membership.
func (Membership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("memberships").
			Field("user_id").
			Unique().
			Required(),
		edge.From("tenant", Tenant.Type).
			Ref("memberships").
			Field("tenant_id").
			Unique().
			Required(),
	}
}

// Indexes of the Membership.
func (Membership) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint ensuring a user belongs to a tenant only once
		index.Fields("user_id", "tenant_id").Unique(),
	}
}
