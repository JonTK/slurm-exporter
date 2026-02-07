//go:build windows
// +build windows

package collector

import (
	"errors"
)

// DiskStats holds disk usage statistics
type DiskStats struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// readDiskUsage is not implemented on Windows
func readDiskUsage(path string) (*DiskStats, error) {
	return nil, errors.New("disk usage monitoring not supported on Windows")
}
