package netaddr

import (
	"net"
	"net/netip"
)

func ParseAddr(s string) (netip.Addr, error) {
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
