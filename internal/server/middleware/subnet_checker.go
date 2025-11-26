package middleware

import (
	"net/http"
	"net/netip"
	"strings"

	"github.com/GoLessons/go-musthave-metrics/internal/common/netaddr"
	"go.uber.org/zap"
)

type TrustedSubnetCheckMiddleware struct {
	cidr   netip.Prefix
	logger *zap.Logger
}

func NewTrustedSubnetChecker(cidr string, logger *zap.Logger) (*TrustedSubnetCheckMiddleware, error) {
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, err
	}

	return &TrustedSubnetCheckMiddleware{cidr: p, logger: logger}, nil
}

func (m *TrustedSubnetCheckMiddleware) AllowOnlyTrusted(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xRealIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if xRealIP == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		addr, err := netaddr.ParseAddr(xRealIP)
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("invalid X-Real-IP", zap.String("ip", xRealIP))
			}

			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if !m.cidr.Contains(addr) {
			if m.logger != nil {
				m.logger.Warn("ip not in trusted subnet", zap.String("ip", xRealIP), zap.String("cidr", m.cidr.String()))
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
