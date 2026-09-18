package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TenantSetting holds the schema definition for tenant-level default monitoring intervals.
type TenantSetting struct {
	ent.Schema
}

// Fields of the TenantSetting.
func (TenantSetting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}).
			Unique(),
		field.Int("http_interval_seconds").
			Default(60),
		field.Int("dns_interval_seconds").
			Default(300),
		field.Int("ssl_interval_seconds").
			Default(3600),
		field.Int("domain_interval_seconds").
			Default(86400),
		field.Int("email_auth_interval_seconds").
			Default(300),
		field.Time("updated_at").
			Default(time.Now),
	}
}

// Edges of the TenantSetting.
func (TenantSetting) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("settings").
			Field("tenant_id").
			Unique().
			Required(),
	}
}

// Indexes of the TenantSetting.
func (TenantSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id").Unique(),
	}
}
