package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	httppprof "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	apiModel "github.com/GoLessons/go-musthave-metrics/internal/model"
	config2 "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	container2 "github.com/GoLessons/go-musthave-metrics/internal/server/container"
	database "github.com/GoLessons/go-musthave-metrics/internal/server/db"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/GoLessons/go-musthave-metrics/pkg/repeater"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func preWarmDecoders() {
	// Прогреваем декодер для одиночного объекта метрики
	var m apiModel.Metrics
	_ = json.Unmarshal([]byte("{}"), &m)

	// Прогреваем декодер для списка метрик (batch обновления)
	var ml []apiModel.Metrics
	_ = json.Unmarshal([]byte("[]"), &ml)
}

func main() {
	c := container2.InitContainer()
	mainCtx := context.Background()

	serverLogger, err := container.GetService[zap.Logger](c, "logger")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := container.GetService[config2.Config](c, "config")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.PprofHTTP {
		addr := cfg.PprofHTTPAddr
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", httppprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", httppprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", httppprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", httppprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", httppprof.Trace)
		go func() {
			serverLogger.Info("pprof HTTP enabled", zap.String("address", addr))
			if err := http.ListenAndServe(addr, mux); err != nil {
				serverLogger.Error("pprof HTTP server error", zap.Error(err))
			}
		}()
	}

	serverLogger.Info("Server config", zap.Any("cfg", cfg))

	var db *sql.DB
	if cfg.DatabaseDsn != "" {
		db, err = container.GetService[sql.DB](c, "db")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}(db)

		tryMigrateDB(cfg, db, serverLogger)
	}

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

	restorer, err := container.GetService[service.MetricRestorer](c, "restorer")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if cfg.DumpConfig.Restore {
		try := repeater.NewRepeater(func(err error) {
			serverLogger.Info("Неудачная попытка восстановить состояние", zap.Error(err))
		})
		repeatStrategy := repeater.NewFixedDelaysStrategy(
			database.NewPostgresErrorClassifier().IsRetriable,
			time.Second*1,
			time.Second*3,
			time.Second*5,
		)
		_, err := try.Repeat(
			repeatStrategy,
			func() (any, error) {
				err := service.RestoreState(metricService, *restorer)
				return nil, err
			},
		)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		serverLogger.Info("server state restored", zap.String("FILE_STORAGE_PATH", cfg.DumpConfig.FileStoragePath))
	}

	dumper, err := container.GetService[service.MetricDumper](c, "dumper")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

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
		Handler:      r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverLogger.Debug("Ошибка при работе сервера: %v", zap.Error(err))
		}
	}()

	go iterateFunc(mainCtx, cfg.DumpConfig.StoreInterval, func() {
		storeMetrics(serverLogger, metricService, dumper)
	})

	<-quit
	serverLogger.Debug("Получен сигнал завершения работы")
	ctx, cancel := context.WithTimeout(mainCtx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		serverLogger.Debug("Ошибка при завершении работы сервера", zap.Error(err))
	}

	// Запись heap-профиля при завершении (если включено)
	if cfg.PprofOnShutdown {
		if err := os.MkdirAll(cfg.PprofDir, 0755); err != nil {
			serverLogger.Error("Ошибка создания директории профилей", zap.Error(err))
		} else {
			p := filepath.Join(cfg.PprofDir, cfg.PprofFilename)
			f, err := os.Create(p)
			if err != nil {
				serverLogger.Error("Ошибка создания файла профиля", zap.Error(err))
			} else {
				runtime.GC()
				if err := pprof.WriteHeapProfile(f); err != nil {
					serverLogger.Error("Ошибка записи heap-профиля", zap.Error(err))
				} else {
					serverLogger.Info("Heap-профиль сохранён", zap.String("path", p))
				}
				_ = f.Close()
			}
		}
	}

	serverLogger.Debug("Сервер остановлен")
}

func storeMetrics(serverLogger *zap.Logger, metricService *service.MetricService, dumper *service.MetricDumper) {
	try := repeater.NewRepeater(func(err error) {
		serverLogger.Error("Ошибка сохранения состояния", zap.Error(err))
	})
	repeatStrategy := repeater.NewFixedDelaysStrategy(
		database.NewPostgresErrorClassifier().IsRetriable,
		time.Second*1,
		time.Second*3,
		time.Second*5,
	)
	_, err := try.Repeat(
		repeatStrategy,
		func() (any, error) {
			err := service.StoreState(metricService, *dumper)
			return nil, err
		},
	)
	if err != nil {
		serverLogger.Error("Ошибка сохранения состояния после повторов", zap.Error(err))
		return
	}
	serverLogger.Info("Состояние сервера сохранено")
}

func tryMigrateDB(cfg *config2.Config, db *sql.DB, serverLogger *zap.Logger) {
	if cfg.DatabaseDsn != "" {
		migrator := database.NewMigrator(db, serverLogger)
		err := migrator.Up()
		if err != nil {
			fmt.Printf("Magrations error: %v\n", err)
			os.Exit(1)
		}
	}
}

func iterateFunc(ctx context.Context, interval uint64, callable func()) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			callable()
			return
		case <-ticker.C:
			callable()
		}
	}
}
