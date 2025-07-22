package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/logger"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/router"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Config struct {
	Address    string `env:"ADDRESS"`
	DumpConfig DumpConfig
}

type DumpConfig struct {
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	serverLogger, err := logger.NewLogger(zap.NewProductionConfig())
	if err != nil {
		panic(err)
	}

	var storageCounter = storage.NewMemStorage[model.Counter]()
	var storageGauge = storage.NewMemStorage[model.Gauge]()

	metricService := service.NewMetricService(storageCounter, storageGauge)
	dumperAndRestorer := service.NewFileMetricDumper(cfg.DumpConfig.FileStoragePath)

	if cfg.DumpConfig.Restore {
		err := service.RestoreState(metricService, dumperAndRestorer)
		if err != nil {
			panic(err)
		}
		serverLogger.Info("server state restored", zap.String("FILE_STORAGE_PATH", cfg.DumpConfig.FileStoragePath))
	}

	loggingMiddleware := middleware.NewLoggingMiddleware(serverLogger)
	storeState := middleware.NewStoreStateMiddleware(metricService, dumperAndRestorer, cfg.DumpConfig.StoreInterval)

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:         listener.Addr().String(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler: loggingMiddleware(
			storeState.Middleware(
				router.InitRouter(metricService, storageCounter, storageGauge, serverLogger),
			),
		),
	}

	storeFunc := func() {
		err := service.StoreState(metricService, dumperAndRestorer)
		if err != nil {
			serverLogger.Error("failed to store state", zap.Error(err))
		}
		serverLogger.Info("server state saved on shutdown")
	}
	defer storeFunc()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverLogger.Debug("Ошибка при работе сервера: %v", zap.Error(err))
		}
	}()

	select {
	case <-quit:
		serverLogger.Debug("Получен сигнал завершения работы")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			serverLogger.Debug("Ошибка при завершении работы сервера", zap.Error(err))
		}
		serverLogger.Debug("Сервер остановлен")
	}
}

func loadConfig() (*Config, error) {
	address := flag.String("address", "localhost:8080", "HTTP server address")
	restore := flag.Bool("restore", false, "Restore metrics before starting")
	storeInterval := flag.Uint64("store-interval", 300, "Store interval in seconds")
	fileStoragePath := flag.String("file-storage-path", "metric-storage.json", "File storage path")

	flag.StringVar(address, "a", *address, "HTTP server address (short)")
	flag.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flag.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flag.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")

	flag.Parse()

	cfg := &Config{
		Address: *address,
		DumpConfig: DumpConfig{
			Restore:         *restore,
			StoreInterval:   *storeInterval,
			FileStoragePath: *fileStoragePath,
		},
	}

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restoreVal, err := strconv.ParseBool(envRestore)
		if err != nil {
			log.Printf("Ошибка парсинга RESTORE: %v", err)
		} else {
			cfg.DumpConfig.Restore = restoreVal
		}
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		interval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil {
			log.Printf("Ошибка парсинга STORE_INTERVAL: %v", err)
		} else {
			cfg.DumpConfig.StoreInterval = interval
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.DumpConfig.FileStoragePath = envFileStoragePath
	}

	fmt.Printf("Конфигурация: %+v\n", *cfg)
	return cfg, nil
}
