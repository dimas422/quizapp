package ent

import (
"entgo.io/ent/dialect"
entsql "entgo.io/ent/dialect/sql"
"encore.dev/storage/sqldb"
)

var quizDB = sqldb.NewDatabase("quiz", sqldb.DatabaseConfig{
Migrations: "../auth/migrations",
})

func OpenEntClient() (*Client, error) {
db := quizDB.Stdlib()
drv := entsql.OpenDB(dialect.Postgres, db)
return NewClient(Driver(drv)), nil
}
