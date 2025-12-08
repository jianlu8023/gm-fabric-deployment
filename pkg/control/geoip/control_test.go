package geoip

import (
	"net"
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "Localhost IPv4",
			ip:       "127.0.0.1",
			expected: true,
		},
		{
			name:     "Private IPv4 10.x.x.x",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "Private IPv4 172.16.x.x",
			ip:       "172.16.0.1",
			expected: true,
		},
		{
			name:     "Private IPv4 192.168.x.x",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "Public IPv4",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "Invalid IP",
			ip:       "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrivateIP(tt.ip)
			if result != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

// IsPrivateIP 检查IP地址是否为私有地址
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func TestNewDefaultConfig(t *testing.T) {
	config := getDefaultConfig()
	if config.DatabasePath == "" {
		t.Error("Expected default database path to be set")
	}
	if config.Enabled {
		t.Log("Default enabled is true as expected")
	}
}
