package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var targets = []string{
	"1.1.1.1",
	"8.8.8.8",
}

func calculateAverage(times []float64) float64 {
	if len(times) == 0 {
		return 0
	}
	var sum float64
	for _, time := range times {
		sum += time
	}
	return sum / float64(len(times))
}

func parsePingOutput(output []byte) []float64 {	lines := strings.Split(string(output), "\n")
	var pingTimes []float64
	re := regexp.MustCompile(`time=([\d.]+) ms`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			time, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				pingTimes = append(pingTimes, time)
			}
		}
	}
	return pingTimes
}

func ping(target string) []float64 {
	cmd := exec.Command("ping", "-c", "4", target)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error pinging %s: %v\n", target, err)
		return nil
	}
	return parsePingOutput(output)
}

func main() {
	for _, target := range targets {
		fmt.Printf("Pinging %s...\n", target)
		pingTimes := ping(target)
		if len(pingTimes) > 0 {
			average := calculateAverage(pingTimes)
			fmt.Printf("Average ping time to %s: %.2f ms\n", target, average)
		} else {
			fmt.Printf("No ping times recorded for %s.\n", target)
		}
	}
}
