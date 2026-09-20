package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Alert holds the schema definition for the Alert entity.
type Alert struct {
	ent.Schema
}

// Fields of the Alert.
func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}),
		field.UUID("monitor_id", uuid.UUID{}),
		field.Enum("type").
			Values("down", "ssl_expiring", "domain_expiring", "dns_changed"),
		field.Int("threshold").
			Default(1),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("alerts").
			Field("tenant_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("monitor", Monitor.Type).
			Ref("alerts").
			Field("monitor_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("events", AlertEvent.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
