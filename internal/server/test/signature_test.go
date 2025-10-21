package test

import (
	"io"
	"net/http"
	"testing"

	"github.com/GoLessons/go-musthave-metrics/internal/common/signature"
	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureVerification(t *testing.T) {
	I, err := NewTester(t, &map[string]any{
		"Key": "test-secret-key",
	})
	require.NoError(t, err)
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
	I, err := NewTester(t, &map[string]any{
		"Key": "test-secret-key",
	})
	require.NoError(t, err)
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
