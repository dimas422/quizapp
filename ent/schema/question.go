package schema

import (
"entgo.io/ent"
"entgo.io/ent/schema/edge"
"entgo.io/ent/schema/field"
"github.com/google/uuid"
)

type Question struct {
ent.Schema
}

func (Question) Fields() []ent.Field {
return []ent.Field{
field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
field.String("text"),
field.Int("order_index").Default(0),
}
}

func (Question) Edges() []ent.Edge {
return []ent.Edge{
edge.From("quiz", Quiz.Type).Ref("questions").Unique(),
edge.To("answers", Answer.Type),
edge.To("attempt_answers", AttemptAnswer.Type),
}
}
