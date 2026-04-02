//go:build windows

package server

import "os"

// restart exits so a process manager (e.g. Windows Service) can restart with the new binary.
func restart(_ string) {
	os.Exit(0)
}
