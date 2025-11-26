package security

import (
	"net/netip"
	"strings"

	"github.com/GoLessons/go-musthave-metrics/internal/common/netaddr"
)

func ParseTrustedCIDR(cidr string) (netip.Prefix, error) {
	return netip.ParsePrefix(cidr)
}

func IsIPTrusted(prefix netip.Prefix, ipString string) (bool, error) {
	trimmed := strings.TrimSpace(ipString)
	if trimmed == "" {
		return false, fmtError("empty ip")
	}
	address, err := netaddr.ParseAddr(trimmed)
	if err != nil {
		return false, err
	}
	return prefix.Contains(address), nil
}

type simpleError struct{ message string }

func (e *simpleError) Error() string { return e.message }

func fmtError(message string) error { return &simpleError{message: message} }
