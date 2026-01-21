package tracing

import (
	"context"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ExampleCollectorWithTracing demonstrates how to integrate tracing into a collector
type ExampleCollectorWithTracing struct {
	name   string
	tracer *CollectionTracer
	client ExampleSLURMClient
	logger *logrus.Logger
}

// ExampleSLURMClient represents a simple SLURM client interface for the example
type ExampleSLURMClient interface {
	ListJobs(ctx context.Context) ([]ExampleJob, error)
	GetJobDetails(ctx context.Context, jobID string) (*ExampleJobDetails, error)
}

// ExampleJob represents a simple job structure
type ExampleJob struct {
	ID    string
	Name  string
	State string
}

// ExampleJobDetails represents detailed job information
type ExampleJobDetails struct {
	ID      string
	CPUs    int
	Memory  int64
	Runtime time.Duration
}

// NewExampleCollectorWithTracing creates a new example collector with tracing
func NewExampleCollectorWithTracing(
	name string,
	tracer *CollectionTracer,
	client ExampleSLURMClient,
	logger *logrus.Logger,
) *ExampleCollectorWithTracing {
	return &ExampleCollectorWithTracing{
		name:   name,
		tracer: tracer,
		client: client,
		logger: logger,
	}
}

// Collect performs the collection with comprehensive tracing
func (c *ExampleCollectorWithTracing) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Start collection tracing
	ctx, finishCollection := c.tracer.TraceCollection(ctx, c.name)
	defer finishCollection()

	c.logger.WithFields(logrus.Fields{
		"collector": c.name,
		"trace_id":  c.tracer.GetTraceID(ctx),
		"span_id":   c.tracer.GetSpanID(ctx),
	}).Debug("Starting collection")

	// Phase 1: List Jobs
	ctx, finishListJobs := c.tracer.TraceAPICall(ctx, "jobs", "LIST")
	jobs, err := c.client.ListJobs(ctx)
	finishListJobs(err)

	if err != nil {
		c.tracer.RecordError(ctx, err)
		return err
	}

	c.tracer.AddSpanAttribute(ctx, "jobs.count", len(jobs))
	c.logger.WithField("job_count", len(jobs)).Debug("Listed jobs")

	// Phase 2: Process each job with individual tracing
	for i, job := range jobs {
		// Create child span for job processing
		jobCtx, jobSpan := c.tracer.CreateChildSpan(ctx, "process_job")

		c.tracer.AddSpanAttribute(jobCtx, "job.id", job.ID)
		c.tracer.AddSpanAttribute(jobCtx, "job.name", job.Name)
		c.tracer.AddSpanAttribute(jobCtx, "job.state", job.State)
		c.tracer.AddSpanAttribute(jobCtx, "job.index", i)

		// Get detailed job information
		jobCtx, finishGetDetails := c.tracer.TraceAPICall(jobCtx, "job_details", "GET")
		details, err := c.client.GetJobDetails(jobCtx, job.ID)
		finishGetDetails(err)

		if err != nil {
			c.tracer.RecordError(jobCtx, err)
			c.logger.WithFields(logrus.Fields{
				"job_id": job.ID,
				"error":  err,
			}).Warn("Failed to get job details")
			jobSpan.End()
			continue
		}

		// Add detailed attributes
		c.tracer.AddSpanAttribute(jobCtx, "job.cpus", details.CPUs)
		c.tracer.AddSpanAttribute(jobCtx, "job.memory_mb", details.Memory/1024/1024)
		c.tracer.AddSpanAttribute(jobCtx, "job.runtime_seconds", int64(details.Runtime.Seconds()))

		// Generate metrics (traced)
		c.generateJobMetrics(jobCtx, job, details, ch)

		jobSpan.End()
	}

	// Phase 3: Generate summary metrics
	ctx, finishMetrics := c.tracer.TraceMetricGeneration(ctx, c.name, len(jobs)*3) // Assume 3 metrics per job
	c.generateSummaryMetrics(ctx, jobs, ch)
	finishMetrics()

	c.logger.WithFields(logrus.Fields{
		"collector":      c.name,
		"jobs_processed": len(jobs),
		"trace_id":       c.tracer.GetTraceID(ctx),
	}).Info("Collection completed")

	return nil
}

// generateJobMetrics generates metrics for a specific job with tracing
func (c *ExampleCollectorWithTracing) generateJobMetrics(
	ctx context.Context,
	job ExampleJob,
	details *ExampleJobDetails,
	ch chan<- prometheus.Metric,
) {
	// Create child span for metric generation
	ctx, span := c.tracer.CreateChildSpan(ctx, "generate_job_metrics")
	defer span.End()

	c.tracer.AddSpanAttribute(ctx, "metrics.job_id", job.ID)

	// Example metrics (in real implementation, these would be actual Prometheus metrics)
	c.logger.WithFields(logrus.Fields{
		"job_id":    job.ID,
		"cpus":      details.CPUs,
		"memory_mb": details.Memory / 1024 / 1024,
		"runtime":   details.Runtime,
		"trace_id":  c.tracer.GetTraceID(ctx),
	}).Debug("Generated job metrics")

	c.tracer.AddSpanAttribute(ctx, "metrics.generated", 3)
}

// generateSummaryMetrics generates summary metrics with tracing
func (c *ExampleCollectorWithTracing) generateSummaryMetrics(
	ctx context.Context,
	jobs []ExampleJob,
	ch chan<- prometheus.Metric,
) {
	// Create child span for summary metrics
	ctx, span := c.tracer.CreateChildSpan(ctx, "generate_summary_metrics")
	defer span.End()

	// Count jobs by state
	stateCount := make(map[string]int)
	for _, job := range jobs {
		stateCount[job.State]++
	}

	c.tracer.AddSpanAttribute(ctx, "summary.total_jobs", len(jobs))
	c.tracer.AddSpanAttribute(ctx, "summary.unique_states", len(stateCount))

	for state, count := range stateCount {
		c.tracer.AddSpanAttribute(ctx, "summary.state_"+state, count)
	}

	c.logger.WithFields(logrus.Fields{
		"total_jobs":  len(jobs),
		"state_count": stateCount,
		"trace_id":    c.tracer.GetTraceID(ctx),
	}).Debug("Generated summary metrics")
}

// Name returns the collector name
func (c *ExampleCollectorWithTracing) Name() string {
	return c.name
}

// ExampleInitialization shows how to initialize tracing in an application
func ExampleInitialization() (*CollectionTracer, error) {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create tracing configuration
	tracingConfig := config.TracingConfig{
		Enabled:    true,
		SampleRate: 0.1,              // Sample 10% of traces
		Endpoint:   "localhost:4318", // OTLP HTTP endpoint
		Insecure:   true,
	}

	// Initialize tracer
	tracer, err := NewCollectionTracer(tracingConfig, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("Tracing initialized successfully")
	return tracer, nil
}

// ExampleUsageInMainApplication demonstrates how to use tracing in the main application
func ExampleUsageInMainApplication() {
	// Initialize tracing
	tracer, err := ExampleInitialization()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize tracing")
	}
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	// Create example client (in real code, this would be the actual SLURM client)
	client := &MockSLURMClient{}

	// Create collector with tracing
	collector := NewExampleCollectorWithTracing("jobs", tracer, client, logrus.New())

	// Create metrics channel
	metricsCh := make(chan prometheus.Metric, 100)

	// Perform collection with tracing
	ctx := context.Background()
	if err := collector.Collect(ctx, metricsCh); err != nil {
		logrus.WithError(err).Error("Collection failed")
	}

	// Enable detailed tracing for debugging
	if err := tracer.EnableDetailedTrace("jobs", "5m"); err != nil {
		logrus.WithError(err).Warn("Failed to enable detailed tracing")
	}

	// Get tracing statistics
	stats := tracer.GetStats()
	logrus.WithFields(logrus.Fields{
		"enabled":          stats.Enabled,
		"sample_rate":      stats.SampleRate,
		"detailed_tracing": stats.DetailedTracing,
	}).Info("Tracing statistics")
}

// MockSLURMClient provides a mock implementation for the example
type MockSLURMClient struct{}

func (m *MockSLURMClient) ListJobs(ctx context.Context) ([]ExampleJob, error) {
	// Simulate API delay
	time.Sleep(10 * time.Millisecond)

	return []ExampleJob{
		{ID: "123", Name: "job1", State: "RUNNING"},
		{ID: "124", Name: "job2", State: "PENDING"},
		{ID: "125", Name: "job3", State: "COMPLETED"},
	}, nil
}

func (m *MockSLURMClient) GetJobDetails(ctx context.Context, jobID string) (*ExampleJobDetails, error) {
	// Simulate API delay
	time.Sleep(5 * time.Millisecond)

	return &ExampleJobDetails{
		ID:      jobID,
		CPUs:    4,
		Memory:  8 * 1024 * 1024 * 1024, // 8GB
		Runtime: 2 * time.Hour,
	}, nil
}
