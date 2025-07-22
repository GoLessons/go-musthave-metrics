package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type Config struct {
	Address    string `env:"ADDRESS" envDefault:"localhost:8080"`
	DumpConfig *DumpConfig
}

type DumpConfig struct {
	StoreInterval   uint64 `env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"metric-storage.json"`
	Restore         bool   `env:"RESTORE" envDefault:"false"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "server",
	}

	cfg, err := loadConfig(rootCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		return run(cfg)
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cfg *Config) error {
	serverLogger, err := logger.NewLogger(zap.NewProductionConfig())
	if err != nil {
		return err
	}

	var storageCounter = storage.NewMemStorage[model.Counter]()
	var storageGauge = storage.NewMemStorage[model.Gauge]()

	metricService := service.NewMetricService(storageCounter, storageGauge)
	metricDumper := service.NewFileMetricDumper(cfg.DumpConfig.FileStoragePath)

	if cfg.DumpConfig.Restore {
		err := service.RestoreState(metricService, metricDumper)
		if err != nil {
			return err
		}
		serverLogger.Info("Server state restored")
	}

	loggingMiddleware := middleware.NewLoggingMiddleware(serverLogger)
	storeState := middleware.NewStoreStateMiddleware(metricService, metricDumper, cfg.DumpConfig.StoreInterval)
	server := &http.Server{
		Addr: cfg.Address,
		Handler: loggingMiddleware(
			storeState.Middleware(
				router.InitRouter(metricService, storageCounter, storageGauge),
			),
		),
	}

	server.RegisterOnShutdown(func() {
		err := service.StoreState(metricService, metricDumper)
		if err != nil {
			serverLogger.Error("failed to store state", zap.Error(err))
		}
		serverLogger.Info("Server state saved on shutdown")
	})

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	dumpConfig := &DumpConfig{}
	err := env.Parse(dumpConfig)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		DumpConfig: dumpConfig,
	}
	err = env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", cfg.Address, "HTTP server address")
	cmd.Flags().BoolVarP(&cfg.DumpConfig.Restore, "restore", "r", cfg.DumpConfig.Restore, "Restore metrics before starting")
	cmd.Flags().Uint64VarP(&cfg.DumpConfig.StoreInterval, "store-interval", "i", cfg.DumpConfig.StoreInterval, "Store interval in seconds")
	cmd.Flags().StringVarP(&cfg.DumpConfig.FileStoragePath, "file-storage-path", "f", cfg.DumpConfig.FileStoragePath, "File storage path")

	return cfg, nil
}
