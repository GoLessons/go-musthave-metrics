package test

import (
	"io"
	"net/http"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCounterJSON(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	var counterDelta int64 = 42
	metric := model.Metrics{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: &counterDelta,
	}

	resp, err := I.DoRequest(http.MethodPost, "/update", metric, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	getMetric := model.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	resp, err = I.DoRequest(http.MethodPost, "/value", getMetric, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Log(string(body))

	var result model.Metrics
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, metric.ID, result.ID)
	assert.Equal(t, metric.MType, result.MType)
	assert.NotNil(t, result.Delta)
	assert.Equal(t, metric.Delta, result.Delta)
}
