package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"

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

		addr, err := parseAddr(xRealIP)
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

func parseAddr(s string) (netip.Addr, error) {
	if ip := net.ParseIP(s); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			return netip.AddrFrom4([4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}), nil
		}

		var a16 [16]byte
		copy(a16[:], ip.To16())
		return netip.AddrFrom16(a16), nil
	}

	return netip.ParseAddr(s)
}
