package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/storage"
	config2 "github.com/GoLessons/go-musthave-metrics/internal/config"
	serverConfig "github.com/GoLessons/go-musthave-metrics/internal/server/config"
	"github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/GoLessons/go-musthave-metrics/internal/server/service"
	"github.com/GoLessons/go-musthave-metrics/pkg/container"
	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func buildRouter(withSignature bool) *chi.Mux {
	opts := map[string]any{
		"DatabaseDsn":                "",
		"DumpConfig.Restore":         false,
		"DumpConfig.FileStoragePath": "bench-metrics.json",
	}
	if withSignature {
		opts["Key"] = "bench_key"
	} else {
		opts["Key"] = ""
	}

	cfg, _ := serverConfig.LoadConfig(&opts)
	storageCounter := storage.NewMemStorage[model.Counter]()
	storageGauge := storage.NewMemStorage[model.Gauge]()
	metricService := service.NewMetricService(storageCounter, storageGauge)
	logger := zap.NewNop()

	c := container.NewSimpleContainer(map[string]any{
		"logger":         logger,
		"config":         cfg,
		"counterStorage": storageCounter,
		"gaugeStorage":   storageGauge,
		"metricService":  metricService,
	})
	container.SimpleRegisterFactory(&c, "db", config2.DBFactory())
	container.SimpleRegisterFactory(&c, "router", RouterFactory())

	r, _ := container.GetService[chi.Mux](c, "router")
	return r
}

func BenchmarkRouter_UpdateJSON_NoSignature(b *testing.B) {
	r := buildRouter(false)
	body := []byte(`{"id":"bench_update","type":"gauge","value":1.23}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}

func BenchmarkRouter_ValueJSON_NoSignature(b *testing.B) {
	r := buildRouter(false)

	// Ensure metric exists
	{
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(`{"id":"bench_value","type":"counter","delta":5}`)))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}

	body, _ := json.Marshal(map[string]any{"id": "bench_value", "type": "counter"})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}
