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

// AlertEvent holds the schema definition for the AlertEvent entity.
type AlertEvent struct {
	ent.Schema
}

// Fields of the AlertEvent.
func (AlertEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}),
		field.UUID("monitor_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("alert_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("type").
			Values("incident", "recovery", "invite_sent", "invite_accepted").
			Default("incident"),
		field.String("message"),
		field.Enum("status").
			Values("unread", "read").
			Default("unread"),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the AlertEvent.
func (AlertEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("alert_events").
			Field("tenant_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("monitor", Monitor.Type).
			Ref("alert_events").
			Field("monitor_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("alert", Alert.Type).
			Ref("events").
			Field("alert_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the AlertEvent.
func (AlertEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("tenant_id", "created_at"),
	}
}
