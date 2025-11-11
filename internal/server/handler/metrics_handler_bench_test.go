package handler

import (
	"context"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"go.uber.org/zap"
)

func BenchmarkMetricsHandler_Update_Plain(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := service.NewMetricService(sCounter, sGauge)
	h := NewMetricsController(*ms, PlainResposeBuilder, zap.NewNop(), nil)

	metric := model.NewGauge("bench_plain", ptrFloat(123.456))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/update/gauge/bench_plain/123.456", nil)
		r = r.WithContext(context.WithValue(r.Context(), server.Metric, *metric))
		w := httptest.NewRecorder()
		h.Update(w, r)
	}
}

func BenchmarkMetricsHandler_Update_JSON(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := service.NewMetricService(sCounter, sGauge)
	h := NewMetricsController(*ms, JSONResposeBuilder, zap.NewNop(), nil)

	metric := model.NewCounter("bench_json", ptrInt(42))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/update", nil)
		r = r.WithContext(context.WithValue(r.Context(), server.Metric, *metric))
		w := httptest.NewRecorder()
		h.Update(w, r)
	}
}

func BenchmarkMetricsHandler_Get_JSON(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := service.NewMetricService(sCounter, sGauge)
	h := NewMetricsController(*ms, JSONResposeBuilder, zap.NewNop(), nil)

	// Seed counter
	_ = ms.Save(*model.NewCounter("bench_get_json", ptrInt(5)))

	metric := model.Metrics{ID: "bench_get_json", MType: model.Counter}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("GET", "/value", nil)
		r = r.WithContext(context.WithValue(r.Context(), server.Metric, metric))
		w := httptest.NewRecorder()
		h.Get(w, r)
	}
}

func BenchmarkMetricsHandler_UpdateBatch_JSON(b *testing.B) {
	sCounter := storage.NewMemStorage[serverModel.Counter]()
	sGauge := storage.NewMemStorage[serverModel.Gauge]()
	ms := service.NewMetricService(sCounter, sGauge)
	h := NewMetricsController(*ms, JSONResposeBuilder, zap.NewNop(), nil)

	metrics := make([]model.Metrics, 64)
	for i := range metrics {
		metrics[i] = *model.NewCounter("bench_batch_"+strconv.Itoa(i), ptrInt(int64(i)))
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/updates", nil)
		r = r.WithContext(context.WithValue(r.Context(), server.MetricsList, metrics))
		w := httptest.NewRecorder()
		h.UpdateBatch(w, r)
	}
}

func ptrInt(v int64) *int64       { return &v }
func ptrFloat(v float64) *float64 { return &v }
