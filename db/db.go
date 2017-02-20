package db

import "github.com/jackc/pgx"

var PgPool *pgx.ConnPool

func init() {
	PgPool, _ = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "db",
			User:     "postgres",
			Password: "postgres",
			Database: "mg4",
		},
	})
}
