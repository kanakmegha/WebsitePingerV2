package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PushSubscription holds the schema definition for the PushSubscription entity.
type PushSubscription struct {
	ent.Schema
}

// Fields of the PushSubscription.
func (PushSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("tenant_id", uuid.UUID{}),
		field.String("endpoint").
			NotEmpty(),
		field.String("p256dh").
			NotEmpty(),
		field.String("auth").
			NotEmpty(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the PushSubscription.
func (PushSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("push_subscriptions").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("tenant", Tenant.Type).
			Ref("push_subscriptions").
			Field("tenant_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the PushSubscription.
func (PushSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "endpoint").
			Unique(),
		index.Fields("tenant_id"),
	}
}
