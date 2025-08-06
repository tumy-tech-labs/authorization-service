package contextprovider

import (
	"net"
	"net/http"
)

// GeoIPProvider extracts the remote IP address and returns a stubbed geo-location.
type GeoIPProvider struct{}

// GetContext returns the client's IP and a stubbed country code.
func (GeoIPProvider) GetContext(req *http.Request) (map[string]string, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		ip = req.RemoteAddr
	}
	// Stubbed country lookup
	return map[string]string{
		"ip":          ip,
		"geo_country": "US",
	}, nil
}
