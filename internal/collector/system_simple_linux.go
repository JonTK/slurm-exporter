//go:build linux
// +build linux

package collector

import (
	"fmt"
	"syscall"
)

// DiskStats holds disk usage statistics
type DiskStats struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// readDiskUsage reads disk usage for a given path using syscall.Statfs (Linux-specific)
func readDiskUsage(path string) (*DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to statfs %s: %w", path, err)
	}

	// Calculate usage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	return &DiskStats{
		Total: total,
		Used:  used,
		Free:  free,
	}, nil
}
