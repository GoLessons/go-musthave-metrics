package config

import (
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"go.uber.org/zap"
	"os"
)

func InitContainer() container.Container {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	serverLogger, err := logger.NewLogger(zap.NewProductionConfig())
	if err != nil {
		panic(err)
	}

	c := container.NewSimpleContainer(map[string]any{
		"logger": serverLogger,
		"config": cfg,
	})

	return c
}
