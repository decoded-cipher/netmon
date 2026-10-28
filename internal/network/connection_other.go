//go:build !darwin && !linux && !windows

package network

// ConnectionInfo describes the active network connection regardless of type.
type ConnectionInfo struct {
	Type     string
	RSSI     int
	Noise    int
	SNR      int
	Channel  int
	Band     string
	LinkRate int
	Duplex   string
}

// GetConnectionInfo is a no-op stub for unsupported platforms.
func GetConnectionInfo() (ConnectionInfo, bool) {
	return ConnectionInfo{}, false
}
