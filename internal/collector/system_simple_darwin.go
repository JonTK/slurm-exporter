//go:build darwin
// +build darwin

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

// readDiskUsage reads disk usage for a given path using syscall.Statfs (Darwin-specific)
func readDiskUsage(path string) (*DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to statfs %s: %w", path, err)
	}

	// Calculate usage (Darwin uses different field names than Linux)
	// On Darwin: Bsize is fundamental block size, Blocks is total blocks
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bfree) * uint64(stat.Bsize)
	used := total - free

	return &DiskStats{
		Total: total,
		Used:  used,
		Free:  free,
	}, nil
}
