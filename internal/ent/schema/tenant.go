package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
	ent.Schema
}

// Fields of the Tenant.
func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("name").
			NotEmpty(),
		field.String("plan").
			Default("free"), // free, pro, enterprise
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Tenant.
func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", Membership.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("monitors", Monitor.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("alerts", Alert.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("notification_channels", NotificationChannel.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("settings", TenantSetting.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
