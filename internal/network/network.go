package network

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Info describes the current network the device is on.
type Info struct {
	// ID is a stable identifier: SSID name for Wi-Fi, or "gw:<ip>" for wired/Docker/unknown.
	ID string
	// SSID is the Wi-Fi network name, empty if wired or not detectable.
	SSID string
	// Gateway is the default gateway IP.
	Gateway string
}

// Detect returns the current network identity using SSID when available,
// falling back to the default gateway IP.
// Works on Linux (including Raspberry Pi and Docker), macOS, and Windows.
func Detect() Info {
	gw := DefaultGateway()
	ssid := detectSSID()

	id := "unknown"
	switch {
	case ssid != "":
		id = ssid
	case gw != "":
		id = "gw:" + gw
	}

	return Info{ID: id, SSID: ssid, Gateway: gw}
}

// DefaultGateway returns the default gateway IP in a cross-platform way.
func DefaultGateway() string {
	switch runtime.GOOS {
	case "windows":
		return gatewayWindows()
	case "linux":
		if gw := gatewayLinuxProc(); gw != "" {
			return gw
		}
		return gatewayLinuxCmd()
	default: // darwin, freebsd, openbsd
		return gatewayDarwin()
	}
}

// darwinRoute parses "route get default" and returns the default (gateway, interface).
// Shared by DefaultGateway and GetConnectionInfo to avoid running the command twice.
func darwinRoute() (gateway, iface string) {
	out, _ := exec.Command("route", "get", "default").CombinedOutput()
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "gateway:"):
			gateway = strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
		case strings.HasPrefix(line, "interface:"):
			iface = strings.TrimSpace(strings.TrimPrefix(line, "interface:"))
		}
	}
	return
}

func gatewayDarwin() string {
	gw, _ := darwinRoute()
	return gw
}

// gatewayLinuxProc parses /proc/net/route for the default route (no external commands).
func gatewayLinuxProc() string {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		// Gateway field is 8-hex-char little-endian: "0101A8C0" → 192.168.1.1
		b, err := hex.DecodeString(fields[2])
		if err != nil || len(b) < 4 {
			continue
		}
		return fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0])
	}
	return ""
}

func gatewayLinuxCmd() string {
	out, err := exec.Command("ip", "route", "show", "default").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "default via ") {
			if parts := strings.Fields(line); len(parts) >= 3 {
				if net.ParseIP(parts[2]) != nil {
					return parts[2]
				}
			}
		}
	}
	return ""
}

func gatewayWindows() string {
	out, err := exec.Command("route", "print", "0.0.0.0").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
			if net.ParseIP(fields[2]) != nil {
				return fields[2]
			}
		}
	}
	return ""
}

// detectSSID returns the Wi-Fi SSID of the current connection,
// or "" if not on Wi-Fi or not detectable (e.g. wired, Docker).
func detectSSID() string {
	switch runtime.GOOS {
	case "darwin":
		return ssidDarwin()
	case "linux":
		return ssidLinux()
	case "windows":
		return ssidWindows()
	}
	return ""
}

func ssidDarwin() string {
	for _, iface := range []string{"en0", "en1", "en2", "en3"} {
		out, err := exec.Command("networksetup", "-getairportnetwork", iface).Output()
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(out))
		const prefix = "Current Wi-Fi Network: "
		if strings.HasPrefix(text, prefix) {
			return strings.TrimPrefix(text, prefix)
		}
	}
	return ""
}

func ssidLinux() string {
	if out, err := exec.Command("iwgetid", "-r").Output(); err == nil {
		if ssid := strings.TrimSpace(string(out)); ssid != "" {
			return ssid
		}
	}
	out, err := exec.Command("iw", "dev").Output()
	if err != nil {
		return ""
	}
	var iface string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Interface ") {
			iface = strings.TrimPrefix(line, "Interface ")
		} else if iface != "" && strings.HasPrefix(line, "ssid ") {
			return strings.TrimPrefix(line, "ssid ")
		}
	}
	return ""
}

func ssidWindows() string {
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "SSID" {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}
