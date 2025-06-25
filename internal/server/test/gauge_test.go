package test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGauge(t *testing.T) {
	I := NewTester()
	defer I.Shutdown()

	for _, test := range providerTestGauge() {
		resp, err := I.DoRequest(test.method, test.path, nil)
		defer resp.Close()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
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
