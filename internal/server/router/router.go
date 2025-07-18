package router

import (
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
)

func InitRouter(storageCounter storage.Storage[serverModel.Counter], storageGauge storage.Storage[serverModel.Gauge]) *chi.Mux {
	r := chi.NewRouter()

	metricService := service.NewMetricService(storageCounter, storageGauge)
	metricController := handler.NewMetricsController(*metricService)

	r.Route("/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:(-?)[a-z0-9\\.]+}",
		func(r chi.Router) {
			r.Use(middleware.MetricCtxFromPath)
			r.Post("/", metricController.Update)
		},
	)
	r.Route("/value/{metricType}/{metricName:[a-zA-Z0-9_-]+}", func(r chi.Router) {
		r.Use(middleware.MetricCtxFromPath)
		r.Get("/", metricController.Get)
	})
	r.Route("/update",
		func(r chi.Router) {
			r.Use(middleware.MetricCtxFromBody)
			r.Post("/", metricController.Update)
		},
	)
	r.Route("/value",
		func(r chi.Router) {
			r.Use(middleware.MetricCtxFromBody)
			r.Post("/", metricController.Get)
		},
	)

	r.Get(
		"/",
		handler.NewListController(
			storageCounter,
			storageGauge,
		).Get,
	)

	return r
}
