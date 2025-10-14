package reader

import (
	"runtime"
	"sync"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
)

type RuntimeMetricsReader struct {
	mu       sync.RWMutex
	memStats *runtime.MemStats
	metrics  map[string]func(*runtime.MemStats) float64
}

func NewRuntimeMetricsReader() *RuntimeMetricsReader {
	return &RuntimeMetricsReader{
		memStats: &runtime.MemStats{},
		metrics: map[string]func(*runtime.MemStats) float64{
			"Alloc":         func(ms *runtime.MemStats) float64 { return float64(ms.Alloc) },
			"BuckHashSys":   func(ms *runtime.MemStats) float64 { return float64(ms.BuckHashSys) },
			"Frees":         func(ms *runtime.MemStats) float64 { return float64(ms.Frees) },
			"GCCPUFraction": func(ms *runtime.MemStats) float64 { return float64(ms.GCCPUFraction) },
			"GCSys":         func(ms *runtime.MemStats) float64 { return float64(ms.GCSys) },
			"HeapAlloc":     func(ms *runtime.MemStats) float64 { return float64(ms.HeapAlloc) },
			"HeapIdle":      func(ms *runtime.MemStats) float64 { return float64(ms.HeapIdle) },
			"HeapInuse":     func(ms *runtime.MemStats) float64 { return float64(ms.HeapInuse) },
			"HeapObjects":   func(ms *runtime.MemStats) float64 { return float64(ms.HeapObjects) },
			"HeapReleased":  func(ms *runtime.MemStats) float64 { return float64(ms.HeapReleased) },
			"HeapSys":       func(ms *runtime.MemStats) float64 { return float64(ms.HeapSys) },
			"LastGC":        func(ms *runtime.MemStats) float64 { return float64(ms.LastGC) },
			"Lookups":       func(ms *runtime.MemStats) float64 { return float64(ms.Lookups) },
			"MCacheInuse":   func(ms *runtime.MemStats) float64 { return float64(ms.MCacheInuse) },
			"MCacheSys":     func(ms *runtime.MemStats) float64 { return float64(ms.MCacheSys) },
			"MSpanInuse":    func(ms *runtime.MemStats) float64 { return float64(ms.MSpanInuse) },
			"MSpanSys":      func(ms *runtime.MemStats) float64 { return float64(ms.MSpanSys) },
			"Mallocs":       func(ms *runtime.MemStats) float64 { return float64(ms.Mallocs) },
			"NextGC":        func(ms *runtime.MemStats) float64 { return float64(ms.NextGC) },
			"NumForcedGC":   func(ms *runtime.MemStats) float64 { return float64(ms.NumForcedGC) },
			"NumGC":         func(ms *runtime.MemStats) float64 { return float64(ms.NumGC) },
			"OtherSys":      func(ms *runtime.MemStats) float64 { return float64(ms.OtherSys) },
			"PauseTotalNs":  func(ms *runtime.MemStats) float64 { return float64(ms.PauseTotalNs) },
			"StackInuse":    func(ms *runtime.MemStats) float64 { return float64(ms.StackInuse) },
			"StackSys":      func(ms *runtime.MemStats) float64 { return float64(ms.StackSys) },
			"Sys":           func(ms *runtime.MemStats) float64 { return float64(ms.Sys) },
			"TotalAlloc":    func(ms *runtime.MemStats) float64 { return float64(ms.TotalAlloc) },
		},
	}
}

func (r *RuntimeMetricsReader) Refresh() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	runtime.ReadMemStats(r.memStats)
	return nil
}

func (r *RuntimeMetricsReader) Fetch() ([]model.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]model.Metrics, 0, len(r.metrics))
	for name, getMetric := range r.metrics {
		value := getMetric(r.memStats)
		metrics = append(metrics, *model.NewGauge(name, &value))
	}

	return metrics, nil
}
