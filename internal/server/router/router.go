package router

import (
	"database/sql"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func RouterFactory() container.Factory[*chi.Mux] {
	return func(c container.Container) (*chi.Mux, error) {
		logger, err := container.GetService[zap.Logger](c, "logger")
		if err != nil {
			return nil, err
		}

		counterStorage, err := container.GetService[storage.MemStorage[serverModel.Counter]](c, "counterStorage")
		if err != nil {
			return nil, err
		}

		gaugeStorage, err := container.GetService[storage.MemStorage[serverModel.Gauge]](c, "gaugeStorage")
		if err != nil {
			return nil, err
		}

		metricService, err := container.GetService[service.MetricService](c, "metricService")
		if err != nil {
			return nil, err
		}

		db, err := container.GetService[sql.DB](c, "db")
		if err != nil {
			return nil, err
		}

		r := chi.NewRouter()

		metricControllerJSON := handler.NewMetricsController(*metricService, handler.JSONResposeBuilder, logger)
		metricControllerPlain := handler.NewMetricsController(*metricService, handler.PlainResposeBuilder, logger)

		r.Route("/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:(-?)[a-z0-9\\.]+}",
			func(r chi.Router) {
				r.Use(middleware.MetricCtxFromPath)
				r.Post("/", metricControllerPlain.Update)
			},
		)

		r.Route("/value/{metricType}/{metricName:[a-zA-Z0-9_-]+}", func(r chi.Router) {
			r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricCtxFromPath)
			r.Get("/", metricControllerPlain.Get)
		})

		r.Route("/update", func(r chi.Router) {
			r.Use(middleware.ValidateRoute)
			r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricCtxFromBody)
			r.Post("/", metricControllerJSON.Update)
			r.Post("/.+", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Not Found", http.StatusNotFound)
			})
		})

		r.Route("/value",
			func(r chi.Router) {
				r.Use(middleware.GzipMiddleware)
				r.Use(middleware.MetricCtxFromBody)
				r.Post("/", metricControllerJSON.Get)
			},
		)

		r.Route("/ping",
			func(r chi.Router) {
				r.Get("/", handler.NewPingHandler(db, logger).Ping)
			},
		)

		r.Route("/",
			func(r chi.Router) {
				r.Use(middleware.GzipMiddleware)
				r.Get(
					"/",
					handler.NewListController(
						counterStorage,
						gaugeStorage,
					).Get,
				)
			},
		)

		return r, nil
	}
}
