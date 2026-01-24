// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfiler(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	t.Run("NewProfiler", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled:          true,
			CPUProfileRate:   100,
			MemProfileRate:   512 * 1024,
			BlockProfileRate: 1,
			AutoProfile: AutoProfileConfig{
				Enabled:                 true,
				DurationThreshold:       100 * time.Millisecond,
				MemoryThreshold:         1024 * 1024, // 1MB
				ProfileOnSlowCollection: true,
			},
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, profiler)
		assert.True(t, profiler.enabled)
	})

	t.Run("DisabledProfiler", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: false,
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, profiler)
		assert.False(t, profiler.enabled)
	})

	t.Run("StartOperation", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		op := profiler.StartOperation("test_collector")
		assert.NotNil(t, op)
		assert.Equal(t, "test_collector", op.collectorName)
		assert.NotNil(t, op.profile)

		// Check that profile is tracked
		profile := profiler.GetProfile("test_collector")
		assert.NotNil(t, profile)
		assert.Equal(t, "test_collector", profile.CollectorName)

		op.Stop()
	})

	t.Run("Phases", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		op := profiler.StartOperation("test_collector")

		// Start phases
		op.Phase("initialization")
		time.Sleep(10 * time.Millisecond)
		op.PhaseEnd("initialization")

		op.Phase("data_fetch")
		time.Sleep(20 * time.Millisecond)
		op.PhaseEnd("data_fetch")

		op.Phase("processing")
		time.Sleep(15 * time.Millisecond)
		op.PhaseEnd("processing")

		op.Stop()

		// Check phases
		assert.Len(t, op.profile.Phases, 3)

		initPhase := op.profile.Phases["initialization"]
		assert.NotNil(t, initPhase)
		assert.True(t, initPhase.Duration >= 10*time.Millisecond)

		fetchPhase := op.profile.Phases["data_fetch"]
		assert.NotNil(t, fetchPhase)
		assert.True(t, fetchPhase.Duration >= 20*time.Millisecond)

		procPhase := op.profile.Phases["processing"]
		assert.NotNil(t, procPhase)
		assert.True(t, procPhase.Duration >= 15*time.Millisecond)
	})

	t.Run("AutoProfile", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			AutoProfile: AutoProfileConfig{
				Enabled:           true,
				DurationThreshold: 50 * time.Millisecond,
				MemoryThreshold:   1024, // 1KB
			},
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		// Test duration threshold
		op1 := profiler.StartOperation("slow_collector")
		time.Sleep(60 * time.Millisecond)
		op1.Stop()

		// Test memory threshold
		op2 := profiler.StartOperation("memory_collector")
		// Allocate memory
		data := make([]byte, 2048)
		_ = data
		runtime.GC()
		op2.Stop()

		// Both should trigger auto-save
		time.Sleep(10 * time.Millisecond) // Allow async saves
	})

	t.Run("ProfileCollection", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled:              true,
			CPUProfileRate:       100,
			BlockProfileRate:     1,
			MutexProfileFraction: 1,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		op := profiler.StartOperation("full_collector")

		// Do some work to generate profiles
		ch := make(chan int)
		go func() {
			time.Sleep(10 * time.Millisecond)
			ch <- 1
		}()
		<-ch

		// Collect various profiles
		op.SaveHeapProfile()
		op.Stop()

		// Check that profiles were collected
		assert.NotNil(t, op.profile.HeapProfile)
		assert.True(t, op.profile.HeapProfile.Len() > 0)

		if op.profile.GoroutineProfile != nil {
			assert.True(t, op.profile.GoroutineProfile.Len() > 0)
		}
	})

	t.Run("Report", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		op := profiler.StartOperation("report_collector")
		op.Phase("phase1")
		time.Sleep(10 * time.Millisecond)
		op.PhaseEnd("phase1")
		op.Stop()

		report := op.GenerateReport()
		assert.Contains(t, report, "report_collector")
		assert.Contains(t, report, "phase1")
		assert.Contains(t, report, "Duration:")
	})

	t.Run("ListProfiles", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		// Create multiple profiles
		for i := 0; i < 3; i++ {
			op := profiler.StartOperation("list_test")
			time.Sleep(5 * time.Millisecond)
			op.Stop()
			_ = op.Save()
		}

		profiles, err := profiler.ListProfiles()
		require.NoError(t, err)
		assert.True(t, len(profiles) >= 3)
	})

	t.Run("Metrics", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		reg := prometheus.NewRegistry()
		err = profiler.RegisterMetrics(reg)
		require.NoError(t, err)

		op := profiler.StartOperation("metrics_test")
		time.Sleep(10 * time.Millisecond)
		op.Stop()

		// Gather metrics
		mfs, err := reg.Gather()
		require.NoError(t, err)
		assert.True(t, len(mfs) > 0)
	})

	t.Run("Stats", func(t *testing.T) {
		config := ProfilerConfig{
			Enabled: true,
			Storage: ProfileStorageConfig{
				Type: "memory",
			},
		}

		profiler, err := NewProfiler(config, logger)
		require.NoError(t, err)

		stats := profiler.GetStats()
		enabled, ok := stats["enabled"].(bool)
		require.True(t, ok, "enabled should be bool")
		assert.True(t, enabled)
		assert.NotNil(t, stats["config"])
		assert.NotNil(t, stats["storage"])
	})
}

func TestFileProfileStorage(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	tmpDir := t.TempDir()

	config := ProfileStorageConfig{
		Type:      "file",
		Path:      tmpDir,
		MaxSize:   10 * 1024 * 1024, // 10MB
		Retention: time.Hour,
	}

	storage, err := NewFileProfileStorage(config, logger)
	require.NoError(t, err)

	t.Run("SaveAndLoad", func(t *testing.T) {
		profile := &CollectorProfile{
			CollectorName: "test_collector",
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Minute),
			Duration:      time.Minute,
			HeapProfile:   &bytes.Buffer{},
			Metadata: map[string]interface{}{
				"test": "value",
			},
		}

		// Write some data to heap profile
		profile.HeapProfile.WriteString("test heap profile data")

		err := storage.Save(profile)
		require.NoError(t, err)

		// List profiles
		profiles, err := storage.List()
		require.NoError(t, err)
		assert.Len(t, profiles, 1)

		// Load profile
		id := profiles[0].ID
		loaded, err := storage.Load(id)
		require.NoError(t, err)
		assert.NotNil(t, loaded.HeapProfile)
	})

	t.Run("Delete", func(t *testing.T) {
		profile := &CollectorProfile{
			CollectorName: "delete_test",
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Second),
			Duration:      time.Second,
		}

		err := storage.Save(profile)
		require.NoError(t, err)

		profiles, err := storage.List()
		require.NoError(t, err)

		var deleteID string
		for _, p := range profiles {
			if p.CollectorName == "delete_test" {
				deleteID = p.ID
				break
			}
		}

		err = storage.Delete(deleteID)
		require.NoError(t, err)

		_, err = storage.Load(deleteID)
		assert.Error(t, err)
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Create an old profile
		oldProfile := &CollectorProfile{
			CollectorName: "old_test",
			StartTime:     time.Now().Add(-2 * time.Hour),
			EndTime:       time.Now().Add(-2 * time.Hour).Add(time.Second),
			Duration:      time.Second,
		}

		err := storage.Save(oldProfile)
		require.NoError(t, err)

		// Change retention to 1 hour
		storage.config.Retention = time.Hour

		err = storage.Cleanup()
		require.NoError(t, err)

		// Old profile should be deleted
		profiles, err := storage.List()
		require.NoError(t, err)

		for _, p := range profiles {
			assert.NotEqual(t, "old_test", p.CollectorName)
		}
	})

	t.Run("Stats", func(t *testing.T) {
		stats := storage.GetStats()
		assert.Equal(t, "file", stats["type"])
		assert.Equal(t, tmpDir, stats["path"])
		assert.NotNil(t, stats["profile_count"])
		assert.NotNil(t, stats["total_size"])
	})
}

func TestMemoryProfileStorage(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	config := ProfileStorageConfig{
		Type:      "memory",
		Retention: time.Hour,
	}

	storage, err := NewMemoryProfileStorage(config, logger)
	require.NoError(t, err)

	t.Run("SaveAndLoad", func(t *testing.T) {
		profile := &CollectorProfile{
			CollectorName: "test_collector",
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Minute),
			Duration:      time.Minute,
			CPUProfile:    &bytes.Buffer{},
			Metadata: map[string]interface{}{
				"test": "value",
			},
		}

		profile.CPUProfile.WriteString("test cpu profile data")

		err := storage.Save(profile)
		require.NoError(t, err)

		// List profiles
		profiles, err := storage.List()
		require.NoError(t, err)
		assert.True(t, len(profiles) >= 1)

		// Load profile
		id := fmt.Sprintf("%s_%d", profile.CollectorName, profile.StartTime.UnixNano())
		loaded, err := storage.Load(id)
		require.NoError(t, err)
		assert.NotNil(t, loaded)
		assert.Equal(t, profile.CollectorName, loaded.CollectorName)
	})

	t.Run("EnforceLimits", func(t *testing.T) {
		// Create many profiles to exceed limit

		for i := 0; i < 150; i++ {
			profile := &CollectorProfile{
				CollectorName: fmt.Sprintf("test_%d", i),
				StartTime:     time.Now().Add(time.Duration(-i) * time.Minute),
				EndTime:       time.Now().Add(time.Duration(-i) * time.Minute).Add(time.Second),
				Duration:      time.Second,
			}
			_ = storage.Save(profile)
		}

		profiles, err := storage.List()
		require.NoError(t, err)
		assert.LessOrEqual(t, len(profiles), 100) // Default max
	})

	t.Run("Cleanup", func(t *testing.T) {
		// Create an old profile
		oldProfile := &CollectorProfile{
			CollectorName: "old_memory_test",
			StartTime:     time.Now().Add(-2 * time.Hour),
			EndTime:       time.Now().Add(-2 * time.Hour).Add(time.Second),
			Duration:      time.Second,
		}

		err := storage.Save(oldProfile)
		require.NoError(t, err)

		err = storage.Cleanup()
		require.NoError(t, err)

		// Old profile should be deleted
		profiles, err := storage.List()
		require.NoError(t, err)

		found := false
		for _, p := range profiles {
			if p.CollectorName == "old_memory_test" {
				found = true
				break
			}
		}
		assert.False(t, found)
	})

	t.Run("Stats", func(t *testing.T) {
		stats := storage.GetStats()
		assert.Equal(t, "memory", stats["type"])
		assert.NotNil(t, stats["profile_count"])
		assert.NotNil(t, stats["total_size"])
	})
}
