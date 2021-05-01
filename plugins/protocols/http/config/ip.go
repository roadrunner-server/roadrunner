package config

import "net"

// Cidrs is a slice of IPNet addresses
type Cidrs []*net.IPNet

// IsTrusted checks if the ip address exists in the provided in the config addresses
func (c *Cidrs) IsTrusted(ip string) bool {
	if len(*c) == 0 {
		return false
	}

	i := net.ParseIP(ip)
	if i == nil {
		return false
	}

	for _, cird := range *c {
		if cird.Contains(i) {
			return true
		}
	}

	return false
}
