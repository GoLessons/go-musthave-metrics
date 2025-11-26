package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewTrustedSubnetChecker_InvalidCIDR_ReturnsError(t *testing.T) {
	_, err := NewTrustedSubnetChecker("not-a-cidr", zap.NewNop())
	require.Error(t, err)
}

func TestTrustedSubnetChecker_AllowOnlyTrusted(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	tests := []struct {
		name           string
		cidr           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Missing X-Real-IP → 400",
			cidr:           "127.0.0.0/8",
			headers:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid X-Real-IP → 400",
			cidr:           "127.0.0.0/8",
			headers:        map[string]string{"X-Real-IP": "invalid-ip"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "IPv4 outside CIDR → 403",
			cidr:           "127.0.0.0/8",
			headers:        map[string]string{"X-Real-IP": "10.0.0.1"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv4 inside CIDR → 200",
			cidr:           "127.0.0.0/8",
			headers:        map[string]string{"X-Real-IP": "127.0.0.1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 inside CIDR → 200",
			cidr:           "fd00::/8",
			headers:        map[string]string{"X-Real-IP": "fd12::1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 outside CIDR → 403",
			cidr:           "fd00::/8",
			headers:        map[string]string{"X-Real-IP": "::1"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Ignore X-Forwarded-For → 400",
			cidr:           "127.0.0.0/8",
			headers:        map[string]string{"X-Forwarded-For": "127.0.0.1"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := NewTrustedSubnetChecker(tt.cidr, zap.NewNop())
			require.NoError(t, err)

			mw := checker.AllowOnlyTrusted(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()

			mw.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
