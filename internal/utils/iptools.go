package utils

import (
	"fmt"
	"net"
	"strings"
)

// IPFamilyFromPrefix returnerer "ipv4", "ipv6" eller en feil.
func IPFamilyFromPrefix(prefix string) (string, error) {
	address := strings.Split(prefix, "/")[0]
	ip := net.ParseIP(address)

	if ip == nil {
		return "", fmt.Errorf("invalid ip prefix: %s", prefix)
	}

	if ip.To4() != nil {
		return "ipv4", nil
	}

	return "ipv6", nil
}
