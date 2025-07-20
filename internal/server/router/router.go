package router

import (
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/middleware"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func InitRouter(storageCounter storage.Storage[serverModel.Counter], storageGauge storage.Storage[serverModel.Gauge]) *chi.Mux {
	r := chi.NewRouter()

	metricService := service.NewMetricService(storageCounter, storageGauge)
	metricControllerJSON := handler.NewMetricsController(*metricService, handler.JSONResposeBuilder)
	metricControllerPlain := handler.NewMetricsController(*metricService, handler.PlainResposeBuilder)

	r.Route("/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:(-?)[a-z0-9\\.]+}",
		func(r chi.Router) {
			r.Use(middleware.MetricCtxFromPath)
			r.Post("/", metricControllerPlain.Update)
		},
	)
	r.Route("/value/{metricType}/{metricName:[a-zA-Z0-9_-]+}", func(r chi.Router) {
		r.Use(middleware.MetricCtxFromPath)
		r.Get("/", metricControllerPlain.Get)
	})

	r.Route("/update", func(r chi.Router) {
		r.Use(middleware.ValidateRoute)
		//r.Use(middleware.GzipMiddleware)
		r.Use(middleware.MetricCtxFromBody)
		r.Post("/", metricControllerJSON.Update)
		r.Post("/.+", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		})
	})

	r.Route("/value",
		func(r chi.Router) {
			//r.Use(middleware.GzipMiddleware)
			r.Use(middleware.MetricCtxFromBody)
			r.Post("/", metricControllerJSON.Get)
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
