package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCounter(t *testing.T) {
	I := NewTester(t)
	defer I.Shutdown()

	for _, test := range providerTestCounter() {
		resp, err := I.DoRequest(test.method, test.path, nil, "text/plain")
		require.NoError(t, err)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, test.status, resp.StatusCode, test.path)
	}
}

type testCounter struct {
	path   string
	method string
	status int
}

func providerTestCounter() []testCounter {
	return []testCounter{
		{"/update/counter/test/100", http.MethodPost, http.StatusNoContent},
		{"/update/counter/test/-100", http.MethodPost, http.StatusNoContent},
		{"/update/counter/test/100.0", http.MethodPost, http.StatusNoContent},
		{"/update/counter/test/NaN", http.MethodPost, http.StatusBadRequest},
		{"/update/unknown/test/100", http.MethodPost, http.StatusBadRequest},
		{"/update/counter/test/100", http.MethodDelete, http.StatusMethodNotAllowed},
		{"/update/counter/test/100", http.MethodPut, http.StatusMethodNotAllowed},
	}
}
