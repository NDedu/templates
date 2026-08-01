package httputil

import (
	"net"
	"net/http"
	"strings"
)

// TrustedProxies contains the list of CIDR blocks of trusted proxies.
var TrustedProxies []*net.IPNet

// AddTrustedProxy adds a CIDR network to the list of trusted proxies.
func AddTrustedProxy(cidr string) error {

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {

		return err
	}

	TrustedProxies = append(TrustedProxies, ipnet)
	return nil
}

func isTrustedProxy(ip net.IP) bool {

	for _, network := range TrustedProxies {

		if network.Contains(ip) {

			return true
		}
	}
	return false
}

// ClientIP returns the true client's IP address.
// If the request comes from a trusted proxy, it parses the X-Forwarded-For header
// from right to left to find the first untrusted IP.
func ClientIP(r *http.Request) string {

	remoteIPStr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {

		remoteIPStr = r.RemoteAddr
	}

	remoteIP := net.ParseIP(remoteIPStr)
	if remoteIP == nil || len(TrustedProxies) == 0 || !isTrustedProxy(remoteIP) {

		return remoteIPStr
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {

		if xri := r.Header.Get("X-Real-IP"); xri != "" {

			xriIP := net.ParseIP(strings.TrimSpace(xri))
			if xriIP != nil {

				return xriIP.String()
			}
		}
		return remoteIPStr
	}

	ips := strings.Split(xff, ",")
	// Read from right to left
	for i := len(ips) - 1; i >= 0; i-- {

		ipStr := strings.TrimSpace(ips[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {

			continue
		}

		if isTrustedProxy(ip) {

			continue
		}
		return ipStr
	}

	if len(ips) > 0 {

		ipStr := strings.TrimSpace(ips[0])
		if net.ParseIP(ipStr) != nil {

			return ipStr
		}
	}

	return remoteIPStr
}

// NetworkID masks an IP address. For IPv4, it returns the IP itself.
// For IPv6, it masks it to a /64 subnet.
func NetworkID(ipStr string) string {

	ip := net.ParseIP(ipStr)
	if ip == nil {

		return ipStr
	}

	if ip.To4() != nil {

		return ipStr // IPv4
	}

	// It's IPv6. Mask to /64 subnet.
	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask).String()
}
