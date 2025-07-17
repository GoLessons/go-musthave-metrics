package router

import (
	"context"
	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	common "github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var storageCounter = storage.NewMemStorage[model.Counter]()
var storageGauge = storage.NewMemStorage[model.Gauge]()

func InitRouter() *chi.Mux {
	routes := map[string]handler.MetricController{
		common.Counter: handler.NewCounterController(storageCounter),
		common.Gauge:   handler.NewGaugeController(storageGauge),
	}

	r := chi.NewRouter()

	metricService := service.NewMetricService(storageCounter, storageGauge)
	metricsJSONController := handler.NewMetricsController(*metricService)
	r.Post("/update", metricsJSONController.Update)
	r.Post("/value", metricsJSONController.Get)

	for metricType, metricController := range routes {
		r.Route("/update/"+metricType+"/{metricName:[a-zA-Z0-9_-]+}/{metricValue:(-?)[a-z0-9\\.]+}", func(r chi.Router) {
			r.Use(initMetricCtx)
			r.Post("/", metricController.Update)
			r.Get("/", metricController.Update)
		})
		r.Route("/value/"+metricType+"/{metricName:[a-zA-Z0-9_-]+}", func(r chi.Router) {
			r.Use(initMetricCtx)
			r.Get("/", metricController.Get)
		})
	}

	r.Post(
		"/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:[a-z0-9\\.]+}",
		func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			http.Error(w, "Wrong metric type: "+metricType, http.StatusBadRequest)
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

func initMetricCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, string(server.MetricName))
		value := chi.URLParam(r, string(server.MetricValue))
		ctx := r.Context()
		ctx = context.WithValue(ctx, server.MetricName, name)
		ctx = context.WithValue(ctx, server.MetricValue, value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
