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

// MockReservationsManager for testing
type MockReservationsManager struct {
	mock.Mock
}

func (m *MockReservationsManager) List(ctx context.Context, opts *slurm.ListReservationsOptions) (*slurm.ReservationList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.ReservationList), args.Error(1)
}

func (m *MockReservationsManager) Get(ctx context.Context, name string) (*slurm.Reservation, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.Reservation), args.Error(1)
}

func TestReservationCollector_Collect(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(2 * time.Hour)
	pastTime := now.Add(-1 * time.Hour)

	tests := []struct {
		name            string
		reservationList *slurm.ReservationList
		reservationErr  error
		wantErr         bool
		validate        func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "successful collection",
			reservationList: &slurm.ReservationList{
				Reservations: []slurm.Reservation{
					{
						Name:           "gpu_reservation",
						State:          "ACTIVE",
						PartitionName:  "gpu",
						NodeCount:      4,
						CoreCount:      128,
						StartTime:      pastTime,
						EndTime:        futureTime,
						Duration:       180, // 3 hours in minutes
						Users:          []string{"user1", "user2"},
						Accounts:       []string{"research"},
						Flags:          []string{"MAINT", "SPEC_NODES"},
					},
					{
						Name:           "expired_reservation",
						State:          "EXPIRED",
						PartitionName:  "compute",
						NodeCount:      2,
						CoreCount:      64,
						StartTime:      pastTime.Add(-2 * time.Hour),
						EndTime:        pastTime,
						Duration:       60, // 1 hour in minutes
						Users:          []string{"user3"},
						Accounts:       []string{"engineering"},
						Flags:          []string{},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// We should have metrics for each reservation
				assert.Greater(t, len(metrics), 14) // At least 7 metrics per reservation

				// Check specific metrics
				foundActiveState := false
				foundExpiredState := false
				foundGpuNodes := false
				foundRemaining := false

				for _, m := range metrics {
					dto := &prometheus.Metric{}
					m.Write(dto)

					if dto.Gauge != nil {
						labels := make(map[string]string)
						for _, label := range dto.Label {
							labels[*label.Name] = *label.Value
						}

						// Check state metrics
						if labels["reservation"] == "gpu_reservation" {
							if strings.Contains(m.Desc().String(), "state") {
								assert.Equal(t, float64(1), *dto.Gauge.Value)
								foundActiveState = true
							}
							if strings.Contains(m.Desc().String(), "remaining_seconds") {
								assert.Greater(t, *dto.Gauge.Value, float64(0))
								foundRemaining = true
							}
						}

						if labels["reservation"] == "expired_reservation" {
							if strings.Contains(m.Desc().String(), "state") {
								assert.Equal(t, float64(0), *dto.Gauge.Value)
								foundExpiredState = true
							}
						}

						// Check node count
						if labels["reservation"] == "gpu_reservation" && labels["partition"] == "gpu" {
							if strings.Contains(m.Desc().String(), "nodes") {
								assert.Equal(t, float64(4), *dto.Gauge.Value)
								foundGpuNodes = true
							}
						}
					}
				}

				assert.True(t, foundActiveState, "Should find active reservation state")
				assert.True(t, foundExpiredState, "Should find expired reservation state")
				assert.True(t, foundGpuNodes, "Should find GPU reservation nodes")
				assert.True(t, foundRemaining, "Should find remaining time for active reservation")
			},
		},
		{
			name:            "error listing reservations",
			reservationList: nil,
			reservationErr:  assert.AnError,
			wantErr:         true,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// Should only have error metrics
				assert.Less(t, len(metrics), 5)
			},
		},
		{
			name: "empty reservation list",
			reservationList: &slurm.ReservationList{
				Reservations: []slurm.Reservation{},
			},
			wantErr: false,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// Should only have base metrics (no reservation-specific metrics)
				assert.Less(t, len(metrics), 5)
			},
		},
		{
			name: "reservation with multiple flags and users",
			reservationList: &slurm.ReservationList{
				Reservations: []slurm.Reservation{
					{
						Name:           "multi_user_res",
						State:          "ACTIVE",
						PartitionName:  "compute",
						NodeCount:      10,
						CoreCount:      320,
						StartTime:      pastTime,
						EndTime:        futureTime,
						Duration:       120,
						Users:          []string{"alice", "bob", "charlie"},
						Accounts:       []string{"dept1", "dept2"},
						Flags:          []string{"DAILY", "WEEKDAY", "FLEX"},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, metrics []prometheus.Metric) {
				// Check info metric has correct labels
				foundInfo := false
				for _, m := range metrics {
					dto := &prometheus.Metric{}
					m.Write(dto)

					if dto.Gauge != nil && len(dto.Label) == 6 { // info metric has 6 labels
						labels := make(map[string]string)
						for _, label := range dto.Label {
							labels[*label.Name] = *label.Value
						}

						if labels["reservation"] == "multi_user_res" {
							assert.Equal(t, "alice,bob,charlie", labels["users"])
							assert.Equal(t, "dept1,dept2", labels["accounts"])
							assert.Equal(t, "DAILY,WEEKDAY,FLEX", labels["flags"])
							foundInfo = true
						}
					}
				}
				assert.True(t, foundInfo, "Should find reservation info metric with all labels")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client and reservations manager
			mockClient := &MockClient{}
			mockReservationsManager := &MockReservationsManager{}

			// Set up expectations
			mockClient.On("Reservations").Return(mockReservationsManager)
			mockReservationsManager.On("List", mock.Anything, mock.Anything).Return(tt.reservationList, tt.reservationErr)

			// Create collector
			logger := logrus.NewEntry(logrus.New())
			collector := NewReservationCollector(mockClient, logger)

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
			mockReservationsManager.AssertExpectations(t)
		})
	}
}

func TestReservationCollector_isReservationActive(t *testing.T) {
	tests := []struct {
		state    string
		expected bool
	}{
		{"ACTIVE", true},
		{"active", true},
		{"RUNNING", true},
		{"running", true},
		{"INACTIVE", false},
		{"inactive", false},
		{"PENDING", false},
		{"pending", false},
		{"EXPIRED", false},
		{"expired", false},
		{"DELETED", false},
		{"deleted", false},
		{"UNKNOWN", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			result := isReservationActive(tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReservationCollector_TimeCalculations(t *testing.T) {
	now := time.Now()
	
	// Create mock client and reservations manager
	mockClient := &MockClient{}
	mockReservationsManager := &MockReservationsManager{}

	reservationList := &slurm.ReservationList{
		Reservations: []slurm.Reservation{
			{
				Name:          "future_res",
				State:         "PENDING",
				PartitionName: "compute",
				NodeCount:     1,
				CoreCount:     32,
				StartTime:     now.Add(1 * time.Hour),
				EndTime:       now.Add(3 * time.Hour),
				Duration:      120, // 2 hours
			},
			{
				Name:          "current_res",
				State:         "ACTIVE",
				PartitionName: "compute",
				NodeCount:     1,
				CoreCount:     32,
				StartTime:     now.Add(-1 * time.Hour),
				EndTime:       now.Add(1 * time.Hour),
				Duration:      120, // 2 hours
			},
			{
				Name:          "past_res",
				State:         "EXPIRED",
				PartitionName: "compute",
				NodeCount:     1,
				CoreCount:     32,
				StartTime:     now.Add(-3 * time.Hour),
				EndTime:       now.Add(-1 * time.Hour),
				Duration:      120, // 2 hours
			},
		},
	}

	mockClient.On("Reservations").Return(mockReservationsManager)
	mockReservationsManager.On("List", mock.Anything, mock.Anything).Return(reservationList, nil)

	// Create collector
	logger := logrus.NewEntry(logrus.New())
	collector := NewReservationCollector(mockClient, logger)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Check remaining time calculations
	remainingTimes := make(map[string]float64)
	durations := make(map[string]float64)
	
	for metric := range ch {
		dto := &prometheus.Metric{}
		metric.Write(dto)

		if dto.Gauge != nil && len(dto.Label) >= 1 {
			resName := *dto.Label[0].Value
			
			if strings.Contains(metric.Desc().String(), "remaining_seconds") {
				remainingTimes[resName] = *dto.Gauge.Value
			}
			if strings.Contains(metric.Desc().String(), "duration_seconds") {
				durations[resName] = *dto.Gauge.Value
			}
		}
	}

	// All reservations should have 2 hours duration (7200 seconds)
	assert.Equal(t, float64(7200), durations["future_res"])
	assert.Equal(t, float64(7200), durations["current_res"])
	assert.Equal(t, float64(7200), durations["past_res"])

	// Past reservation should have 0 remaining time
	assert.Equal(t, float64(0), remainingTimes["past_res"])
	
	// Current reservation should have positive remaining time (less than 1 hour)
	assert.Greater(t, remainingTimes["current_res"], float64(0))
	assert.Less(t, remainingTimes["current_res"], float64(3600))
}