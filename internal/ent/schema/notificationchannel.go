package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// NotificationChannel holds the schema definition for the NotificationChannel entity.
type NotificationChannel struct {
	ent.Schema
}

// Fields of the NotificationChannel.
func (NotificationChannel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}),
		field.Enum("type").
			Values("email", "webhook", "slack", "pagerduty"),
		field.JSON("config", map[string]interface{}{}),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the NotificationChannel.
func (NotificationChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("notification_channels").
			Field("tenant_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
