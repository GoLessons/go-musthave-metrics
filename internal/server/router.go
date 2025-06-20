package server

import (
	"context"
	common "github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/handler"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

var storageCounter = storage.NewMemStorage[model.Counter]()
var storageGauge = storage.NewMemStorage[model.Gauge]()

func InitRouter() *chi.Mux {
	routes := map[string]http.Handler{
		common.Counter: handler.NewUpdateCounter(storageCounter),
		common.Gauge:   handler.NewUpdateGauge(storageGauge),
	}

	r := chi.NewRouter()
	for metricType, metricHandler := range routes {
		r.Route("/update/"+metricType+"/{metricName:[a-zA-Z0-9_-]+}/{metricValue:[a-z0-9\\.]+}", func(r chi.Router) {
			r.Use(initMetricCtx)
			r.Get("/", metricHandler.ServeHTTP)
			r.Post("/", metricHandler.ServeHTTP)
		})
	}
	r.Post(
		"/update/{metricType}/{metricName:[a-zA-Z0-9_-]+}/{metricValue:[a-z0-9\\.]+}",
		func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			http.Error(w, "Wrong metric type: "+metricType, http.StatusBadRequest)
		},
	)

	return r
}

func initMetricCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		ctx := r.Context()
		ctx = context.WithValue(ctx, "metricName", metricName)
		ctx = context.WithValue(ctx, "metricValue", metricValue)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
