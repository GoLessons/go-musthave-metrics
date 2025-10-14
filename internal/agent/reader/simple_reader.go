package reader

import (
	"math/rand"
	"sync"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
)

type SimpleMetricsReader struct {
	mu          sync.RWMutex
	pollCount   int64
	randomValue float64
}

func NewSimpleMetricsReader() *SimpleMetricsReader {
	return &SimpleMetricsReader{}
}

func (r *SimpleMetricsReader) Refresh() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pollCount++
	r.randomValue = rand.Float64()
	return nil
}

func (r *SimpleMetricsReader) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pollCount = 0
}

func (r *SimpleMetricsReader) Fetch() ([]model.Metrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := []model.Metrics{
		*model.NewCounter("PollCount", &r.pollCount),
		*model.NewGauge("RandomValue", &r.randomValue),
	}

	return metrics, nil
}
