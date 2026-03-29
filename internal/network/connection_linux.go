//go:build linux

package network

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ConnectionInfo describes the active network connection regardless of type.
type ConnectionInfo struct {
	Type     string // "wifi", "ethernet", or ""
	RSSI     int    // signal strength dBm (WiFi only; 0 = unavailable)
	Noise    int    // noise floor dBm (unavailable on Linux without extra tools)
	SNR      int    // signal-to-noise ratio dB (unavailable when noise unknown)
	Channel  int    // WiFi channel (0 = unavailable)
	Band     string // "2GHz", "5GHz", "6GHz", or ""
	LinkRate int    // link rate / speed Mbps (0 = unavailable)
	Duplex   string // "full", "half", or "" (Ethernet)
}

// GetConnectionInfo detects the active interface type and returns its details.
func GetConnectionInfo() (ConnectionInfo, bool) {
	iface := linuxActiveInterface()
	if iface == "" {
		return ConnectionInfo{}, false
	}
	if linuxIsWifi(iface) {
		return linuxWifiInfo(iface)
	}
	return linuxEthernetInfo(iface)
}

// linuxActiveInterface returns the interface used for the default route.
func linuxActiveInterface() string {
	// Fast path: /proc/net/route
	if data, err := os.ReadFile("/proc/net/route"); err == nil {
		for _, line := range strings.Split(string(data), "\n")[1:] {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == "00000000" {
				return fields[0]
			}
		}
	}
	// Fallback: ip route get
	out, err := exec.Command("ip", "route", "get", "1.1.1.1").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "dev" {
			return fields[i+1]
		}
	}
	return ""
}

// linuxIsWifi returns true when the interface has a wireless sysfs directory.
func linuxIsWifi(iface string) bool {
	_, err := os.Stat("/sys/class/net/" + iface + "/wireless")
	return err == nil
}

// linuxWifiInfo reads WiFi stats via iw for the given interface.
func linuxWifiInfo(iface string) (ConnectionInfo, bool) {
	out, err := exec.Command("iw", "dev", iface, "link").Output()
	if err != nil {
		return ConnectionInfo{}, false
	}
	if strings.Contains(string(out), "Not connected") {
		return ConnectionInfo{}, false
	}
	var info ConnectionInfo
	info.Type = "wifi"
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "signal:"):
			val := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "signal:")), " dBm")
			info.RSSI, _ = strconv.Atoi(strings.TrimSpace(val))
		case strings.HasPrefix(line, "rx bitrate:"):
			if f := strings.Fields(line); len(f) >= 3 {
				mbps, err := strconv.ParseFloat(f[2], 64)
				if err == nil {
					info.LinkRate = int(mbps)
				}
			}
		case strings.HasPrefix(line, "freq:"):
			if f := strings.Fields(line); len(f) >= 2 {
				freq, _ := strconv.Atoi(f[1])
				switch {
				case freq >= 2401 && freq <= 2495:
					info.Band = "2GHz"
					info.Channel = (freq - 2407) / 5
				case freq >= 4900 && freq <= 5895:
					info.Band = "5GHz"
					info.Channel = (freq - 5000) / 5
				case freq >= 5925:
					info.Band = "6GHz"
					info.Channel = (freq - 5950) / 5
				}
			}
		}
	}
	if info.RSSI == 0 {
		return ConnectionInfo{}, false
	}
	return info, true
}

// linuxEthernetInfo reads link speed and duplex from sysfs.
func linuxEthernetInfo(iface string) (ConnectionInfo, bool) {
	base := "/sys/class/net/" + iface + "/"
	speedData, err := os.ReadFile(base + "speed")
	if err != nil {
		return ConnectionInfo{}, false
	}
	speed, _ := strconv.Atoi(strings.TrimSpace(string(speedData)))
	if speed <= 0 {
		return ConnectionInfo{}, false
	}
	duplexData, _ := os.ReadFile(base + "duplex")
	duplex := strings.TrimSpace(string(duplexData))
	return ConnectionInfo{Type: "ethernet", LinkRate: speed, Duplex: duplex}, true
}
