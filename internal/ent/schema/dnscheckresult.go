package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// DNSCheckResult holds the schema definition for the DNSCheckResult entity.
type DNSCheckResult struct {
	ent.Schema
}

// Fields of the DNSCheckResult.
func (DNSCheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("check_id", uuid.UUID{}).
			Unique(),
		field.Bool("has_a_record"),
		field.Bool("has_aaaa_record"),
		field.JSON("a_records", []string{}).
			Optional(),
		field.JSON("aaaa_records", []string{}).
			Optional(),
		field.JSON("mx_records", []string{}).
			Optional(),
		field.JSON("txt_records", []string{}).
			Optional(),
		field.Bool("spf_valid"),
		field.Bool("dmarc_valid"),
	}
}

// Edges of the DNSCheckResult.
func (DNSCheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("check", MonitorCheck.Type).
			Ref("dns_result").
			Field("check_id").
			Unique().
			Required(),
	}
}
