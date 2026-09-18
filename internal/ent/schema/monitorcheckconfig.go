package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MonitorCheckConfig holds the schema definition for per-check interval configuration.
type MonitorCheckConfig struct {
	ent.Schema
}

// Fields of the MonitorCheckConfig.
func (MonitorCheckConfig) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("monitor_id", uuid.UUID{}),
		field.Enum("check_type").
			Values("http", "dns", "ssl", "domain", "email_auth"),
		field.Int("interval_seconds").
			Positive(),
		field.Int("timeout_seconds").
			Default(10),
		field.Bool("is_enabled").
			Default(true),
		field.Time("last_checked_at").
			Optional().
			Nillable(),
		field.Time("next_check_at").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the MonitorCheckConfig.
func (MonitorCheckConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("check_configs").
			Field("monitor_id").
			Unique().
			Required(),
	}
}

// Indexes of the MonitorCheckConfig.
func (MonitorCheckConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id", "check_type"),
		index.Fields("is_enabled", "next_check_at"),
	}
}
