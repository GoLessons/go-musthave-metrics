package main

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("try staring server")

	c := config.InitContainer()

	serverLogger, err := container.GetService[zap.Logger](c, "logger")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := container.GetService[config.Config](c, "config")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	serverLogger.Info("server config", zap.Any("cfg", cfg))

	storageCounter, err := container.GetService[storage.MemStorage[model.Counter]](c, "counterStorage")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	storageGauge, err := container.GetService[storage.MemStorage[model.Gauge]](c, "gaugeStorage")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

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

	r, err := container.GetService[chi.Mux](c, "router")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	serverLogger.Info("server listening", zap.String("address", listener.Addr().String()))

	server := &http.Server{
		Addr:         listener.Addr().String(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler: loggingMiddleware(
			storeState.Middleware(r),
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverLogger.Debug("Ошибка при работе сервера: %v", zap.Error(err))
		}
	}()

	<-quit
	serverLogger.Debug("Получен сигнал завершения работы")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		serverLogger.Debug("Ошибка при завершении работы сервера", zap.Error(err))
	}
	serverLogger.Debug("Сервер остановлен")
}
