package ent

import (
"context"
"fmt"

"entgo.io/ent/dialect"
entsql "entgo.io/ent/dialect/sql"
"encore.dev/storage/sqldb"
)

var quizDB = sqldb.Named("quiz")

func OpenEntClient() (*Client, error) {
db := quizDB.Stdlib()
drv := entsql.OpenDB(dialect.Postgres, db)
return NewClient(Driver(drv)), nil
}

func GetEntClient(ctx context.Context) (*Client, error) {
client, err := OpenEntClient()
if err != nil {
return nil, fmt.Errorf("failed to open ent client: %w", err)
}
return client, nil
}
