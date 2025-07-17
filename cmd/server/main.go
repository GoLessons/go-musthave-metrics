package main

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/caarlos0/env"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
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

	loggingMiddleware := middleware.NewLoggingMiddleware(serverLogger)
	return http.ListenAndServe(cfg.Address, loggingMiddleware(router.InitRouter(storageCounter, storageGauge)))
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&cfg.Address, "address", "a", cfg.Address, "HTTP server address")

	return cfg, nil
}
