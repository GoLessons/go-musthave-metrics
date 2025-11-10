package container

import (
	"database/sql"

	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	config2 "github.com/GoLessons/go-musthave-metrics/internal/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func InitContainer() (container.Container, error) {
	cfg, err := config.LoadConfig(nil)
	if err != nil {
		return nil, err
	}

	serverLogger, err := logger.NewLogger(zap.NewProductionConfig())
	if err != nil {
		return nil, err
	}

	storageCounter := storage.NewMemStorage[model.Counter]()
	storageGauge := storage.NewMemStorage[model.Gauge]()
	metricService := service.NewMetricService(storageCounter, storageGauge)

	services := map[string]any{
		"logger":         serverLogger,
		"config":         cfg,
		"counterStorage": storageCounter,
		"gaugeStorage":   storageGauge,
		"metricService":  metricService,
	}

	if cfg.DatabaseDsn != "" {
		sqlDB, err := sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			return nil, err
		}
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetMaxIdleConns(10)
		if err := sqlDB.Ping(); err != nil {
			return nil, err
		}
		services["db"] = sqlDB
	}

	c := container.NewSimpleContainer(services)

	container.SimpleRegisterFactory(&c, "router", router.RouterFactory())
	container.SimpleRegisterFactory(&c, "dumper", config2.MetricDumperFactory())
	container.SimpleRegisterFactory(&c, "restorer", config2.MetricRestorerFactory())

	return c, nil
}
