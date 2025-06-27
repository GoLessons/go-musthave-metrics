package agent

import "runtime"

type GetMetric func(*runtime.MemStats) float64

type memStatsReader[T float64] struct {
	memStats       *runtime.MemStats
	runtimeMetrics map[string]GetMetric
}

func NewMemStatsReader[T float64]() *memStatsReader[T] {
	return &memStatsReader[T]{
		runtimeMetrics: map[string]GetMetric{
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

func (r *memStatsReader[T]) Refresh() {
	if r.memStats == nil {
		r.memStats = &runtime.MemStats{}
	}

	runtime.ReadMemStats(r.memStats)
}

func (r *memStatsReader[T]) Get(name string) (float64, bool) {
	if r.memStats == nil {
		r.Refresh()
	}

	getMetric, ok := r.runtimeMetrics[name]
	if !ok {
		return 0, false
	}

	return getMetric(r.memStats), true
}

func (r *memStatsReader[T]) SupportedMetrics() []string {
	keys := []string{}
	for key := range r.runtimeMetrics {
		keys = append(keys, key)
	}

	return keys
}
