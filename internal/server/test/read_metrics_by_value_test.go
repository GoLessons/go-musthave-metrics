package test

import (
	"io"
	"net/http"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	serverModel "github.com/GoLessons/go-musthave-metrics/internal/server/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type metricTest struct {
	name   string
	method string
	value  float64
	delta  int64
}

var testsByBody = []metricTest{
	{method: model.Counter, name: "PollCount"},
	{method: model.Gauge, name: "RandomValue"},
	{method: model.Gauge, name: "Alloc"},
	{method: model.Gauge, name: "BuckHashSys"},
	{method: model.Gauge, name: "Frees"},
	{method: model.Gauge, name: "GCCPUFraction"},
	{method: model.Gauge, name: "GCSys"},
	{method: model.Gauge, name: "HeapAlloc"},
	{method: model.Gauge, name: "HeapIdle"},
	{method: model.Gauge, name: "HeapInuse"},
	{method: model.Gauge, name: "HeapObjects"},
	{method: model.Gauge, name: "HeapReleased"},
	{method: model.Gauge, name: "HeapSys"},
	{method: model.Gauge, name: "LastGC"},
	{method: model.Gauge, name: "Lookups"},
	{method: model.Gauge, name: "MCacheInuse"},
	{method: model.Gauge, name: "MCacheSys"},
	{method: model.Gauge, name: "MSpanInuse"},
	{method: model.Gauge, name: "MSpanSys"},
	{method: model.Gauge, name: "Mallocs"},
	{method: model.Gauge, name: "NextGC"},
	{method: model.Gauge, name: "NumForcedGC"},
	{method: model.Gauge, name: "NumGC"},
	{method: model.Gauge, name: "OtherSys"},
	{method: model.Gauge, name: "PauseTotalNs"},
	{method: model.Gauge, name: "StackInuse"},
	{method: model.Gauge, name: "StackSys"},
	{method: model.Gauge, name: "Sys"},
	{method: model.Gauge, name: "TotalAlloc"},
}

func TestMetricsNotFound(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	for _, tt := range testsByBody {
		getMetric := model.Metrics{
			ID:    tt.name,
			MType: tt.method,
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{"Content-Type": "application/json"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, tt.name)
	}
}

func TestMetricsExists(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	expectedCounters := map[string]int64{}
	expectedGauges := map[string]float64{}
	for _, tt := range testsByBody {
		switch tt.method {
		case model.Counter:
			delta := tt.delta
			if delta == 0 {
				delta = 100
			}
			c := serverModel.NewCounter(tt.name)
			c.Inc(delta)
			require.NoError(t, I.HaveCouner(*c))
			expectedCounters[tt.name] = delta

		case model.Gauge:
			val := tt.value
			if val == 0 {
				val = 1.234567
			}
			g := serverModel.NewGauge(tt.name)
			g.Set(val)
			require.NoError(t, I.HaveGauge(*g))
			expectedGauges[tt.name] = val
		}
	}

	for _, tt := range testsByBody {
		getMetric := model.Metrics{
			ID:    tt.name,
			MType: tt.method,
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{"Content-Type": "application/json"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, tt.name)
		assert.Containsf(t, resp.Header.Get("Content-Type"), "application/json", tt.name)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, tt.name)

		var result model.Metrics
		require.NoError(t, json.Unmarshal(body, &result), tt.name)

		assert.Equal(t, tt.name, result.ID, tt.name)
		assert.Equal(t, tt.method, result.MType, tt.name)

		switch tt.method {
		case model.Counter:
			require.NotNil(t, result.Delta, tt.name)
			assert.Equal(t, expectedCounters[tt.name], *result.Delta, tt.name)

		case model.Gauge:
			require.NotNil(t, result.Value, tt.name)
			assert.Equal(t, expectedGauges[tt.name], *result.Value, tt.name)
		}
	}
}

func TestMetricsNotFoundGzip(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	for _, tt := range testsByBody {
		getMetric := model.Metrics{
			ID:    tt.name,
			MType: tt.method,
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{
			"Content-Type":    "application/json",
			"Accept-Encoding": "gzip",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, tt.name)
		assert.Containsf(t, resp.Header.Get("Content-Encoding"), "gzip", tt.name)

		_, err = I.ReadGzip(resp)
		require.NoError(t, err, tt.name)
	}
}

func TestMetricsExistsGzip(t *testing.T) {
	I, err := NewTester(t, nil)
	require.NoError(t, err)
	defer I.Shutdown()

	expectedCounters := map[string]int64{}
	expectedGauges := map[string]float64{}

	for _, tt := range testsByBody {
		switch tt.method {
		case model.Counter:
			delta := tt.delta
			if delta == 0 {
				delta = 100
			}
			c := serverModel.NewCounter(tt.name)
			c.Inc(delta)
			require.NoError(t, I.HaveCouner(*c))
			expectedCounters[tt.name] = delta

		case model.Gauge:
			val := tt.value
			if val == 0 {
				val = 1.234567
			}
			g := serverModel.NewGauge(tt.name)
			g.Set(val)
			require.NoError(t, I.HaveGauge(*g))
			expectedGauges[tt.name] = val
		}
	}

	for _, tt := range testsByBody {
		getMetric := model.Metrics{
			ID:    tt.name,
			MType: tt.method,
		}

		resp, err := I.DoRequest(http.MethodPost, "/value/", getMetric, map[string]string{
			"Content-Type":    "application/json",
			"Accept-Encoding": "gzip",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, tt.name)
		assert.Containsf(t, resp.Header.Get("Content-Encoding"), "gzip", tt.name)
		assert.Containsf(t, resp.Header.Get("Content-Type"), "application/json", tt.name)

		body, err := I.ReadGzip(resp)
		require.NoError(t, err, tt.name)

		var result model.Metrics
		require.NoError(t, json.Unmarshal(body, &result), tt.name)

		assert.Equal(t, tt.name, result.ID, tt.name)
		assert.Equal(t, tt.method, result.MType, tt.name)

		switch tt.method {
		case model.Counter:
			require.NotNil(t, result.Delta, tt.name)
			assert.Equal(t, expectedCounters[tt.name], *result.Delta, tt.name)
		case model.Gauge:
			require.NotNil(t, result.Value, tt.name)
			assert.InDelta(t, expectedGauges[tt.name], *result.Value, 1e-9, tt.name)
		}
	}
}
