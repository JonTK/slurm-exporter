//go:build ignore
// +build ignore

// TODO: This test file is excluded from builds due to compilation errors
// Mock implementations don't match current interfaces or use outdated types.

package collector

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/performance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCollector implements the Collector interface for testing
type mockCollector struct {
	name        string
	enabled     bool
	collectFunc func(ctx context.Context, ch chan<- prometheus.Metric) error
}

func (m *mockCollector) Name() string {
	return m.name
}

func (m *mockCollector) Describe(ch chan<- *prometheus.Desc) {
	// Mock implementation
}

func (m *mockCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if m.collectFunc != nil {
		return m.collectFunc(ctx, ch)
	}
	return nil
}

func (m *mockCollector) IsEnabled() bool {
	return m.enabled
}

func (m *mockCollector) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func TestProfiledCollector(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	
	profilerConfig := performance.ProfilerConfig{
		Enabled: true,
		Storage: performance.ProfileStorageConfig{
			Type: "memory",
		},
		AutoProfile: performance.AutoProfileConfig{
			Enabled:           true,
			DurationThreshold: 50 * time.Millisecond,
		},
	}

	profiler, err := performance.NewProfiler(profilerConfig, logger)
	require.NoError(t, err)

	t.Run("NewProfiledCollector", func(t *testing.T) {
		mock := &mockCollector{
			name:    "test_collector",
			enabled: true,
		}

		pc, err := NewProfiledCollector(mock, profiler, logger)
		require.NoError(t, err)
		assert.NotNil(t, pc)
		assert.Equal(t, "test_collector", pc.Name())
		assert.True(t, pc.IsEnabled())
	})

	t.Run("NilCollector", func(t *testing.T) {
		_, err := NewProfiledCollector(nil, profiler, logger)
		assert.Error(t, err)
	})

	t.Run("Collect", func(t *testing.T) {
		collectCalled := false
		mock := &mockCollector{
			name:    "test_collector",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				collectCalled = true
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}

		pc, err := NewProfiledCollector(mock, profiler, logger)
		require.NoError(t, err)

		ch := make(chan prometheus.Metric)
		go func() {
			for range ch {
				// Drain channel
			}
		}()

		err = pc.Collect(context.Background(), ch)
		require.NoError(t, err)
		assert.True(t, collectCalled)

		// Check that a profile was created
		profile := profiler.GetProfile("test_collector")
		assert.Nil(t, profile) // Profile is removed after collection
	})

	t.Run("CollectWithError", func(t *testing.T) {
		testErr := fmt.Errorf("test error")
		mock := &mockCollector{
			name:    "error_collector",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return testErr
			},
		}

		pc, err := NewProfiledCollector(mock, profiler, logger)
		require.NoError(t, err)

		ch := make(chan prometheus.Metric)
		err = pc.Collect(context.Background(), ch)
		assert.Equal(t, testErr, err)
	})

	t.Run("SlowCollection", func(t *testing.T) {
		mock := &mockCollector{
			name:    "slow_collector",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				time.Sleep(60 * time.Millisecond) // Trigger auto-profile
				return nil
			},
		}

		pc, err := NewProfiledCollector(mock, profiler, logger)
		require.NoError(t, err)

		ch := make(chan prometheus.Metric)
		err = pc.Collect(context.Background(), ch)
		require.NoError(t, err)

		// Should have triggered auto-save due to duration threshold
		time.Sleep(10 * time.Millisecond) // Allow async save
	})

	t.Run("ProfilingDisabled", func(t *testing.T) {
		mock := &mockCollector{
			name:    "test_collector",
			enabled: true,
		}

		pc, err := NewProfiledCollector(mock, profiler, logger)
		require.NoError(t, err)

		// Disable profiling
		pc.SetProfilingEnabled(false)

		ch := make(chan prometheus.Metric)
		err = pc.Collect(context.Background(), ch)
		require.NoError(t, err)

		// No profile should be created
		profile := profiler.GetProfile("test_collector")
		assert.Nil(t, profile)
	})
}

func TestProfiledCollectorManager(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	
	profilerConfig := performance.ProfilerConfig{
		Enabled: true,
		Storage: performance.ProfileStorageConfig{
			Type: "memory",
		},
	}

	profiler, err := performance.NewProfiler(profilerConfig, logger)
	require.NoError(t, err)

	pcm := NewProfiledCollectorManager(profiler, logger)

	t.Run("WrapCollector", func(t *testing.T) {
		mock := &mockCollector{
			name:    "test_collector",
			enabled: true,
		}

		wrapped, err := pcm.WrapCollector(mock)
		require.NoError(t, err)
		assert.NotNil(t, wrapped)
		assert.Equal(t, "test_collector", wrapped.Name())

		// Wrap again should return same instance
		wrapped2, err := pcm.WrapCollector(mock)
		require.NoError(t, err)
		assert.Equal(t, wrapped, wrapped2)
	})

	t.Run("SetProfilingEnabled", func(t *testing.T) {
		mock := &mockCollector{
			name:    "toggle_collector",
			enabled: true,
		}

		_, err := pcm.WrapCollector(mock)
		require.NoError(t, err)

		// Disable profiling
		err = pcm.SetProfilingEnabled("toggle_collector", false)
		require.NoError(t, err)

		// Try non-existent collector
		err = pcm.SetProfilingEnabled("non_existent", false)
		assert.Error(t, err)
	})

	t.Run("SetProfilingEnabledAll", func(t *testing.T) {
		// Wrap multiple collectors
		for i := 0; i < 3; i++ {
			mock := &mockCollector{
				name:    fmt.Sprintf("collector_%d", i),
				enabled: true,
			}
			_, err := pcm.WrapCollector(mock)
			require.NoError(t, err)
		}

		// Disable all
		pcm.SetProfilingEnabledAll(false)

		// Enable all
		pcm.SetProfilingEnabledAll(true)
	})

	t.Run("GetCollectorProfiles", func(t *testing.T) {
		mock := &mockCollector{
			name:    "profile_test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}

		wrapped, err := pcm.WrapCollector(mock)
		require.NoError(t, err)

		// Perform collection to generate profile
		ch := make(chan prometheus.Metric)
		go func() {
			for range ch {
			}
		}()

		pc := wrapped.(*ProfiledCollector)
		err = pc.Collect(context.Background(), ch)
		require.NoError(t, err)

		// Save the profile
		op := profiler.StartOperation("profile_test")
		op.Stop()
		op.Save()

		// Get profiles
		profiles, err := pcm.GetCollectorProfiles("profile_test")
		require.NoError(t, err)
		assert.True(t, len(profiles) >= 1)
	})

	t.Run("GetAllProfiles", func(t *testing.T) {
		allProfiles, err := pcm.GetAllProfiles()
		require.NoError(t, err)
		assert.NotNil(t, allProfiles)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := pcm.GetStats()
		assert.NotNil(t, stats["total_collectors"])
		assert.NotNil(t, stats["collectors"])
		assert.NotNil(t, stats["profiler_stats"])
	})
}