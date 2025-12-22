package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQoSManager for testing
type MockQoSManager struct {
	mock.Mock
}

func (m *MockQoSManager) List(ctx context.Context, opts *slurm.ListQoSOptions) (*slurm.QoSList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.QoSList), args.Error(1)
}

func (m *MockQoSManager) Get(ctx context.Context, name string) (*slurm.QoS, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.QoS), args.Error(1)
}

func TestQoSCollector_Collect(t *testing.T) {
	tests := []struct {
		name     string
		qosList  *slurm.QoSList
		qosErr   error
		wantErr  bool
		validate func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "successful collection",
			qosList: &slurm.QoSList{
				QoS: []slurm.QoS{
					{
						Name:                "normal",
						Description:         "Normal QoS",
						Priority:            100,
						UsageFactor:         1.0,
						MaxJobs:             1000,
						MaxJobsPerUser:      100,
						MaxJobsPerAccount:   500,
						MaxSubmitJobs:       2000,
						MaxCPUs:             5000,
						MaxCPUsPerUser:      500,
						MaxNodes:            100,
						MaxWallTime:         1440, // 24 hours in minutes
						MinCPUs:             1,
						MinNodes:            1,
						PreemptMode:         []string{"CANCEL"},
						Flags:               []string{"DenyOnLimit"},
					},
					{
						Name:                "high",
						Description:         "High Priority QoS",
						Priority:            1000,
						UsageFactor:         2.0,
						MaxJobs:             2000000, // Infinite
						MaxJobsPerUser:      -1,
						MaxJobsPerAccount:   -1,
						MaxSubmitJobs:       -1,
						MaxCPUs:             -1,
						MaxCPUsPerUser:      -1,
						MaxNodes:            -1,
						MaxWallTime:         -1,
						MinCPUs:             4,
						MinNodes:            2,
						PreemptMode:         []string{"REQUEUE", "CANCEL"},
						Flags:               []string{"NoDecay", "UsageFactorSafe"},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// We should have metrics for each QoS
				assert.Greater(t, len(metrics), 20) // At least metrics for 2 QoS entries

				// Check some specific metrics
				foundNormalPriority := false
				foundHighUsageFactor := false
				foundNormalInfo := false

				for _, m := range metrics {
					dto := &prometheus.Metric{}
					m.Write(dto)

					if dto.Gauge != nil {
						for _, label := range dto.Label {
							if *label.Name == "qos" && *label.Value == "normal" {
								if m.Desc().String() == prometheus.BuildFQName(namespace, qosCollectorSubsystem, "priority") {
									assert.Equal(t, float64(100), *dto.Gauge.Value)
									foundNormalPriority = true
								}
							}
							if *label.Name == "qos" && *label.Value == "high" {
								if m.Desc().String() == prometheus.BuildFQName(namespace, qosCollectorSubsystem, "usage_factor") {
									assert.Equal(t, float64(2.0), *dto.Gauge.Value)
									foundHighUsageFactor = true
								}
							}
						}

						// Check info metric
						labelCount := len(dto.Label)
						if labelCount == 4 { // info metric has 4 labels
							for _, label := range dto.Label {
								if *label.Name == "qos" && *label.Value == "normal" {
									foundNormalInfo = true
								}
							}
						}
					}
				}

				assert.True(t, foundNormalPriority, "Should find normal QoS priority metric")
				assert.True(t, foundHighUsageFactor, "Should find high QoS usage factor metric")
				assert.True(t, foundNormalInfo, "Should find normal QoS info metric")
			},
		},
		{
			name:    "error listing QoS",
			qosList: nil,
			qosErr:  assert.AnError,
			wantErr: true,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// Should only have error metrics
				assert.Less(t, len(metrics), 5)
			},
		},
		{
			name: "empty QoS list",
			qosList: &slurm.QoSList{
				QoS: []slurm.QoS{},
			},
			wantErr: false,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// Should only have base metrics (no QoS-specific metrics)
				assert.Less(t, len(metrics), 5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client and QoS manager
			mockClient := &MockClient{}
			mockQoSManager := &MockQoSManager{}

			// Set up expectations
			mockClient.On("QoS").Return(mockQoSManager)
			mockQoSManager.On("List", mock.Anything, mock.Anything).Return(tt.qosList, tt.qosErr)

			// Create collector
			logger := logrus.NewEntry(logrus.New())
			collector := NewQoSCollector(mockClient, logger)

			// Collect metrics
			ch := make(chan prometheus.Metric, 100)
			go func() {
				collector.Collect(ch)
				close(ch)
			}()

			// Gather metrics
			var metrics []prometheus.Metric
			for metric := range ch {
				metrics = append(metrics, metric)
			}

			// Validate
			if tt.validate != nil {
				tt.validate(t, metrics)
			}

			// Verify expectations
			mockClient.AssertExpectations(t)
			mockQoSManager.AssertExpectations(t)
		})
	}
}

func TestQoSCollector_sendMetricIfNotInfinite(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	collector := NewQoSCollector(nil, logger)

	tests := []struct {
		name          string
		value         int
		expectedValue float64
	}{
		{"zero value", 0, 0},
		{"normal value", 100, 100},
		{"large but not infinite", 999999, 999999},
		{"infinite value", 1000000, -1},
		{"very large value", 2000000, -1},
		{"negative value", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan prometheus.Metric, 1)

			collector.sendMetricIfNotInfinite(ch, collector.qosMaxJobs, "test_qos", tt.value)

			metric := <-ch
			dto := &prometheus.Metric{}
			metric.Write(dto)

			assert.Equal(t, tt.expectedValue, *dto.Gauge.Value)
		})
	}
}

func TestQoSCollector_WallTimeConversion(t *testing.T) {
	// Create mock client and QoS manager
	mockClient := &MockClient{}
	mockQoSManager := &MockQoSManager{}

	qosList := &slurm.QoSList{
		QoS: []slurm.QoS{
			{
				Name:        "short",
				MaxWallTime: 60, // 1 hour in minutes
			},
			{
				Name:        "long",
				MaxWallTime: 525600, // 1 year in minutes (365*24*60)
			},
			{
				Name:        "unlimited",
				MaxWallTime: -1,
			},
		},
	}

	mockClient.On("QoS").Return(mockQoSManager)
	mockQoSManager.On("List", mock.Anything, mock.Anything).Return(qosList, nil)

	// Create collector
	logger := logrus.NewEntry(logrus.New())
	collector := NewQoSCollector(mockClient, logger)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Check wall time conversions
	wallTimes := make(map[string]float64)
	for metric := range ch {
		dto := &prometheus.Metric{}
		metric.Write(dto)

		if dto.Gauge != nil && len(dto.Label) == 1 {
			if strings.Contains(metric.Desc().String(), "max_wall_time_seconds") {
				wallTimes[*dto.Label[0].Value] = *dto.Gauge.Value
			}
		}
	}

	assert.Equal(t, float64(3600), wallTimes["short"], "1 hour should be 3600 seconds")
	assert.Equal(t, float64(-1), wallTimes["long"], "1 year should be treated as infinite")
	// unlimited (-1) may not appear in metrics as it's filtered out
}