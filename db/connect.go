package db

import (
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

func ConnectClickhouse(dsn string) (*sqlx.DB, error) {
	c, err := sqlx.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	return c, nil
}
