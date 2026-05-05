package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type AttemptAnswer struct {
	ent.Schema
}

func (AttemptAnswer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "attempt_answers"},
	}
}

func (AttemptAnswer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
	}
}

func (AttemptAnswer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attempt", Attempt.Type).Ref("attempt_answers").Unique(),
		edge.From("question", Question.Type).Ref("attempt_answers").Unique(),
		edge.From("answer", Answer.Type).Ref("attempt_answers").Unique(),
	}
}
