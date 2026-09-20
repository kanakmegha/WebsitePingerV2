package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
		field.UUID("alert_id", uuid.UUID{}),
		field.UUID("monitor_id", uuid.UUID{}),
		field.Enum("status").
			Values("triggered", "resolved"),
		field.String("message"),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the AlertEvent.
func (AlertEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert", Alert.Type).
			Ref("events").
			Field("alert_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("monitor", Monitor.Type).
			Ref("alert_events").
			Field("monitor_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
