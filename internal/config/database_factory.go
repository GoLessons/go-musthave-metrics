package config

import (
	"database/sql"

	config2 "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func DBFactory() container.Factory[*sql.DB] {
	return func(c container.Container) (*sql.DB, error) {
		conf, err := container.GetService[config2.Config](c, "config")
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
