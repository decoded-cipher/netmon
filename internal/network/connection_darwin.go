//go:build darwin

package network

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// ConnectionInfo describes the active network connection regardless of type.
type ConnectionInfo struct {
	Type     string // "wifi", "ethernet", or ""
	// WiFi-specific
	RSSI    int    // signal strength dBm (negative; 0 = unavailable)
	Noise   int    // noise floor dBm (negative; 0 = unavailable)
	SNR     int    // signal-to-noise ratio dB (0 = unavailable)
	Channel int    // WiFi channel number (0 = unavailable)
	Band    string // "2GHz", "5GHz", "6GHz", or ""
	// Shared
	LinkRate int // link rate / speed Mbps (0 = unavailable)
	// Ethernet-specific
	Duplex string // "full", "half", or ""
}

// GetConnectionInfo detects the active interface and returns its connection details.
// It reuses darwinRoute() (defined in network.go) to avoid running route twice.
func GetConnectionInfo() (ConnectionInfo, bool) {
	_, activeIface := darwinRoute()

	// Try WiFi via system_profiler
	if spOut, err := exec.Command("system_profiler", "SPAirPortDataType", "-json").Output(); err == nil {
		if info, ok := parseMacWifi(spOut, activeIface); ok {
			return info, true
		}
	}

	// Fall back to Ethernet
	if activeIface != "" {
		return macEthernetInfo(activeIface)
	}
	return ConnectionInfo{}, false
}

// spAirPortData is the relevant subset of system_profiler SPAirPortDataType JSON.
type spAirPortData struct {
	SPAirPortDataType []struct {
		Interfaces []struct {
			Name    string `json:"_name"`
			Current *struct {
				Channel     string `json:"spairport_network_channel"`
				Rate        int    `json:"spairport_network_rate"`
				SignalNoise string `json:"spairport_signal_noise"`
			} `json:"spairport_current_network_information"`
		} `json:"spairport_airport_interfaces"`
	} `json:"SPAirPortDataType"`
}

// parseMacWifi extracts WiFi stats from system_profiler JSON.
// If activeIface is non-empty, only the matching interface is considered.
func parseMacWifi(data []byte, activeIface string) (ConnectionInfo, bool) {
	var sp spAirPortData
	if err := json.Unmarshal(data, &sp); err != nil {
		return ConnectionInfo{}, false
	}
	for _, section := range sp.SPAirPortDataType {
		for _, iface := range section.Interfaces {
			if activeIface != "" && iface.Name != activeIface {
				continue
			}
			cur := iface.Current
			if cur == nil || cur.SignalNoise == "" {
				continue
			}
			var info ConnectionInfo
			info.Type = "wifi"

			// Parse "-24 dBm / -97 dBm"
			parts := strings.SplitN(cur.SignalNoise, " / ", 2)
			if len(parts) == 2 {
				rssi, e1 := strconv.Atoi(strings.TrimSuffix(parts[0], " dBm"))
				noise, e2 := strconv.Atoi(strings.TrimSuffix(parts[1], " dBm"))
				if e1 == nil && e2 == nil {
					info.RSSI = rssi
					info.Noise = noise
					info.SNR = rssi - noise
				}
			}

			// Parse "1 (2GHz, 20MHz)"
			if chParts := strings.SplitN(cur.Channel, " ", 2); len(chParts) >= 1 {
				info.Channel, _ = strconv.Atoi(chParts[0])
			}
			switch {
			case strings.Contains(cur.Channel, "6GHz"):
				info.Band = "6GHz"
			case strings.Contains(cur.Channel, "5GHz"):
				info.Band = "5GHz"
			case strings.Contains(cur.Channel, "2GHz"):
				info.Band = "2GHz"
			}
			info.LinkRate = cur.Rate
			return info, info.RSSI != 0
		}
	}
	return ConnectionInfo{}, false
}

// macEthernetInfo reads the link speed and duplex from ifconfig for a wired interface.
func macEthernetInfo(iface string) (ConnectionInfo, bool) {
	out, err := exec.Command("ifconfig", iface).Output()
	if err != nil {
		return ConnectionInfo{}, false
	}
	info := parseIfconfigMedia(string(out))
	if info.LinkRate == 0 {
		return ConnectionInfo{}, false
	}
	info.Type = "ethernet"
	return info, true
}

// parseIfconfigMedia parses "media: autoselect (1000baseT <full-duplex>)" lines.
func parseIfconfigMedia(output string) ConnectionInfo {
	var info ConnectionInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "media:") {
			continue
		}
		start := strings.Index(line, "(")
		end := strings.LastIndex(line, ")")
		if start < 0 || end <= start {
			break
		}
		inner := line[start+1 : end] // e.g. "1000baseT <full-duplex>"
		fields := strings.Fields(inner)
		if len(fields) == 0 {
			break
		}
		// Extract leading digits from "1000baseT" → 1000
		speedStr := fields[0]
		nonDigit := strings.IndexFunc(speedStr, func(r rune) bool { return r < '0' || r > '9' })
		if nonDigit > 0 {
			info.LinkRate, _ = strconv.Atoi(speedStr[:nonDigit])
		}
		for _, f := range fields[1:] {
			switch f {
			case "<full-duplex>":
				info.Duplex = "full"
			case "<half-duplex>":
				info.Duplex = "half"
			}
		}
		break
	}
	return info
}
