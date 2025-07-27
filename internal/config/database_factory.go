package config

import (
	"database/sql"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func DbFactory() container.Factory[*sql.DB] {
	return func(c container.Container) (*sql.DB, error) {
		conf, err := container.GetService[Config](c, "config")
		if err != nil {
			return nil, err
		}

		db, err := sql.Open("pgx", conf.DatabaseDsn)
		if err != nil {
			return nil, err
		}

		return db, nil
	}
}
