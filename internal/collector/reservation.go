package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	reservationCollectorSubsystem = "reservation"
)

// ReservationCollector collects reservation-related metrics
type ReservationCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Reservation state metrics
	reservationState *prometheus.Desc

	// Reservation resource metrics
	reservationNodes *prometheus.Desc
	reservationCores *prometheus.Desc

	// Reservation time metrics
	reservationStartTime *prometheus.Desc
	reservationEndTime   *prometheus.Desc
	reservationDuration  *prometheus.Desc
	reservationRemaining *prometheus.Desc

	// Reservation info
	reservationInfo *prometheus.Desc
}

// NewReservationCollector creates a new Reservation collector
func NewReservationCollector(client slurm.SlurmClient, logger *logrus.Entry) *ReservationCollector {
	c := &ReservationCollector{
		client:  client,
		logger:  logger.WithField("collector", "reservation"),
		enabled: true,
	}

	// Initialize metrics
	c.reservationState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "state"),
		"Reservation state (1=active, 0=inactive)",
		[]string{"reservation"},
		nil,
	)

	c.reservationNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "nodes"),
		"Number of nodes in the reservation",
		[]string{"reservation", "partition"},
		nil,
	)

	c.reservationCores = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "cores"),
		"Number of cores in the reservation",
		[]string{"reservation", "partition"},
		nil,
	)

	c.reservationStartTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "start_time_epoch"),
		"Start time of the reservation (Unix timestamp)",
		[]string{"reservation"},
		nil,
	)

	c.reservationEndTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "end_time_epoch"),
		"End time of the reservation (Unix timestamp)",
		[]string{"reservation"},
		nil,
	)

	c.reservationDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "duration_seconds"),
		"Duration of the reservation in seconds",
		[]string{"reservation"},
		nil,
	)

	c.reservationRemaining = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "remaining_seconds"),
		"Remaining time for the reservation in seconds",
		[]string{"reservation"},
		nil,
	)

	c.reservationInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, reservationCollectorSubsystem, "info"),
		"Reservation information with all labels",
		[]string{"reservation", "state", "partition", "users", "accounts", "flags"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *ReservationCollector) Name() string {
	return "reservations"
}

// IsEnabled returns whether this collector is enabled
func (c *ReservationCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *ReservationCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *ReservationCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.reservationState
	ch <- c.reservationNodes
	ch <- c.reservationCores
	ch <- c.reservationStartTime
	ch <- c.reservationEndTime
	ch <- c.reservationDuration
	ch <- c.reservationRemaining
	ch <- c.reservationInfo
}

// Collect implements the Collector interface
func (c *ReservationCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *ReservationCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Reservations manager from client
	reservationsManager := c.client.Reservations()
	if reservationsManager == nil {
		return fmt.Errorf("reservations manager not available")
	}

	// List all reservations
	reservationList, err := reservationsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list reservations")
		return err
	}

	c.logger.WithField("count", len(reservationList.Reservations)).Debug("Collected reservation entries")

	now := time.Now()

	for _, reservation := range reservationList.Reservations {
		// Reservation state
		stateValue := 0.0
		if isReservationActive(reservation.State) {
			stateValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.reservationState,
			prometheus.GaugeValue,
			stateValue,
			reservation.Name,
		)

		// Resource metrics
		ch <- prometheus.MustNewConstMetric(
			c.reservationNodes,
			prometheus.GaugeValue,
			float64(reservation.NodeCount),
			reservation.Name, reservation.PartitionName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.reservationCores,
			prometheus.GaugeValue,
			float64(reservation.CoreCount),
			reservation.Name, reservation.PartitionName,
		)

		// Time metrics
		startTime := reservation.StartTime.Unix()
		endTime := reservation.EndTime.Unix()

		ch <- prometheus.MustNewConstMetric(
			c.reservationStartTime,
			prometheus.GaugeValue,
			float64(startTime),
			reservation.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.reservationEndTime,
			prometheus.GaugeValue,
			float64(endTime),
			reservation.Name,
		)

		// Duration (convert minutes to seconds if needed)
		duration := float64(reservation.Duration * 60) // Convert minutes to seconds
		ch <- prometheus.MustNewConstMetric(
			c.reservationDuration,
			prometheus.GaugeValue,
			duration,
			reservation.Name,
		)

		// Calculate remaining time
		remaining := 0.0
		if reservation.EndTime.After(now) {
			remaining = reservation.EndTime.Sub(now).Seconds()
		}
		ch <- prometheus.MustNewConstMetric(
			c.reservationRemaining,
			prometheus.GaugeValue,
			remaining,
			reservation.Name,
		)

		// Reservation info
		users := strings.Join(reservation.Users, ",")
		accounts := strings.Join(reservation.Accounts, ",")
		flags := strings.Join(reservation.Flags, ",")

		ch <- prometheus.MustNewConstMetric(
			c.reservationInfo,
			prometheus.GaugeValue,
			1,
			reservation.Name,
			reservation.State,
			reservation.PartitionName,
			users,
			accounts,
			flags,
		)
	}

	return nil
}

// isReservationActive returns true if the reservation is in an active state
func isReservationActive(state string) bool {
	state = strings.ToLower(state)
	switch state {
	case "active", "running":
		return true
	case "inactive", "pending", "expired", "deleted":
		return false
	default:
		// Default to considering unknown states as inactive
		return false
	}
}
