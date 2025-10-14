package container

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	config2 "github.com/GoLessons/go-musthave-metrics/internal/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
	"os"
)

func InitContainer() container.Container {
	cfg, err := config.LoadConfig(nil)
	if err != nil {
		fmt.Printf("DI Error: %v\n", err)
		os.Exit(1)
	}

	serverLogger, err := logger.NewLogger(zap.NewProductionConfig())
	if err != nil {
		fmt.Printf("DI Error: %v\n", err)
		os.Exit(1)
	}

	storageCounter := storage.NewMemStorage[model.Counter]()
	storageGauge := storage.NewMemStorage[model.Gauge]()
	metricService := service.NewMetricService(storageCounter, storageGauge)

	c := container.NewSimpleContainer(
		map[string]any{
			"logger":         serverLogger,
			"config":         cfg,
			"counterStorage": storageCounter,
			"gaugeStorage":   storageGauge,
			"metricService":  metricService,
		},
	)

	container.SimpleRegisterFactory(&c, "db", config2.DBFactory())
	container.SimpleRegisterFactory(&c, "router", router.RouterFactory())
	container.SimpleRegisterFactory(&c, "dumper", config2.MetricDumperFactory())
	container.SimpleRegisterFactory(&c, "restorer", config2.MetricRestorerFactory())

	return c
}
