package test

import (
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestUpdateCounterJSON(t *testing.T) {
	I := NewTester(t)
	defer I.Shutdown()

	var counterDelta int64 = 42
	metric := model.Metrics{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: &counterDelta,
	}

	resp, err := I.DoRequest(http.MethodPost, "/update", metric, "application/json")
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	getMetric := model.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	resp, err = I.DoRequest(http.MethodPost, "/value", getMetric, "application/json")
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result model.Metrics
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, metric.ID, result.ID)
	assert.Equal(t, metric.MType, result.MType)
	assert.NotNil(t, result.Delta)
	assert.Equal(t, metric.Delta, result.Delta)
}
