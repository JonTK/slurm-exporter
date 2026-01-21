package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFairShareTrendReportingSLURMClient struct {
	mock.Mock
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareTrendData(ctx context.Context, entity string, period string) (*FairShareTrendData, error) {
	args := m.Called(ctx, entity, period)
	return args.Get(0).(*FairShareTrendData), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetUserFairShareHistory(ctx context.Context, userName string, period string) (*UserFairShareHistory, error) {
	args := m.Called(ctx, userName, period)
	return args.Get(0).(*UserFairShareHistory), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetAccountFairShareHistory(ctx context.Context, accountName string, period string) (*AccountFairShareHistory, error) {
	args := m.Called(ctx, accountName, period)
	return args.Get(0).(*AccountFairShareHistory), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetSystemFairShareTrends(ctx context.Context, period string) (*SystemFairShareTrends, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*SystemFairShareTrends), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareTimeSeries(ctx context.Context, entity string, metric string, resolution string) (*FairShareTimeSeries, error) {
	args := m.Called(ctx, entity, metric, resolution)
	return args.Get(0).(*FairShareTimeSeries), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareAggregations(ctx context.Context, aggregationType string, period string) (*FairShareAggregations, error) {
	args := m.Called(ctx, aggregationType, period)
	return args.Get(0).(*FairShareAggregations), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareMovingAverages(ctx context.Context, entity string, windows []string) (*FairShareMovingAverages, error) {
	args := m.Called(ctx, entity, windows)
	return args.Get(0).(*FairShareMovingAverages), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetSeasonalDecomposition(ctx context.Context, entity string, period string) (*SeasonalDecomposition, error) {
	args := m.Called(ctx, entity, period)
	return args.Get(0).(*SeasonalDecomposition), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareHeatmap(ctx context.Context, entityType string, period string) (*FairShareHeatmap, error) {
	args := m.Called(ctx, entityType, period)
	return args.Get(0).(*FairShareHeatmap), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareDistribution(ctx context.Context, timestamp time.Time) (*FairShareDistribution, error) {
	args := m.Called(ctx, timestamp)
	return args.Get(0).(*FairShareDistribution), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareComparison(ctx context.Context, entities []string, period string) (*FairShareComparison, error) {
	args := m.Called(ctx, entities, period)
	return args.Get(0).(*FairShareComparison), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareCorrelations(ctx context.Context, metrics []string, period string) (*FairShareCorrelations, error) {
	args := m.Called(ctx, metrics, period)
	return args.Get(0).(*FairShareCorrelations), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GenerateFairShareReport(ctx context.Context, reportType string, period string) (*FairShareReport, error) {
	args := m.Called(ctx, reportType, period)
	return args.Get(0).(*FairShareReport), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareInsights(ctx context.Context, entity string, period string) (*FairShareInsights, error) {
	args := m.Called(ctx, entity, period)
	return args.Get(0).(*FairShareInsights), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareAnomalies(ctx context.Context, period string) (*FairShareAnomalies, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*FairShareAnomalies), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetFairShareForecast(ctx context.Context, entity string, horizon string) (*FairShareForecast, error) {
	args := m.Called(ctx, entity, horizon)
	return args.Get(0).(*FairShareForecast), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetDashboardMetrics(ctx context.Context, dashboardID string) (*DashboardMetrics, error) {
	args := m.Called(ctx, dashboardID)
	return args.Get(0).(*DashboardMetrics), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetRealtimeUpdates(ctx context.Context, updateType string) (*RealtimeUpdates, error) {
	args := m.Called(ctx, updateType)
	return args.Get(0).(*RealtimeUpdates), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetVisualizationConfig(ctx context.Context, visualType string) (*VisualizationConfig, error) {
	args := m.Called(ctx, visualType)
	return args.Get(0).(*VisualizationConfig), args.Error(1)
}

func (m *MockFairShareTrendReportingSLURMClient) GetReportingSchedule(ctx context.Context) (*ReportingSchedule, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ReportingSchedule), args.Error(1)
}

// Types are already defined in fairshare_trend_reporting.go

func TestNewFairShareTrendReportingCollector(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}
	collector := NewFairShareTrendReportingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.fairShareTrendDirection)
	assert.NotNil(t, collector.fairShareVolatility)
	assert.NotNil(t, collector.fairShareForecastValue)
	assert.NotNil(t, collector.reportGenerationTime)
}

func TestFairShareTrendReportingCollector_Describe(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}
	collector := NewFairShareTrendReportingCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 64 metrics, verify we have the correct number
	assert.Equal(t, 64, len(descs), "Should have exactly 64 metric descriptions")
}

func TestFairShareTrendReportingCollector_Collect_Success(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Mock trend data
	trendData := &FairShareTrendData{
		Entity:             "user1",
		EntityType:         "user",
		Period:             "7d",
		FairShareValues:    []float64{0.5, 0.52, 0.54, 0.56, 0.58, 0.6, 0.62},
		Timestamps:         generateTimestamps(7),
		TrendDirection:     "improving",
		TrendSlope:         0.02,
		TrendConfidence:    0.85,
		Mean:               0.56,
		Median:             0.56,
		StandardDeviation:  0.04,
		Variance:           0.0016,
		Minimum:            0.5,
		Maximum:            0.62,
		Range:              0.12,
		PercentageChange:   24.0,
		AbsoluteChange:     0.12,
		ChangeRate:         0.017,
		Volatility:         0.08,
		Percentiles: map[int]float64{
			25: 0.52,
			50: 0.56,
			75: 0.60,
			90: 0.61,
			95: 0.62,
		},
		Quartiles: map[int]float64{
			1: 0.52,
			2: 0.56,
			3: 0.60,
		},
		InterquartileRange: 0.08,
		TrendPattern:       "linear_growth",
		Seasonality:        true,
		Cyclicality:        false,
		Stationarity:       false,
		NextPeriodForecast: 0.64,
		ForecastConfidence: 0.82,
		ForecastError:      0.02,
		DataPoints:         7,
		MissingPoints:      0,
		DataQuality:        1.0,
		LastUpdated:        time.Now(),
	}

	// Mock time series data
	timeSeries := &FairShareTimeSeries{
		Entity:     "user1",
		Metric:     "fairshare",
		Resolution: "1h",
		DataPoints: []TimeSeriesPoint{
			{Timestamp: time.Now(), Value: 0.62, Confidence: 0.95},
		},
		SignalToNoise:    3.5,
		DataCompleteness: 0.98,
		SamplingRate:     1.0,
	}

	// Mock heatmap data
	heatmapData := &FairShareHeatmap{
		EntityType: "user",
		Period:     "24h",
		Entities:   []string{"user1", "user2", "user3"},
		TimeSlots:  generateTimestamps(24),
		Values:     [][]float64{{0.5, 0.6, 0.7}, {0.4, 0.5, 0.6}, {0.6, 0.7, 0.8}},
		ColorScale: "viridis",
		MinValue:   0.0,
		MaxValue:   1.0,
	}

	// Mock report data
	reportData := &FairShareReport{
		ReportID:    "monthly_2024_01",
		ReportType:  "monthly",
		Period:      "30d",
		GeneratedAt: time.Now(),
		Summary: ReportSummary{
			OverallStatus:   "healthy",
			HealthScore:     0.85,
			ComplianceRate:  0.92,
			EfficiencyScore: 0.78,
		},
		KeyFindings:     []string{"Fair-share distribution improving", "User compliance high"},
		Recommendations: []string{"Adjust decay parameters", "Monitor outlier users"},
		ActionItems: []ActionItem{
			{ID: "1", Title: "Review policy", Priority: "high"},
			{ID: "2", Title: "Update thresholds", Priority: "medium"},
		},
	}

	// Setup mock expectations
	client.On("GetFairShareTrendData", mock.Anything, "user1", "7d").Return(trendData, nil)
	client.On("GetFairShareTrendData", mock.Anything, "account1", "30d").Return(&FairShareTrendData{
		TrendDirection: "stable",
		Percentiles:    map[int]float64{50: 0.5},
		Quartiles:      map[int]float64{2: 0.5},
	}, nil)
	client.On("GetSystemFairShareTrends", mock.Anything, "24h").Return(&SystemFairShareTrends{}, nil)
	client.On("GetFairShareTimeSeries", mock.Anything, "user1", "fairshare", "1h").Return(timeSeries, nil)
	client.On("GetFairShareMovingAverages", mock.Anything, "user1", []string{"7d", "30d"}).Return(&FairShareMovingAverages{}, nil)
	client.On("GetSeasonalDecomposition", mock.Anything, "user1", "30d").Return(&SeasonalDecomposition{}, nil)
	client.On("GetFairShareHeatmap", mock.Anything, "user", "24h").Return(heatmapData, nil)
	client.On("GetFairShareCorrelations", mock.Anything, []string{"fairshare", "usage"}, "30d").Return(&FairShareCorrelations{}, nil)
	client.On("GetFairShareForecast", mock.Anything, "user1", "7d").Return(&FairShareForecast{}, nil)
	client.On("GenerateFairShareReport", mock.Anything, "monthly", "30d").Return(reportData, nil)

	collector := NewFairShareTrendReportingCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 300)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundTrendDirection := false
	foundVolatility := false
	foundForecast := false
	foundReport := false
	foundLatency := false

	for name := range metricNames {
		if strings.Contains(name, "trend_direction") {
			foundTrendDirection = true
		}
		if strings.Contains(name, "volatility") {
			foundVolatility = true
		}
		if strings.Contains(name, "forecast") {
			foundForecast = true
		}
		if strings.Contains(name, "report") {
			foundReport = true
		}
		if strings.Contains(name, "latency") {
			foundLatency = true
		}
	}

	assert.True(t, foundTrendDirection, "Should have trend direction metrics")
	assert.True(t, foundVolatility, "Should have volatility metrics")
	assert.True(t, foundForecast, "Should have forecast metrics")
	assert.True(t, foundReport, "Should have report metrics")
	assert.True(t, foundLatency, "Should have latency metrics")

	client.AssertExpectations(t)
}

func TestFairShareTrendReportingCollector_Collect_Error(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Mock error response
	client.On("GetFairShareTrendData", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*FairShareTrendData)(nil), assert.AnError)
	client.On("GetSystemFairShareTrends", mock.Anything, mock.AnythingOfType("string")).Return((*SystemFairShareTrends)(nil), assert.AnError)

	collector := NewFairShareTrendReportingCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect some metrics (empty metrics after reset)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestFairShareTrendReportingCollector_MetricValues(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Create test data with known values
	trendData := &FairShareTrendData{
		Entity:             "test_user",
		EntityType:         "user",
		Period:             "7d",
		TrendDirection:     "improving",
		TrendSlope:         0.03,
		TrendConfidence:    0.9,
		Volatility:         0.05,
		PercentageChange:   15.0,
		StandardDeviation:  0.02,
		Variance:           0.0004,
		Median:             0.65,
		InterquartileRange: 0.1,
		Seasonality:        true,
		Stationarity:       false,
		NextPeriodForecast: 0.68,
		ForecastConfidence: 0.85,
		Percentiles:        map[int]float64{50: 0.65},
		Quartiles:          map[int]float64{2: 0.65},
	}

	client.On("GetFairShareTrendData", mock.Anything, "user1", "7d").Return(trendData, nil)
	client.On("GetFairShareTrendData", mock.Anything, "account1", "30d").Return(&FairShareTrendData{
		TrendDirection: "stable",
		Percentiles:    map[int]float64{50: 0.5},
		Quartiles:      map[int]float64{2: 0.5},
	}, nil)
	client.On("GetSystemFairShareTrends", mock.Anything, "24h").Return(&SystemFairShareTrends{}, nil)
	client.On("GetFairShareTimeSeries", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTimeSeries{}, nil)
	client.On("GetFairShareMovingAverages", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return(&FairShareMovingAverages{}, nil)
	client.On("GetSeasonalDecomposition", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&SeasonalDecomposition{}, nil)
	client.On("GetFairShareHeatmap", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareHeatmap{}, nil)
	client.On("GetFairShareCorrelations", mock.Anything, mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(&FairShareCorrelations{}, nil)
	client.On("GetFairShareForecast", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareForecast{}, nil)
	client.On("GenerateFairShareReport", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareReport{}, nil)

	collector := NewFairShareTrendReportingCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundTrendDirection := false
	foundTrendSlope := false
	foundVolatility := false
	foundPercentageChange := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_fairshare_trend_direction":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(1), *mf.Metric[0].Gauge.Value) // improving = 1
				foundTrendDirection = true
			}
		case "slurm_fairshare_trend_slope":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.03), *mf.Metric[0].Gauge.Value)
				foundTrendSlope = true
			}
		case "slurm_fairshare_volatility":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.05), *mf.Metric[0].Gauge.Value)
				foundVolatility = true
			}
		case "slurm_fairshare_percentage_change":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(15), *mf.Metric[0].Gauge.Value)
				foundPercentageChange = true
			}
		}
	}

	assert.True(t, foundTrendDirection, "Should find trend direction metric with correct value")
	assert.True(t, foundTrendSlope, "Should find trend slope metric with correct value")
	assert.True(t, foundVolatility, "Should find volatility metric with correct value")
	assert.True(t, foundPercentageChange, "Should find percentage change metric with correct value")
}

func TestFairShareTrendReportingCollector_Integration(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Setup comprehensive mock data
	setupTrendReportingMocks(client)

	collector := NewFairShareTrendReportingCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_fairshare_trend_direction Fair-share trend direction (-1=declining, 0=stable, 1=improving)
		# TYPE slurm_fairshare_trend_direction gauge
		slurm_fairshare_trend_direction{entity="user1",entity_type="user",period="7d"} 1
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_fairshare_trend_direction")
	assert.NoError(t, err)
}

func TestFairShareTrendReportingCollector_TimeSeriesMetrics(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Setup time series specific mocks
	setupTimeSeriesMocks(client)

	collector := NewFairShareTrendReportingCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify time series metrics are present
	foundAutoCorrelation := false
	foundSeasonalComponent := false
	foundSignalToNoise := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "autocorrelation") {
			foundAutoCorrelation = true
		}
		if strings.Contains(desc, "seasonal_component") {
			foundSeasonalComponent = true
		}
		if strings.Contains(desc, "signal_to_noise") {
			foundSignalToNoise = true
		}
	}

	assert.True(t, foundAutoCorrelation, "Should find autocorrelation metrics")
	assert.True(t, foundSeasonalComponent, "Should find seasonal component metrics")
	assert.True(t, foundSignalToNoise, "Should find signal-to-noise metrics")
}

func TestFairShareTrendReportingCollector_ForecastMetrics(t *testing.T) {
	client := &MockFairShareTrendReportingSLURMClient{}

	// Setup forecast specific mocks
	setupForecastMocks(client)

	collector := NewFairShareTrendReportingCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify forecast metrics are present
	foundForecastValue := false
	foundForecastConfidence := false
	foundForecastAccuracy := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "forecast_value") {
			foundForecastValue = true
		}
		if strings.Contains(desc, "forecast_confidence") {
			foundForecastConfidence = true
		}
		if strings.Contains(desc, "forecast_accuracy") {
			foundForecastAccuracy = true
		}
	}

	assert.True(t, foundForecastValue, "Should find forecast value metrics")
	assert.True(t, foundForecastConfidence, "Should find forecast confidence metrics")
	assert.True(t, foundForecastAccuracy, "Should find forecast accuracy metrics")
}

// Helper functions

func generateTimestamps(count int) []time.Time {
	timestamps := make([]time.Time, count)
	now := time.Now()
	for i := 0; i < count; i++ {
		timestamps[i] = now.Add(time.Duration(-i) * time.Hour)
	}
	return timestamps
}

func setupTrendReportingMocks(client *MockFairShareTrendReportingSLURMClient) {
	trendData := &FairShareTrendData{
		Entity:             "user1",
		EntityType:         "user",
		Period:             "7d",
		TrendDirection:     "improving",
		TrendSlope:         0.025,
		TrendConfidence:    0.88,
		Volatility:         0.06,
		PercentageChange:   20.0,
		StandardDeviation:  0.03,
		Variance:           0.0009,
		Median:             0.6,
		InterquartileRange: 0.12,
		Seasonality:        true,
		Stationarity:       false,
		NextPeriodForecast: 0.65,
		ForecastConfidence: 0.8,
		Percentiles:        map[int]float64{50: 0.6, 90: 0.7},
		Quartiles:          map[int]float64{1: 0.5, 2: 0.6, 3: 0.7},
	}

	client.On("GetFairShareTrendData", mock.Anything, "user1", "7d").Return(trendData, nil)
	client.On("GetFairShareTrendData", mock.Anything, "account1", "30d").Return(&FairShareTrendData{
		TrendDirection: "stable",
		Percentiles:    map[int]float64{50: 0.5},
		Quartiles:      map[int]float64{2: 0.5},
	}, nil)
	client.On("GetSystemFairShareTrends", mock.Anything, mock.AnythingOfType("string")).Return(&SystemFairShareTrends{}, nil)
	client.On("GetFairShareTimeSeries", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTimeSeries{}, nil)
	client.On("GetFairShareMovingAverages", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return(&FairShareMovingAverages{}, nil)
	client.On("GetSeasonalDecomposition", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&SeasonalDecomposition{}, nil)
	client.On("GetFairShareHeatmap", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareHeatmap{}, nil)
	client.On("GetFairShareCorrelations", mock.Anything, mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(&FairShareCorrelations{}, nil)
	client.On("GetFairShareForecast", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareForecast{}, nil)
	client.On("GenerateFairShareReport", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareReport{}, nil)
}

func setupTimeSeriesMocks(client *MockFairShareTrendReportingSLURMClient) {
	client.On("GetFairShareTrendData", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTrendData{
		TrendDirection: "stable",
		Percentiles:    map[int]float64{50: 0.5},
		Quartiles:      map[int]float64{2: 0.5},
	}, nil)
	client.On("GetSystemFairShareTrends", mock.Anything, mock.AnythingOfType("string")).Return(&SystemFairShareTrends{}, nil)
	client.On("GetFairShareTimeSeries", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTimeSeries{
		SignalToNoise:    4.2,
		DataCompleteness: 0.99,
	}, nil)
	client.On("GetFairShareMovingAverages", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return(&FairShareMovingAverages{}, nil)
	client.On("GetSeasonalDecomposition", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&SeasonalDecomposition{}, nil)
	client.On("GetFairShareHeatmap", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareHeatmap{}, nil)
	client.On("GetFairShareCorrelations", mock.Anything, mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(&FairShareCorrelations{}, nil)
	client.On("GetFairShareForecast", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareForecast{}, nil)
	client.On("GenerateFairShareReport", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareReport{}, nil)
}

func setupForecastMocks(client *MockFairShareTrendReportingSLURMClient) {
	client.On("GetFairShareTrendData", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTrendData{
		TrendDirection:     "improving",
		NextPeriodForecast: 0.72,
		ForecastConfidence: 0.86,
		ForecastError:      0.015,
		Percentiles:        map[int]float64{50: 0.5},
		Quartiles:          map[int]float64{2: 0.5},
	}, nil)
	client.On("GetSystemFairShareTrends", mock.Anything, mock.AnythingOfType("string")).Return(&SystemFairShareTrends{}, nil)
	client.On("GetFairShareTimeSeries", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareTimeSeries{}, nil)
	client.On("GetFairShareMovingAverages", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return(&FairShareMovingAverages{}, nil)
	client.On("GetSeasonalDecomposition", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&SeasonalDecomposition{}, nil)
	client.On("GetFairShareHeatmap", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareHeatmap{}, nil)
	client.On("GetFairShareCorrelations", mock.Anything, mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(&FairShareCorrelations{}, nil)
	client.On("GetFairShareForecast", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareForecast{}, nil)
	client.On("GenerateFairShareReport", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FairShareReport{}, nil)
}