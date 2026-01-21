package tracing

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// CollectionTracer provides lightweight tracing for collection operations
type CollectionTracer struct {
	tracer           trace.Tracer
	provider         *sdktrace.TracerProvider
	config           config.TracingConfig
	logger           *logrus.Logger
	enabled          bool
	detailModeUntil  time.Time
	detailCollectors map[string]bool
	mu               sync.RWMutex
}

// NewCollectionTracer creates a new collection tracer
func NewCollectionTracer(cfg config.TracingConfig, logger *logrus.Logger) (*CollectionTracer, error) {
	ct := &CollectionTracer{
		config:           cfg,
		logger:           logger,
		enabled:          cfg.Enabled,
		detailCollectors: make(map[string]bool),
	}

	if !cfg.Enabled {
		ct.logger.Info("OpenTelemetry tracing disabled")
		return ct, nil
	}

	if err := ct.initializeTracer(); err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	ct.logger.WithFields(logrus.Fields{
		"sample_rate": cfg.SampleRate,
		"endpoint":    cfg.Endpoint,
	}).Info("OpenTelemetry tracing initialized")

	return ct, nil
}

// initializeTracer sets up the OpenTelemetry tracer
func (ct *CollectionTracer) initializeTracer() error {
	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("slurm-exporter"),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("component", "metric-collection"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	exporter, err := ct.createExporter()
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler based on configuration
	sampler := ct.createSampler()

	// Create tracer provider
	ct.provider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(ct.provider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Get tracer
	ct.tracer = ct.provider.Tracer("slurm-exporter")

	return nil
}

// createExporter creates the OTLP HTTP exporter
func (ct *CollectionTracer) createExporter() (sdktrace.SpanExporter, error) {
	endpoint := ct.config.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		if ct.config.Insecure {
			endpoint = fmt.Sprintf("http://%s", endpoint)
		} else {
			endpoint = fmt.Sprintf("https://%s", endpoint)
		}
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}

	if ct.config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}

// createSampler creates a sampler based on configuration
func (ct *CollectionTracer) createSampler() sdktrace.Sampler {
	if ct.config.SampleRate <= 0 {
		return sdktrace.NeverSample()
	}
	if ct.config.SampleRate >= 1.0 {
		return sdktrace.AlwaysSample()
	}

	// Use TraceIDRatioBased for probabilistic sampling
	return sdktrace.TraceIDRatioBased(ct.config.SampleRate)
}

// TraceCollection creates a span for a collection operation
func (ct *CollectionTracer) TraceCollection(ctx context.Context, collector string) (context.Context, func()) {
	if !ct.enabled || ct.tracer == nil {
		return ctx, func() {}
	}

	// Check if we should use detailed tracing for this collector
	useDetailedTracing := ct.shouldUseDetailedTracing(collector)

	spanName := fmt.Sprintf("collect.%s", collector)
	ctx, span := ct.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("collector", collector),
			attribute.String("operation", "collection"),
			attribute.Int64("timestamp", time.Now().Unix()),
			attribute.Bool("detailed", useDetailedTracing),
		),
	)

	startTime := time.Now()

	return ctx, func() {
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		)
		span.End()

		if useDetailedTracing {
			ct.logger.WithFields(logrus.Fields{
				"collector": collector,
				"duration":  duration,
				"span_id":   span.SpanContext().SpanID().String(),
				"trace_id":  span.SpanContext().TraceID().String(),
			}).Debug("Collection trace completed")
		}
	}
}

// TraceAPICall creates a span for an API call
func (ct *CollectionTracer) TraceAPICall(ctx context.Context, endpoint, method string) (context.Context, func(error)) {
	if !ct.enabled || ct.tracer == nil {
		return ctx, func(error) {}
	}

	spanName := fmt.Sprintf("api.%s", endpoint)
	ctx, span := ct.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("api.endpoint", endpoint),
			attribute.String("api.method", method),
			attribute.String("operation", "api_call"),
		),
	)

	startTime := time.Now()

	return ctx, func(err error) {
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		)

		if err != nil {
			span.SetAttributes(
				attribute.Bool("error", true),
				attribute.String("error.message", err.Error()),
			)
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.Bool("success", true))
		}

		span.End()
	}
}

// TraceMetricGeneration creates a span for metric generation
func (ct *CollectionTracer) TraceMetricGeneration(ctx context.Context, collector string, metricCount int) (context.Context, func()) {
	if !ct.enabled || ct.tracer == nil {
		return ctx, func() {}
	}

	spanName := fmt.Sprintf("metrics.%s", collector)
	ctx, span := ct.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("collector", collector),
			attribute.String("operation", "metric_generation"),
			attribute.Int("metric.count", metricCount),
		),
	)

	startTime := time.Now()

	return ctx, func() {
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		)
		span.End()
	}
}

// EnableDetailedTrace enables detailed tracing for a specific collector for a duration
func (ct *CollectionTracer) EnableDetailedTrace(collector string, duration string) error {
	if !ct.enabled {
		return fmt.Errorf("tracing is disabled")
	}

	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.detailCollectors[collector] = true
	ct.detailModeUntil = time.Now().Add(parsedDuration)

	ct.logger.WithFields(logrus.Fields{
		"collector": collector,
		"duration":  parsedDuration,
		"until":     ct.detailModeUntil,
	}).Info("Enabled detailed tracing")

	// Set up automatic cleanup
	go func() {
		time.Sleep(parsedDuration)
		ct.mu.Lock()
		delete(ct.detailCollectors, collector)
		ct.mu.Unlock()

		ct.logger.WithField("collector", collector).Info("Disabled detailed tracing")
	}()

	return nil
}

// shouldUseDetailedTracing checks if detailed tracing should be used for a collector
func (ct *CollectionTracer) shouldUseDetailedTracing(collector string) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if time.Now().After(ct.detailModeUntil) {
		return false
	}

	return ct.detailCollectors[collector]
}

// AddSpanAttribute adds an attribute to the current span if one exists
func (ct *CollectionTracer) AddSpanAttribute(ctx context.Context, key string, value interface{}) {
	if !ct.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// RecordError records an error in the current span
func (ct *CollectionTracer) RecordError(ctx context.Context, err error) {
	if !ct.enabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err)
	span.SetAttributes(
		attribute.Bool("error", true),
		attribute.String("error.message", err.Error()),
	)
}

// GetTraceID returns the trace ID from the current context
func (ct *CollectionTracer) GetTraceID(ctx context.Context) string {
	if !ct.enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID from the current context
func (ct *CollectionTracer) GetSpanID(ctx context.Context) string {
	if !ct.enabled {
		return ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// Shutdown gracefully shuts down the tracer
func (ct *CollectionTracer) Shutdown(ctx context.Context) error {
	if ct.provider == nil {
		return nil
	}

	return ct.provider.Shutdown(ctx)
}

// IsEnabled returns whether tracing is enabled
func (ct *CollectionTracer) IsEnabled() bool {
	return ct.enabled
}

// GetConfig returns the current tracing configuration
func (ct *CollectionTracer) GetConfig() config.TracingConfig {
	return ct.config
}

// TracingMiddleware provides middleware for HTTP handlers to propagate trace context
func (ct *CollectionTracer) TracingMiddleware(handler func(ctx context.Context)) func(ctx context.Context) {
	return func(ctx context.Context) {
		if !ct.enabled {
			handler(ctx)
			return
		}

		// Extract trace context if it exists
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier{})

		// Create a span for the operation
		ctx, span := ct.tracer.Start(ctx, "http.request")
		defer span.End()

		// Call the handler
		handler(ctx)
	}
}

// CreateChildSpan creates a child span with the given name and attributes
func (ct *CollectionTracer) CreateChildSpan(ctx context.Context, name string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	if !ct.enabled || ct.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return ct.tracer.Start(ctx, name, trace.WithAttributes(attributes...))
}

// SpanFromContext returns the current span from context
func (ct *CollectionTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// TracingStats provides statistics about tracing
type TracingStats struct {
	Enabled           bool              `json:"enabled"`
	SampleRate        float64           `json:"sample_rate"`
	Endpoint          string            `json:"endpoint"`
	DetailedTracing   map[string]string `json:"detailed_tracing,omitempty"`
	DetailModeUntil   time.Time         `json:"detail_mode_until,omitempty"`
	TotalSpansCreated int64             `json:"total_spans_created"`
}

// GetStats returns current tracing statistics
func (ct *CollectionTracer) GetStats() TracingStats {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	stats := TracingStats{
		Enabled:    ct.enabled,
		SampleRate: ct.config.SampleRate,
		Endpoint:   ct.config.Endpoint,
	}

	if len(ct.detailCollectors) > 0 {
		stats.DetailedTracing = make(map[string]string)
		for collector := range ct.detailCollectors {
			if time.Now().Before(ct.detailModeUntil) {
				stats.DetailedTracing[collector] = time.Until(ct.detailModeUntil).String()
			}
		}
		stats.DetailModeUntil = ct.detailModeUntil
	}

	return stats
}
