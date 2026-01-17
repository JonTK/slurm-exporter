package benchmark

import (
	"runtime"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
)

// BenchmarkRegistryCreation tests registry creation performance
func BenchmarkRegistryCreation(b *testing.B) {
	cfg := &config.CollectorsConfig{
		Jobs:       config.CollectorConfig{Enabled: true},
		Nodes:      config.CollectorConfig{Enabled: true},
		Partitions: config.CollectorConfig{Enabled: true},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		promRegistry := prometheus.NewRegistry()
		registry, err := collector.NewRegistry(cfg, promRegistry)
		if err != nil {
			b.Fatalf("Failed to create registry: %v", err)
		}
		_ = registry // Use the registry to avoid optimization
	}
}

// BenchmarkMetricCreation tests basic metric creation
func BenchmarkMetricCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		gauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "test_metric",
				Help: "Test metric for benchmarking",
			},
			[]string{"label1", "label2"},
		)
		gauge.WithLabelValues("value1", "value2").Set(1.0)
	}
}

// BenchmarkMemoryUsage tests memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	cfg := &config.CollectorsConfig{
		Jobs:       config.CollectorConfig{Enabled: true},
		Nodes:      config.CollectorConfig{Enabled: true},
		Partitions: config.CollectorConfig{Enabled: true},
	}
	promRegistry := prometheus.NewRegistry()

	registry, err := collector.NewRegistry(cfg, promRegistry)
	if err != nil {
		b.Fatalf("Failed to create registry: %v", err)
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		registry.Collect(ch)
		close(ch)

		// Drain the channel
		for range ch {
		}
	}

	b.StopTimer()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}
