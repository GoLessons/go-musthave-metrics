package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGauge(t *testing.T) {
	I := NewTester(t)
	defer I.Shutdown()

	for _, test := range providerTestGauge() {
		resp, err := I.DoRequest(test.method, test.path, nil, "text/plain")
		require.NoError(t, err)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, test.status, resp.StatusCode)
	}
}

type testGauge struct {
	path   string
	method string
	status int
}

func providerTestGauge() []testGauge {
	return []testGauge{
		{"/update/gauge/test/100.01", http.MethodPost, http.StatusOK},
		{"/update/gauge/test/-100.01", http.MethodPost, http.StatusOK},
		{"/update/gauge/test/100", http.MethodPost, http.StatusOK},
		{"/update/gauge/test/NaN", http.MethodPost, http.StatusNotFound},
		{"/update/unknown/test/100.01", http.MethodPost, http.StatusBadRequest},
		{"/update/gauge/test/100.01", http.MethodDelete, http.StatusMethodNotAllowed},
		{"/update/gauge/test/100.01", http.MethodPut, http.StatusMethodNotAllowed},
	}
}
