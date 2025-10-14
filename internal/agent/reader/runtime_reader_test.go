package reader

import (
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRuntimeMetricsRefresh(t *testing.T) {
	reader := NewRuntimeMetricsReader()

	err := reader.Refresh()
	require.NoError(t, err)
}

func TestRuntimeMetricsFetch(t *testing.T) {
	reader := NewRuntimeMetricsReader()

	err := reader.Refresh()
	require.NoError(t, err)
	metrics, err := reader.Fetch()
	require.NoError(t, err)

	assert.Equal(t, 27, len(metrics))

	for _, metric := range metrics {
		assert.Equal(t, model.Gauge, metric.MType)
		assert.NotNil(t, metric.Value)
		assert.NotEmpty(t, metric.ID)
	}
}
