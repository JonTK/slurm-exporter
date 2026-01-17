package tracing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestNewCollectionTracer_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise

	cfg := config.TracingConfig{
		Enabled: false,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.False(t, tracer.IsEnabled())
	assert.Nil(t, tracer.tracer)
	assert.Nil(t, tracer.provider)
}

func TestNewCollectionTracer_Enabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0, // Sample everything for testing
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.True(t, tracer.IsEnabled())
	assert.NotNil(t, tracer.tracer)
	assert.NotNil(t, tracer.provider)

	// Clean up
	err = tracer.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestCollectionTracer_TraceCollection_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{Enabled: false}
	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	newCtx, finish := tracer.TraceCollection(ctx, "test_collector")

	// Should return the same context when disabled
	assert.Equal(t, ctx, newCtx)
	assert.NotNil(t, finish)

	// Finish should not panic
	finish()
}

func TestCollectionTracer_TraceCollection_Enabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()
	newCtx, finish := tracer.TraceCollection(ctx, "test_collector")

	// Should get a new context with span
	assert.NotEqual(t, ctx, newCtx)
	assert.NotNil(t, finish)

	// Should have a valid span
	span := trace.SpanFromContext(newCtx)
	assert.True(t, span.SpanContext().IsValid())

	// Finish should not panic
	finish()
}

func TestCollectionTracer_TraceAPICall_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{Enabled: false}
	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	newCtx, finish := tracer.TraceAPICall(ctx, "jobs", "GET")

	assert.Equal(t, ctx, newCtx)
	assert.NotNil(t, finish)

	// Finish should not panic
	finish(nil)
	finish(errors.New("test error"))
}

func TestCollectionTracer_TraceAPICall_Enabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Test successful API call
	newCtx, finish := tracer.TraceAPICall(ctx, "jobs", "GET")
	assert.NotEqual(t, ctx, newCtx)

	span := trace.SpanFromContext(newCtx)
	assert.True(t, span.SpanContext().IsValid())

	finish(nil) // Success

	// Test failed API call
	_, finish = tracer.TraceAPICall(ctx, "jobs", "GET")
	finish(errors.New("connection failed")) // Error
}

func TestCollectionTracer_TraceMetricGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()
	newCtx, finish := tracer.TraceMetricGeneration(ctx, "test_collector", 42)

	assert.NotEqual(t, ctx, newCtx)
	span := trace.SpanFromContext(newCtx)
	assert.True(t, span.SpanContext().IsValid())

	finish()
}

func TestCollectionTracer_EnableDetailedTrace(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	// Test enabling detailed trace
	err = tracer.EnableDetailedTrace("test_collector", "100ms")
	assert.NoError(t, err)

	// Should be using detailed tracing
	assert.True(t, tracer.shouldUseDetailedTracing("test_collector"))
	assert.False(t, tracer.shouldUseDetailedTracing("other_collector"))

	// Wait for it to expire
	time.Sleep(150 * time.Millisecond)
	assert.False(t, tracer.shouldUseDetailedTracing("test_collector"))

	// Test invalid duration
	err = tracer.EnableDetailedTrace("test_collector", "invalid")
	assert.Error(t, err)
}

func TestCollectionTracer_EnableDetailedTrace_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{Enabled: false}
	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)

	err = tracer.EnableDetailedTrace("test_collector", "1s")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tracing is disabled")
}

func TestCollectionTracer_AddSpanAttribute(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()
	ctx, finish := tracer.TraceCollection(ctx, "test_collector")
	defer finish()

	// Test various attribute types
	tracer.AddSpanAttribute(ctx, "string_attr", "test")
	tracer.AddSpanAttribute(ctx, "int_attr", 42)
	tracer.AddSpanAttribute(ctx, "int64_attr", int64(123))
	tracer.AddSpanAttribute(ctx, "float64_attr", 3.14)
	tracer.AddSpanAttribute(ctx, "bool_attr", true)
	tracer.AddSpanAttribute(ctx, "other_attr", []string{"test"}) // Should convert to string

	// Should not panic
}

func TestCollectionTracer_AddSpanAttribute_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{Enabled: false}
	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	tracer.AddSpanAttribute(ctx, "test", "value") // Should not panic
}

func TestCollectionTracer_RecordError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()
	ctx, finish := tracer.TraceCollection(ctx, "test_collector")
	defer finish()

	// Test recording error
	testErr := errors.New("test error")
	tracer.RecordError(ctx, testErr)

	// Test nil error (should not panic)
	tracer.RecordError(ctx, nil)
}

func TestCollectionTracer_GetTraceAndSpanID(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()

	// Without span
	assert.Empty(t, tracer.GetTraceID(ctx))
	assert.Empty(t, tracer.GetSpanID(ctx))

	// With span
	ctx, finish := tracer.TraceCollection(ctx, "test_collector")
	defer finish()

	traceID := tracer.GetTraceID(ctx)
	spanID := tracer.GetSpanID(ctx)

	assert.NotEmpty(t, traceID)
	assert.NotEmpty(t, spanID)
	assert.Len(t, traceID, 32) // Trace ID should be 32 hex characters
	assert.Len(t, spanID, 16)  // Span ID should be 16 hex characters
}

func TestCollectionTracer_GetTraceAndSpanID_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{Enabled: false}
	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	assert.Empty(t, tracer.GetTraceID(ctx))
	assert.Empty(t, tracer.GetSpanID(ctx))
}

func TestCollectionTracer_CreateChildSpan(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 1.0,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	ctx := context.Background()
	ctx, parentFinish := tracer.TraceCollection(ctx, "test_collector")
	defer parentFinish()

	// Create child span
	childCtx, childSpan := tracer.CreateChildSpan(ctx, "child_operation",
		attribute.String("test", "value"))
	defer childSpan.End()

	assert.NotEqual(t, ctx, childCtx)
	assert.True(t, childSpan.SpanContext().IsValid())

	// Child should have same trace ID but different span ID
	parentSpan := tracer.SpanFromContext(ctx)
	assert.Equal(t, parentSpan.SpanContext().TraceID(), childSpan.SpanContext().TraceID())
	assert.NotEqual(t, parentSpan.SpanContext().SpanID(), childSpan.SpanContext().SpanID())
}

func TestCollectionTracer_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 0.5,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	stats := tracer.GetStats()
	assert.True(t, stats.Enabled)
	assert.Equal(t, 0.5, stats.SampleRate)
	assert.Equal(t, "localhost:4317", stats.Endpoint)
	assert.Empty(t, stats.DetailedTracing)

	// Enable detailed tracing
	err = tracer.EnableDetailedTrace("test_collector", "1s")
	require.NoError(t, err)

	stats = tracer.GetStats()
	assert.NotEmpty(t, stats.DetailedTracing)
	assert.Contains(t, stats.DetailedTracing, "test_collector")
	assert.False(t, stats.DetailModeUntil.IsZero())
}

func TestCollectionTracer_GetConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 0.25,
		Endpoint:   "test-endpoint",
		Insecure:   false,
	}

	tracer, err := NewCollectionTracer(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = tracer.Shutdown(context.Background()) }()

	returnedCfg := tracer.GetConfig()
	assert.Equal(t, cfg.Enabled, returnedCfg.Enabled)
	assert.Equal(t, cfg.SampleRate, returnedCfg.SampleRate)
	assert.Equal(t, cfg.Endpoint, returnedCfg.Endpoint)
	assert.Equal(t, cfg.Insecure, returnedCfg.Insecure)
}
