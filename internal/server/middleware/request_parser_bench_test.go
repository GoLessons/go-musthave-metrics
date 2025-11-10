package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
)

func BenchmarkMetricCtxFromPath_Counter(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value(server.Metric)
	})
	mw := MetricCtxFromPath(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/update/counter/x/123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add(string(server.MetricType), "counter")
		rctx.URLParams.Add(string(server.MetricName), "x")
		rctx.URLParams.Add(string(server.MetricValue), "123")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func BenchmarkMetricCtxFromPath_Gauge(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value(server.Metric)
	})
	mw := MetricCtxFromPath(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/update/gauge/y/1.23", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add(string(server.MetricType), "gauge")
		rctx.URLParams.Add(string(server.MetricName), "y")
		rctx.URLParams.Add(string(server.MetricValue), "1.23")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func BenchmarkMetricCtxFromBody(b *testing.B) {
	m := model.NewGauge("metric_json", ptrFloat(42.5))
	body, _ := json.Marshal(m)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value(server.Metric)
	})
	mw := MetricCtxFromBody(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func BenchmarkMetricsListCtxFromBody(b *testing.B) {
	metrics := make([]model.Metrics, 64)
	for i := range metrics {
		metrics[i] = *model.NewCounter("mc_"+string(rune(i)), ptrInt(int64(i)))
	}
	body, _ := json.Marshal(metrics)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.Context().Value(server.MetricsList)
	})
	mw := MetricsListCtxFromBody(next)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, r)
	}
}

func ptrInt(v int64) *int64       { return &v }
func ptrFloat(v float64) *float64 { return &v }
