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

// Monitor holds the schema definition for the Monitor entity.
type Monitor struct {
	ent.Schema
}

// Fields of the Monitor.
func (Monitor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}),
		field.String("name").
			NotEmpty(),
		field.String("url").
			NotEmpty(),
		field.String("domain").
			NotEmpty(),
		field.Enum("type").
			Values("http", "ssl", "dns", "whois").
			Default("http"),
		field.JSON("enabled_checks", map[string]bool{}).
			Optional(),
		field.Int("interval_seconds").
			Default(60),
		field.Int("timeout_seconds").
			Default(10),
		field.Bool("is_active").
			Default(true),
		field.Enum("last_status").
			Values("up", "down").
			Optional().
			Nillable(),
		field.Time("last_checked_at").
			Optional().
			Nillable(),
		field.Time("last_alert_sent_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Monitor.
func (Monitor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("monitors").
			Field("tenant_id").
			Unique().
			Required(),
		edge.To("checks", MonitorCheck.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("check_configs", MonitorCheckConfig.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("alerts", Alert.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("alert_events", AlertEvent.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Monitor.
func (Monitor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("domain"),
		index.Fields("tenant_id", "is_active"),
	}
}
