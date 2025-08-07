package config

import (
	"database/sql"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
	"os"
)

func MetricDumperFactory() container.Factory[*service.MetricDumper] {
	return func(c container.Container) (*service.MetricDumper, error) {
		cfg, err := container.GetService[Config](c, "config")
		if err != nil {
			return nil, err
		}

		var dumper service.MetricDumper
		if cfg.DatabaseDsn == "" {
			dumper = service.NewFileMetricDumper(cfg.DumpConfig.FileStoragePath)
		} else {
			db, err := container.GetService[sql.DB](c, "db")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			logger, err := container.GetService[zap.Logger](c, "logger")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dumper = service.NewDBMetricDumper(db, logger)
		}

		return &dumper, nil
	}
}

func MetricRestorerFactory() container.Factory[*service.MetricRestorer] {
	return func(c container.Container) (*service.MetricRestorer, error) {
		cfg, err := container.GetService[Config](c, "config")
		if err != nil {
			return nil, err
		}

		var restorer service.MetricRestorer
		if cfg.DatabaseDsn == "" {
			restorer = service.NewFileMetricRestorer(cfg.DumpConfig.FileStoragePath)
		} else {
			db, err := container.GetService[sql.DB](c, "db")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			restorer = service.NewDBMetricRestorer(db)
		}

		return &restorer, nil
	}
}
