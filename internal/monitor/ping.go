package monitor

import (
	"context"
	"errors"
	"math"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var errNoPingStats = errors.New("no ping statistics in output")

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
	deadline := int(math.Ceil(float64(count)*0.2)) + 3
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deadline+5)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "ping", "-n", strconv.Itoa(count), "-w", "1000", target)
	case "linux":
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(count), "-i", "0.2", "-w", strconv.Itoa(deadline), target)
	default:
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(count), "-i", "0.2", "-t", strconv.Itoa(deadline), target)
	}
	out, runErr := cmd.CombinedOutput()
	text := string(out)

	if runtime.GOOS == "windows" {
		avg, jitter, loss, err = parsePingWindows(text)
	} else {
		avg, jitter, loss, err = parsePingUnix(text)
	}
	if err != nil {
		if msg := strings.TrimSpace(text); msg != "" {
			err = errors.New(strings.SplitN(msg, "\n", 2)[0])
		} else if runErr != nil {
			err = runErr
		}
	}
	return
}

// parsePingUnix handles Linux (mdev) and macOS (stddev) ping output.
func parsePingUnix(text string) (avg, jitter, loss float64, err error) {
	lossRe := regexp.MustCompile(`([\d.]+)% packet loss`)
	m := lossRe.FindStringSubmatch(text)
	if m == nil {
		return 0, 0, 0, errNoPingStats
	}
	loss, _ = strconv.ParseFloat(m[1], 64)
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
	m := lossRe.FindStringSubmatch(text)
	if m == nil {
		return 0, 0, 0, errNoPingStats
	}
	loss, _ = strconv.ParseFloat(m[1], 64)
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
