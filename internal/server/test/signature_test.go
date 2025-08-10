package test

import (
	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestSignatureVerification(t *testing.T) {
	I := NewTester(t, &map[string]any{
		"Key": "test-secret-key",
	})
	defer I.Shutdown()

	var counterDelta int64 = 42
	metric := model.Metrics{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: &counterDelta,
	}

	body, err := json.Marshal(metric)
	require.NoError(t, err)

	t.Log("!Request: ", string(body))

	signer := signature.NewSign("test-secret-key")
	correctHash, err := signer.Hash(body)
	require.NoError(t, err)

	t.Log("!CorrectHash: ", correctHash)

	resp, err := I.DoRequest(
		http.MethodPost,
		"/update",
		body,
		map[string]string{
			"Content-Type": "application/json",
			"HashSHA256":   correctHash,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	invalidHash := "invalid-hash-value"
	resp, err = I.DoRequest(
		http.MethodPost,
		"/update",
		metric,
		map[string]string{
			"Content-Type": "application/json",
			"HashSHA256":   invalidHash,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignatureInResponse(t *testing.T) {
	I := NewTester(t, &map[string]any{
		"Key": "test-secret-key",
	})
	defer I.Shutdown()

	var counterDelta int64 = 42
	metric := model.Metrics{
		ID:    "test_counter_response",
		MType: model.Counter,
		Delta: &counterDelta,
	}

	body, err := json.Marshal(metric)
	require.NoError(t, err)

	signer := signature.NewSign("test-secret-key")
	hash, err := signer.Hash(body)
	require.NoError(t, err)

	resp, err := I.DoRequest(
		http.MethodPost,
		"/update",
		body,
		map[string]string{
			"Content-Type": "application/json",
			"HashSHA256":   hash,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NotEmpty(t, resp.Header.Get("HashSHA256"))

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	respHash := resp.Header.Get("HashSHA256")
	calculatedHash, err := signer.Hash(respBody)
	require.NoError(t, err)

	assert.Equal(t, calculatedHash, respHash)
}

func TestBatchUpdateWithSignature(t *testing.T) {
	options := map[string]any{
		"Key": "test-secret-key",
	}
	I := NewTester(t, &options)
	defer I.Shutdown()

	var counterDelta1 int64 = 42
	var counterDelta2 int64 = 100
	gaugeValue1 := 42.123
	gaugeValue2 := 100.5

	metrics := []model.Metrics{
		{
			ID:    "test_counter_1",
			MType: model.Counter,
			Delta: &counterDelta1,
		},
		{
			ID:    "test_counter_2",
			MType: model.Counter,
			Delta: &counterDelta2,
		},
		{
			ID:    "test_gauge_1",
			MType: model.Gauge,
			Value: &gaugeValue1,
		},
		{
			ID:    "test_gauge_2",
			MType: model.Gauge,
			Value: &gaugeValue2,
		},
	}

	body, err := json.Marshal(metrics)
	require.NoError(t, err)

	signer := signature.NewSign("test-secret-key")
	correctHash, err := signer.Hash(body)
	require.NoError(t, err)

	resp, err := I.DoRequest(
		http.MethodPost,
		"/updates",
		metrics,
		map[string]string{
			"Content-Type": "application/json",
			"HashSHA256":   correctHash,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("HashSHA256")) // no hash because body is empty
}
