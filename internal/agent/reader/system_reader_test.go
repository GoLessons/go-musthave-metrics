package reader

import (
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestSystemMetricsRefresh(t *testing.T) {
	reader := NewSystemMetricsReader()

	err := reader.Refresh()
	require.NoError(t, err)
}

func TestSystemMetricsFetch(t *testing.T) {
	reader := NewSystemMetricsReader()

	err := reader.Refresh()
	require.NoError(t, err)
	metrics, err := reader.Fetch()
	require.NoError(t, err)

	assert.Equal(t, 3, len(metrics))

	metricMap := make(map[string]model.Metrics)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	tests := []struct {
		name string
		min  float64
		max  float64
	}{
		{"TotalMemory", 0.0, math.MaxFloat64},
		{"FreeMemory", 0.0, math.MaxFloat64},
		{"CPUutilization1", 0.0, 100.0},
	}

	for _, tt := range tests {
		metric, exists := metricMap[tt.name]
		assert.True(t, exists)
		assert.Equal(t, model.Gauge, metric.MType)
		assert.NotNil(t, metric.Value)
		assert.GreaterOrEqual(t, *metric.Value, tt.min)
		assert.LessOrEqual(t, *metric.Value, tt.max)
	}
}

func TestSystemMetricsInGoroutines(t *testing.T) {
	reader := NewSystemMetricsReader()

	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			err := reader.Refresh()
			assert.NoError(t, err)
			metrics, err := reader.Fetch()
			assert.NoError(t, err)

			assert.Equal(t, 3, len(metrics))
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
