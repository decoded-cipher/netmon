//go:build windows

package network

import (
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ConnectionInfo describes the active network connection regardless of type.
type ConnectionInfo struct {
	Type     string // "wifi", "ethernet", or ""
	RSSI     int    // signal strength dBm (WiFi only; 0 = unavailable)
	Noise    int    // noise floor dBm (not exposed on Windows)
	SNR      int    // signal-to-noise ratio dB (unavailable when noise unknown)
	Channel  int    // WiFi channel (0 = unavailable)
	Band     string // "2GHz", "5GHz", "6GHz", or ""
	LinkRate int    // link rate / speed Mbps (0 = unavailable)
	Duplex   string // not exposed on Windows
}

// GetConnectionInfo detects the adapter carrying the default route and returns its details.
func GetConnectionInfo() (ConnectionInfo, bool) {
	adapter, ok := windowsActiveAdapter()
	if !ok {
		return ConnectionInfo{}, false
	}
	linkRate := int(adapter.speed / 1_000_000)

	switch adapter.ifType {
	case windows.IF_TYPE_IEEE80211:
		info := ConnectionInfo{Type: "wifi", LinkRate: linkRate}
		if out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output(); err == nil {
			if wifi, ok := parseNetshWlan(string(out), adapter.name); ok {
				if wifi.LinkRate == 0 {
					wifi.LinkRate = linkRate
				}
				info = wifi
			}
		}
		return info, true
	case windows.IF_TYPE_ETHERNET_CSMACD:
		if linkRate <= 0 {
			return ConnectionInfo{}, false
		}
		return ConnectionInfo{Type: "ethernet", LinkRate: linkRate}, true
	}
	return ConnectionInfo{}, false
}

type windowsAdapter struct {
	name   string
	ifType uint32
	speed  uint64
}

// windowsActiveAdapter returns the adapter Windows would use to reach the internet.
func windowsActiveAdapter() (windowsAdapter, bool) {
	var bestIdx uint32
	if err := windows.GetBestInterfaceEx(&windows.SockaddrInet4{Addr: [4]byte{1, 1, 1, 1}}, &bestIdx); err != nil {
		return windowsAdapter{}, false
	}

	size := uint32(15000)
	for {
		buf := make([]byte, size)
		head := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
		flags := uint32(windows.GAA_FLAG_SKIP_ANYCAST | windows.GAA_FLAG_SKIP_MULTICAST | windows.GAA_FLAG_SKIP_DNS_SERVER)
		err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, flags, 0, head, &size)
		if err == windows.ERROR_BUFFER_OVERFLOW {
			continue
		}
		if err != nil {
			return windowsAdapter{}, false
		}
		for aa := head; aa != nil; aa = aa.Next {
			if aa.IfIndex != bestIdx || aa.OperStatus != windows.IfOperStatusUp {
				continue
			}
			return windowsAdapter{
				name:   windows.UTF16PtrToString(aa.FriendlyName),
				ifType: aa.IfType,
				speed:  aa.ReceiveLinkSpeed,
			}, true
		}
		return windowsAdapter{}, false
	}
}

// parseNetshWlan extracts WiFi stats from "netsh wlan show interfaces" output.
func parseNetshWlan(output, iface string) (ConnectionInfo, bool) {
	var blocks []map[string]string
	for _, line := range strings.Split(output, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, val = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(val)
		if key == "name" || len(blocks) == 0 {
			blocks = append(blocks, map[string]string{})
		}
		blocks[len(blocks)-1][key] = val
	}

	var fields map[string]string
	for _, b := range blocks {
		if strings.EqualFold(b["name"], iface) && b["signal"] != "" {
			fields = b
			break
		}
	}
	if fields == nil {
		for _, b := range blocks {
			if b["signal"] != "" {
				fields = b
				break
			}
		}
	}
	if fields == nil {
		return ConnectionInfo{}, false
	}

	info := ConnectionInfo{Type: "wifi"}

	// Signal quality: 0% = -100 dBm, 100% = -50 dBm
	if rssi, err := strconv.Atoi(fields["rssi"]); err == nil && rssi < 0 {
		info.RSSI = rssi
	} else if pct, err := strconv.Atoi(strings.TrimSuffix(fields["signal"], "%")); err == nil {
		info.RSSI = pct/2 - 100
	}

	info.Channel, _ = strconv.Atoi(fields["channel"])
	switch band := fields["band"]; {
	case strings.HasPrefix(band, "6"):
		info.Band = "6GHz"
	case strings.HasPrefix(band, "5"):
		info.Band = "5GHz"
	case strings.HasPrefix(band, "2"):
		info.Band = "2GHz"
	case info.Channel >= 1 && info.Channel <= 14:
		info.Band = "2GHz"
	case info.Channel >= 32:
		info.Band = "5GHz"
	}

	if mbps, err := strconv.ParseFloat(fields["receive rate (mbps)"], 64); err == nil {
		info.LinkRate = int(mbps)
	}
	return info, info.RSSI != 0
}
