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

func TestUpdateBatchJSON(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	var counterDelta1 int64 = 42
	var counterDelta2 int64 = 100
	gaugeValue1 := 42.123
	gaugeValue2 := 100.5

	metrics := []model.Metrics{
		{
			ID:    "test_counter_1",
			MType: "counter",
			Delta: &counterDelta1,
		},
		{
			ID:    "test_counter_2",
			MType: "counter",
			Delta: &counterDelta2,
		},
		{
			ID:    "test_gauge_1",
			MType: "gauge",
			Value: &gaugeValue1,
		},
		{
			ID:    "test_gauge_2",
			MType: "gauge",
			Value: &gaugeValue2,
		},
	}

	resp, err := I.DoRequest(http.MethodPost, "/updates", metrics, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	for _, metric := range metrics {
		getMetric := model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
		}

		valueResp, err := I.DoRequest(http.MethodPost, "/value", getMetric, map[string]string{"Content-Type": "application/json"})
		require.NoError(t, err)
		require.NotNil(t, valueResp)
		defer valueResp.Body.Close()

		assert.Equal(t, http.StatusOK, valueResp.StatusCode)

		body, err := io.ReadAll(valueResp.Body)
		require.NoError(t, err)

		var result model.Metrics
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Equal(t, metric.ID, result.ID)
		assert.Equal(t, metric.MType, result.MType)

		switch metric.MType {
		case "counter":
			assert.NotNil(t, result.Delta)
			assert.Equal(t, *metric.Delta, *result.Delta)
		case "gauge":
			assert.NotNil(t, result.Value)
			assert.Equal(t, *metric.Value, *result.Value)
		}
	}
}

func TestUpdateBatchJSONInvalidMetrics(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	var counterDelta int64 = 42
	gaugeValue := 42.123

	tests := []struct {
		name           string
		metrics        []model.Metrics
		expectedStatus int
	}{
		{
			name:           "Пустой массив метрик",
			metrics:        []model.Metrics{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Отсутствует ID метрики",
			metrics: []model.Metrics{
				{
					MType: "counter",
					Delta: &counterDelta,
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Отсутствует тип метрики",
			metrics: []model.Metrics{
				{
					ID:    "test_metric",
					Delta: &counterDelta,
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Неизвестный тип метрики",
			metrics: []model.Metrics{
				{
					ID:    "test_metric",
					MType: "unknown",
					Delta: &counterDelta,
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Отсутствует значение для counter",
			metrics: []model.Metrics{
				{
					ID:    "test_counter",
					MType: "counter",
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Отсутствует значение для gauge",
			metrics: []model.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Смешанные валидные и невалидные метрики",
			metrics: []model.Metrics{
				{
					ID:    "valid_counter",
					MType: "counter",
					Delta: &counterDelta,
				},
				{
					ID:    "valid_gauge",
					MType: "gauge",
					Value: &gaugeValue,
				},
				{
					ID:    "invalid_metric",
					MType: "unknown",
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := I.DoRequest(http.MethodPost, "/updates", tt.metrics, map[string]string{"Content-Type": "application/json"})
			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestUpdateBatchJSONMultipleUpdates(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	var counterDelta1 int64 = 42
	var counterDelta2 int64 = 100

	firstUpdate := []model.Metrics{
		{
			ID:    "accumulating_counter",
			MType: "counter",
			Delta: &counterDelta1,
		},
	}

	resp, err := I.DoRequest(http.MethodPost, "/updates", firstUpdate, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	secondUpdate := []model.Metrics{
		{
			ID:    "accumulating_counter",
			MType: "counter",
			Delta: &counterDelta2,
		},
	}

	resp, err = I.DoRequest(http.MethodPost, "/updates", secondUpdate, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	getMetric := model.Metrics{
		ID:    "accumulating_counter",
		MType: "counter",
	}

	valueResp, err := I.DoRequest(http.MethodPost, "/value", getMetric, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, valueResp)
	defer valueResp.Body.Close()

	assert.Equal(t, http.StatusOK, valueResp.StatusCode)

	body, err := io.ReadAll(valueResp.Body)
	require.NoError(t, err)

	var result model.Metrics
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	expectedTotal := counterDelta1 + counterDelta2
	assert.Equal(t, expectedTotal, *result.Delta)
}

func TestUpdateBatchJSONGaugeOverwrite(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	gaugeValue1 := 42.123
	gaugeValue2 := 100.5

	firstUpdate := []model.Metrics{
		{
			ID:    "overwritten_gauge",
			MType: "gauge",
			Value: &gaugeValue1,
		},
	}

	resp, err := I.DoRequest(http.MethodPost, "/updates", firstUpdate, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	secondUpdate := []model.Metrics{
		{
			ID:    "overwritten_gauge",
			MType: "gauge",
			Value: &gaugeValue2,
		},
	}

	resp, err = I.DoRequest(http.MethodPost, "/updates", secondUpdate, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	getMetric := model.Metrics{
		ID:    "overwritten_gauge",
		MType: "gauge",
	}

	valueResp, err := I.DoRequest(http.MethodPost, "/value", getMetric, map[string]string{"Content-Type": "application/json"})
	require.NoError(t, err)
	require.NotNil(t, valueResp)
	defer valueResp.Body.Close()

	assert.Equal(t, http.StatusOK, valueResp.StatusCode)

	body, err := io.ReadAll(valueResp.Body)
	require.NoError(t, err)

	var result model.Metrics
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, gaugeValue2, *result.Value)
}

func TestUpdateBatchJSONMethodNotAllowed(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		resp, err := I.DoRequest(method, "/updates", make([]string, 0), map[string]string{"Content-Type": "application/json"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
