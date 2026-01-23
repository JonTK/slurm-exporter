// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProfileStorage defines the interface for storing profiles
type ProfileStorage interface {
	// Save saves a profile
	Save(profile *CollectorProfile) error

	// Load loads a profile by ID
	Load(id string) (*CollectorProfile, error)

	// List lists all profile metadata
	List() ([]*ProfileMetadata, error)

	// Delete deletes a profile
	Delete(id string) error

	// Cleanup removes old profiles based on retention policy
	Cleanup() error

	// GetStats returns storage statistics
	GetStats() map[string]interface{}
}

// ProfileMetadata represents profile metadata
type ProfileMetadata struct {
	ID            string                 `json:"id"`
	CollectorName string                 `json:"collector_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Size          int64                  `json:"size"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// FileProfileStorage stores profiles in the filesystem
type FileProfileStorage struct {
	config ProfileStorageConfig
	logger *logrus.Entry
	mu     sync.RWMutex
}

// NewProfileStorage creates a new profile storage based on config
func NewProfileStorage(config ProfileStorageConfig) (ProfileStorage, error) {
	switch config.Type {
	case "file", "":
		return NewFileProfileStorage(config, logrus.NewEntry(logrus.StandardLogger()))
	case "memory":
		return NewMemoryProfileStorage(config, logrus.NewEntry(logrus.StandardLogger()))
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// NewFileProfileStorage creates a new file-based profile storage
func NewFileProfileStorage(config ProfileStorageConfig, logger *logrus.Entry) (*FileProfileStorage, error) {
	if config.Path == "" {
		config.Path = "/tmp/slurm-exporter-profiles"
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("creating profile directory: %w", err)
	}

	return &FileProfileStorage{
		config: config,
		logger: logger,
	}, nil
}

// Save saves a profile to disk
func (fs *FileProfileStorage) Save(profile *CollectorProfile) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Generate profile ID with nanosecond precision to avoid collisions
	id := fmt.Sprintf("%s_%d", profile.CollectorName, profile.StartTime.UnixNano())

	// Create profile directory
	profileDir := filepath.Join(fs.config.Path, id)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	// Save metadata
	metadata := ProfileMetadata{
		ID:            id,
		CollectorName: profile.CollectorName,
		StartTime:     profile.StartTime,
		EndTime:       profile.EndTime,
		Duration:      profile.Duration,
		Metadata:      profile.Metadata,
	}

	// Save profile data
	var totalSize int64

	// CPU profile
	if profile.CPUProfile != nil && profile.CPUProfile.Len() > 0 {
		cpuFile := filepath.Join(profileDir, "cpu.pprof")
		size, err := fs.saveBuffer(cpuFile, profile.CPUProfile)
		if err != nil {
			return fmt.Errorf("saving CPU profile: %w", err)
		}
		totalSize += size
	}

	// Heap profile
	if profile.HeapProfile != nil && profile.HeapProfile.Len() > 0 {
		heapFile := filepath.Join(profileDir, "heap.pprof")
		size, err := fs.saveBuffer(heapFile, profile.HeapProfile)
		if err != nil {
			return fmt.Errorf("saving heap profile: %w", err)
		}
		totalSize += size
	}

	// Goroutine profile
	if profile.GoroutineProfile != nil && profile.GoroutineProfile.Len() > 0 {
		goroutineFile := filepath.Join(profileDir, "goroutine.pprof")
		size, err := fs.saveBuffer(goroutineFile, profile.GoroutineProfile)
		if err != nil {
			return fmt.Errorf("saving goroutine profile: %w", err)
		}
		totalSize += size
	}

	// Block profile
	if profile.BlockProfile != nil && profile.BlockProfile.Len() > 0 {
		blockFile := filepath.Join(profileDir, "block.pprof")
		size, err := fs.saveBuffer(blockFile, profile.BlockProfile)
		if err != nil {
			return fmt.Errorf("saving block profile: %w", err)
		}
		totalSize += size
	}

	// Mutex profile
	if profile.MutexProfile != nil && profile.MutexProfile.Len() > 0 {
		mutexFile := filepath.Join(profileDir, "mutex.pprof")
		size, err := fs.saveBuffer(mutexFile, profile.MutexProfile)
		if err != nil {
			return fmt.Errorf("saving mutex profile: %w", err)
		}
		totalSize += size
	}

	// Trace data
	if profile.TraceData != nil && profile.TraceData.Len() > 0 {
		traceFile := filepath.Join(profileDir, "trace.out")
		size, err := fs.saveBuffer(traceFile, profile.TraceData)
		if err != nil {
			return fmt.Errorf("saving trace data: %w", err)
		}
		totalSize += size
	}

	metadata.Size = totalSize

	// Check storage limits
	if err := fs.checkStorageLimits(); err != nil {
		return err
	}

	fs.logger.WithFields(logrus.Fields{
		"profile_id": id,
		"collector":  profile.CollectorName,
		"size":       totalSize,
		"duration":   profile.Duration,
	}).Info("Profile saved")

	return nil
}

// saveBuffer saves a buffer to a file
func (fs *FileProfileStorage) saveBuffer(filename string, buf io.Reader) (int64, error) {
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	return io.Copy(file, buf)
}

// Load loads a profile from disk
func (fs *FileProfileStorage) Load(id string) (*CollectorProfile, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	profileDir := filepath.Join(fs.config.Path, id)
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile not found: %s", id)
	}

	profile := &CollectorProfile{
		Phases:   make(map[string]*ProfilePhase),
		Metadata: make(map[string]interface{}),
	}

	// Load CPU profile
	cpuFile := filepath.Join(profileDir, "cpu.pprof")
	if buf, err := fs.loadFile(cpuFile); err == nil {
		profile.CPUProfile = buf
	}

	// Load heap profile
	heapFile := filepath.Join(profileDir, "heap.pprof")
	if buf, err := fs.loadFile(heapFile); err == nil {
		profile.HeapProfile = buf
	}

	// Load goroutine profile
	goroutineFile := filepath.Join(profileDir, "goroutine.pprof")
	if buf, err := fs.loadFile(goroutineFile); err == nil {
		profile.GoroutineProfile = buf
	}

	// Load block profile
	blockFile := filepath.Join(profileDir, "block.pprof")
	if buf, err := fs.loadFile(blockFile); err == nil {
		profile.BlockProfile = buf
	}

	// Load mutex profile
	mutexFile := filepath.Join(profileDir, "mutex.pprof")
	if buf, err := fs.loadFile(mutexFile); err == nil {
		profile.MutexProfile = buf
	}

	// Load trace data
	traceFile := filepath.Join(profileDir, "trace.out")
	if buf, err := fs.loadFile(traceFile); err == nil {
		profile.TraceData = buf
	}

	return profile, nil
}

// loadFile loads a file into a buffer
func (fs *FileProfileStorage) loadFile(filename string) (*bytes.Buffer, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}

// List lists all profile metadata
func (fs *FileProfileStorage) List() ([]*ProfileMetadata, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	entries, err := os.ReadDir(fs.config.Path)
	if err != nil {
		return nil, fmt.Errorf("reading profile directory: %w", err)
	}

	var profiles []*ProfileMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse profile ID
		id := entry.Name()
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Get profile size
		profileDir := filepath.Join(fs.config.Path, id)
		size, _ := fs.getDirectorySize(profileDir)

		metadata := &ProfileMetadata{
			ID:        id,
			StartTime: info.ModTime(),
			Size:      size,
		}

		profiles = append(profiles, metadata)
	}

	// Sort by start time (newest first)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].StartTime.After(profiles[j].StartTime)
	})

	return profiles, nil
}

// Delete deletes a profile
func (fs *FileProfileStorage) Delete(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	profileDir := filepath.Join(fs.config.Path, id)
	return os.RemoveAll(profileDir)
}

// Cleanup removes old profiles based on retention policy
func (fs *FileProfileStorage) Cleanup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.config.Retention <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-fs.config.Retention)

	entries, err := os.ReadDir(fs.config.Path)
	if err != nil {
		return fmt.Errorf("reading profile directory: %w", err)
	}

	var deleted int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			profileDir := filepath.Join(fs.config.Path, entry.Name())
			if err := os.RemoveAll(profileDir); err != nil {
				fs.logger.WithError(err).WithField("profile", entry.Name()).Error("Failed to delete old profile")
			} else {
				deleted++
			}
		}
	}

	if deleted > 0 {
		fs.logger.WithField("count", deleted).Info("Cleaned up old profiles")
	}

	return nil
}

// checkStorageLimits checks and enforces storage limits
func (fs *FileProfileStorage) checkStorageLimits() error {
	if fs.config.MaxSize <= 0 {
		return nil
	}

	totalSize, err := fs.getDirectorySize(fs.config.Path)
	if err != nil {
		return err
	}

	if totalSize <= fs.config.MaxSize {
		return nil
	}

	// Need to delete oldest profiles
	profiles, err := fs.List()
	if err != nil {
		return err
	}

	// Sort by age (oldest first)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].StartTime.Before(profiles[j].StartTime)
	})

	// Delete oldest profiles until under limit
	for _, profile := range profiles {
		if err := fs.Delete(profile.ID); err != nil {
			fs.logger.WithError(err).WithField("profile", profile.ID).Error("Failed to delete profile")
			continue
		}

		totalSize -= profile.Size
		if totalSize <= fs.config.MaxSize {
			break
		}
	}

	return nil
}

// getDirectorySize calculates the total size of a directory
func (fs *FileProfileStorage) getDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// GetStats returns storage statistics
func (fs *FileProfileStorage) GetStats() map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	profiles, _ := fs.List()
	totalSize, _ := fs.getDirectorySize(fs.config.Path)

	return map[string]interface{}{
		"type":          "file",
		"path":          fs.config.Path,
		"profile_count": len(profiles),
		"total_size":    totalSize,
		"max_size":      fs.config.MaxSize,
		"retention":     fs.config.Retention,
	}
}

// MemoryProfileStorage stores profiles in memory
type MemoryProfileStorage struct {
	config   ProfileStorageConfig
	logger   *logrus.Entry
	mu       sync.RWMutex
	profiles map[string]*CollectorProfile
	metadata map[string]*ProfileMetadata
}

// NewMemoryProfileStorage creates a new memory-based profile storage
func NewMemoryProfileStorage(config ProfileStorageConfig, logger *logrus.Entry) (*MemoryProfileStorage, error) {
	return &MemoryProfileStorage{
		config:   config,
		logger:   logger,
		profiles: make(map[string]*CollectorProfile),
		metadata: make(map[string]*ProfileMetadata),
	}, nil
}

// Save saves a profile to memory
func (ms *MemoryProfileStorage) Save(profile *CollectorProfile) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Generate profile ID with nanosecond precision to avoid collisions
	id := fmt.Sprintf("%s_%d", profile.CollectorName, profile.StartTime.UnixNano())

	// Calculate size
	var size int64
	if profile.CPUProfile != nil {
		size += int64(profile.CPUProfile.Len())
	}
	if profile.HeapProfile != nil {
		size += int64(profile.HeapProfile.Len())
	}
	if profile.GoroutineProfile != nil {
		size += int64(profile.GoroutineProfile.Len())
	}
	if profile.BlockProfile != nil {
		size += int64(profile.BlockProfile.Len())
	}
	if profile.MutexProfile != nil {
		size += int64(profile.MutexProfile.Len())
	}
	if profile.TraceData != nil {
		size += int64(profile.TraceData.Len())
	}

	metadata := &ProfileMetadata{
		ID:            id,
		CollectorName: profile.CollectorName,
		StartTime:     profile.StartTime,
		EndTime:       profile.EndTime,
		Duration:      profile.Duration,
		Size:          size,
		Metadata:      profile.Metadata,
	}

	ms.profiles[id] = profile
	ms.metadata[id] = metadata

	// Enforce limits
	ms.enforceLimits()

	return nil
}

// Load loads a profile from memory
func (ms *MemoryProfileStorage) Load(id string) (*CollectorProfile, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	profile, exists := ms.profiles[id]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", id)
	}

	return profile, nil
}

// List lists all profile metadata
func (ms *MemoryProfileStorage) List() ([]*ProfileMetadata, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	profiles := make([]*ProfileMetadata, 0, len(ms.metadata))
	for _, metadata := range ms.metadata {
		profiles = append(profiles, metadata)
	}

	// Sort by start time (newest first)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].StartTime.After(profiles[j].StartTime)
	})

	return profiles, nil
}

// Delete deletes a profile
func (ms *MemoryProfileStorage) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.profiles, id)
	delete(ms.metadata, id)
	return nil
}

// Cleanup removes old profiles based on retention policy
func (ms *MemoryProfileStorage) Cleanup() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.config.Retention <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-ms.config.Retention)

	for id, metadata := range ms.metadata {
		if metadata.StartTime.Before(cutoff) {
			delete(ms.profiles, id)
			delete(ms.metadata, id)
		}
	}

	return nil
}

// enforceLimits enforces storage limits
// NOTE: This method assumes the caller already holds ms.mu write lock
func (ms *MemoryProfileStorage) enforceLimits() {
	// Enforce count limit
	maxProfiles := 100 // Default max profiles in memory
	if len(ms.profiles) <= maxProfiles {
		return
	}

	// Build sorted list of profiles without calling List() to avoid lock reentry
	profiles := make([]*ProfileMetadata, 0, len(ms.metadata))
	for _, metadata := range ms.metadata {
		profiles = append(profiles, metadata)
	}

	// Sort by start time (oldest first for deletion)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].StartTime.Before(profiles[j].StartTime)
	})

	// Delete oldest profiles
	for i := 0; i < len(profiles)-maxProfiles; i++ {
		delete(ms.profiles, profiles[i].ID)
		delete(ms.metadata, profiles[i].ID)
	}
}

// GetStats returns storage statistics
func (ms *MemoryProfileStorage) GetStats() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var totalSize int64
	for _, metadata := range ms.metadata {
		totalSize += metadata.Size
	}

	return map[string]interface{}{
		"type":          "memory",
		"profile_count": len(ms.profiles),
		"total_size":    totalSize,
	}
}
