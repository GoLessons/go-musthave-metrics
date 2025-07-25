package middleware

import (
	"context"
	"fmt"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"strconv"
)

func MetricCtxFromPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, string(server.MetricName))
		metricType := chi.URLParam(r, string(server.MetricType))
		valueRaw := chi.URLParam(r, string(server.MetricValue))

		metric := model.Metrics{
			ID:    name,
			MType: metricType,
		}

		switch metricType {
		case model.Counter:
			value, err := strconv.ParseInt(valueRaw, 10, 64)
			if err != nil && valueRaw != "" {
				http.Error(w, fmt.Sprintf("Unsupported metric value %s = %s", metricType, valueRaw), http.StatusBadRequest)
			}
			metric.Delta = &value
		case model.Gauge:
			value, err := strconv.ParseFloat(valueRaw, 64)
			if err != nil && valueRaw != "" {
				http.Error(w, fmt.Sprintf("Unsupported metric value %s = %s", metricType, valueRaw), http.StatusBadRequest)
			}
			metric.Value = &value
		default:
			http.Error(w, fmt.Sprintf("Metric type %s not defined", metricType), http.StatusBadRequest)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, server.Metric, metric)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MetricCtxFromBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer r.Body.Close()

		var metric model.Metrics
		err = json.Unmarshal(body, &metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, server.Metric, metric)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
