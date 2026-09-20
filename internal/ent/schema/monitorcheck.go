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

// MonitorCheck holds the schema definition for the MonitorCheck entity.
type MonitorCheck struct {
	ent.Schema
}

// Fields of the MonitorCheck.
func (MonitorCheck) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("monitor_id", uuid.UUID{}),
		field.Enum("check_type").
			Values("http", "ssl", "dns", "whois"),
		field.Enum("status").
			Values("success", "failure", "rate_limited"),
		field.String("error").
			Optional().
			Nillable(),
		field.Time("checked_at").
			Default(time.Now),
	}
}

// Edges of the MonitorCheck.
func (MonitorCheck) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("monitor", Monitor.Type).
			Ref("checks").
			Field("monitor_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("http_result", HTTPCheckResult.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("ssl_result", SSLCheckResult.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("domain_result", DomainCheckResult.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("dns_result", DNSCheckResult.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the MonitorCheck.
func (MonitorCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monitor_id", "checked_at"),
		index.Fields("check_type", "checked_at"),
	}
}
