package monitor

import (
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

type pingResult struct {
	Host       string
	IP         string
	AvgMs      float64
	JitterMs   float64
	PacketLoss float64
	Err        error
}

// runPing pings target count times and returns avg RTT, jitter, and packet loss.
// Works on Linux, macOS, and Windows.
func runPing(target string, count int) (avg, jitter, loss float64, err error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), target)
	} else {
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-i", "0.2", target)
	}
	out, _ := cmd.CombinedOutput()
	text := string(out)

	if runtime.GOOS == "windows" {
		return parsePingWindows(text)
	}
	return parsePingUnix(text)
}

// parsePingUnix handles Linux (mdev) and macOS (stddev) ping output.
func parsePingUnix(text string) (avg, jitter, loss float64, err error) {
	lossRe := regexp.MustCompile(`([\d.]+)% packet loss`)
	if m := lossRe.FindStringSubmatch(text); m != nil {
		loss, _ = strconv.ParseFloat(m[1], 64)
	}
	// macOS: min/avg/max/stddev — Linux: min/avg/max/mdev
	rttRe := regexp.MustCompile(`min/avg/max/\w+ = [\d.]+/([\d.]+)/[\d.]+/([\d.]+)`)
	if m := rttRe.FindStringSubmatch(text); m != nil {
		avg, _ = strconv.ParseFloat(m[1], 64)
		jitter, _ = strconv.ParseFloat(m[2], 64)
	}
	return
}

// parsePingWindows handles Windows ping output.
// Windows doesn't report stddev; jitter is approximated as (max-min)/2.
func parsePingWindows(text string) (avg, jitter, loss float64, err error) {
	lossRe := regexp.MustCompile(`\((\d+)% loss\)`)
	if m := lossRe.FindStringSubmatch(text); m != nil {
		loss, _ = strconv.ParseFloat(m[1], 64)
	}
	avgRe := regexp.MustCompile(`Average = (\d+)ms`)
	minRe := regexp.MustCompile(`Minimum = (\d+)ms`)
	maxRe := regexp.MustCompile(`Maximum = (\d+)ms`)
	if m := avgRe.FindStringSubmatch(text); m != nil {
		avg, _ = strconv.ParseFloat(m[1], 64)
	}
	var minMs, maxMs float64
	if m := minRe.FindStringSubmatch(text); m != nil {
		minMs, _ = strconv.ParseFloat(m[1], 64)
	}
	if m := maxRe.FindStringSubmatch(text); m != nil {
		maxMs, _ = strconv.ParseFloat(m[1], 64)
	}
	jitter = (maxMs - minMs) / 2
	return
}

func resolveIP(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return ""
	}
	return ips[0].String()
}

func measureDNS(host string) (time.Duration, error) {
	start := time.Now()
	_, err := net.LookupHost(host)
	return time.Since(start), err
}
