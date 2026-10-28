package monitor

import (
	"math"
	"testing"
)

const pingMacOS = `PING 1.1.1.1 (1.1.1.1): 56 data bytes
64 bytes from 1.1.1.1: icmp_seq=0 ttl=58 time=13.190 ms
64 bytes from 1.1.1.1: icmp_seq=1 ttl=58 time=15.017 ms
64 bytes from 1.1.1.1: icmp_seq=2 ttl=58 time=13.597 ms
64 bytes from 1.1.1.1: icmp_seq=3 ttl=58 time=13.194 ms
64 bytes from 1.1.1.1: icmp_seq=4 ttl=58 time=13.801 ms

--- 1.1.1.1 ping statistics ---
5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 13.190/13.760/15.017/0.671 ms
`

const pingLinux = `PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.
64 bytes from 1.1.1.1: icmp_seq=1 ttl=57 time=10.0 ms
64 bytes from 1.1.1.1: icmp_seq=2 ttl=57 time=14.0 ms
64 bytes from 1.1.1.1: icmp_seq=3 ttl=57 time=12.0 ms

--- 1.1.1.1 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 402ms
rtt min/avg/max/mdev = 10.000/12.000/14.000/1.633 ms
`

const pingBusyBox = `PING 1.1.1.1 (1.1.1.1): 56 data bytes
64 bytes from 1.1.1.1: seq=0 ttl=63 time=15.662 ms
64 bytes from 1.1.1.1: seq=1 ttl=63 time=14.958 ms
64 bytes from 1.1.1.1: seq=2 ttl=63 time=23.532 ms
64 bytes from 1.1.1.1: seq=3 ttl=63 time=25.083 ms
64 bytes from 1.1.1.1: seq=4 ttl=63 time=23.205 ms

--- 1.1.1.1 ping statistics ---
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max = 14.958/20.488/25.083 ms
`

const pingMacOSAllLost = `PING 10.255.255.1 (10.255.255.1): 56 data bytes
Request timeout for icmp_seq 0
Request timeout for icmp_seq 1

--- 10.255.255.1 ping statistics ---
2 packets transmitted, 0 packets received, 100.0% packet loss
`

const pingWindows = "Pinging 192.168.1.1 with 32 bytes of data:\r\n" +
	"Reply from 192.168.1.1: bytes=32 time<1ms TTL=64\r\n" +
	"Reply from 192.168.1.1: bytes=32 time=3ms TTL=64\r\n" +
	"Request timed out.\r\n" +
	"Reply from 192.168.1.1: bytes=32 time=2ms TTL=64\r\n\r\n" +
	"Ping statistics for 192.168.1.1:\r\n" +
	"    Packets: Sent = 4, Received = 3, Lost = 1 (25% loss),\r\n" +
	"Approximate round trip times in milli-seconds:\r\n" +
	"    Minimum = 0ms, Maximum = 3ms, Average = 2ms\r\n"

func TestParsePingUnix(t *testing.T) {
	tests := []struct {
		name              string
		text              string
		avg, jitter, loss float64
		wantErr           bool
	}{
		{"macOS", pingMacOS, 13.76, 1.064, 0, false},
		{"linux", pingLinux, 12, 3, 0, false},
		{"busybox", pingBusyBox, 20.488, 3.177, 0, false},
		{"all lost", pingMacOSAllLost, 0, 0, 100, false},
		{"unknown host macOS", "ping: cannot resolve no-such-host.invalid: Unknown host\n", 0, 0, 0, true},
		{"unknown host busybox", "ping: bad address 'no-such-host.invalid'\n", 0, 0, 0, true},
		{"empty", "", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, jitter, loss, err := parsePingUnix(tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			assertNear(t, "avg", avg, tt.avg)
			assertNear(t, "jitter", jitter, tt.jitter)
			assertNear(t, "loss", loss, tt.loss)
		})
	}
}

func TestParsePingWindows(t *testing.T) {
	avg, jitter, loss, err := parsePingWindows(pingWindows)
	if err != nil {
		t.Fatal(err)
	}
	assertNear(t, "avg", avg, 2)
	assertNear(t, "jitter", jitter, 1.5)
	assertNear(t, "loss", loss, 25)

	if _, _, _, err := parsePingWindows("Ping request could not find host foo. Please check the name and try again.\r\n"); err == nil {
		t.Error("expected error for unknown host")
	}
}

func TestReplyJitter(t *testing.T) {
	tests := []struct {
		name string
		text string
		want float64
	}{
		{"no replies", "Request timeout for icmp_seq 0\n", 0},
		{"single reply", "64 bytes from 1.1.1.1: icmp_seq=0 ttl=58 time=13.1 ms\n", 0},
		{"summary line ignored", "3 packets transmitted, 3 received, 0% packet loss, time 402ms\n", 0},
		{"constant", "time=10 ms\ntime=10 ms\ntime=10 ms\n", 0},
		{"alternating", "time=10 ms\ntime=20 ms\ntime=10 ms\n", 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertNear(t, "jitter", replyJitter(tt.text), tt.want)
		})
	}
}

func assertNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s = %.3f, want %.3f", name, got, want)
	}
}
