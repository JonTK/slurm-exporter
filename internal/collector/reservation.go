// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	slurm "github.com/jontk/slurm-client"
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
	return c.collect(ctx, ch)
}

// collect gathers metrics from SLURM
func (c *ReservationCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		// Extract name from pointer
		name := ""
		if reservation.Name != nil {
			name = *reservation.Name
		}

		// Calculate state from times (no State field in new API)
		isActive := !reservation.StartTime.IsZero() && !reservation.EndTime.IsZero() &&
			now.After(reservation.StartTime) && now.Before(reservation.EndTime)

		// Reservation state
		stateValue := 0.0
		if isActive {
			stateValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.reservationState,
			prometheus.GaugeValue,
			stateValue,
			name,
		)

		// Extract partition from pointer (field renamed from PartitionName to Partition)
		partition := ""
		if reservation.Partition != nil {
			partition = *reservation.Partition
		}

		// Extract node count from pointer
		nodeCount := int32(0)
		if reservation.NodeCount != nil {
			nodeCount = *reservation.NodeCount
		}

		// Extract core count from pointer
		coreCount := int32(0)
		if reservation.CoreCount != nil {
			coreCount = *reservation.CoreCount
		}

		// Resource metrics
		ch <- prometheus.MustNewConstMetric(
			c.reservationNodes,
			prometheus.GaugeValue,
			float64(nodeCount),
			name, partition,
		)

		ch <- prometheus.MustNewConstMetric(
			c.reservationCores,
			prometheus.GaugeValue,
			float64(coreCount),
			name, partition,
		)

		// Time metrics
		startTime := reservation.StartTime.Unix()
		endTime := reservation.EndTime.Unix()

		ch <- prometheus.MustNewConstMetric(
			c.reservationStartTime,
			prometheus.GaugeValue,
			float64(startTime),
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.reservationEndTime,
			prometheus.GaugeValue,
			float64(endTime),
			name,
		)

		// Duration (calculate from end - start, no Duration field in new API)
		duration := 0.0
		if !reservation.StartTime.IsZero() && !reservation.EndTime.IsZero() {
			duration = reservation.EndTime.Sub(reservation.StartTime).Seconds()
		}
		ch <- prometheus.MustNewConstMetric(
			c.reservationDuration,
			prometheus.GaugeValue,
			duration,
			name,
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
			name,
		)

		// Reservation info
		// Users and Accounts are now *string (comma-separated), not []string
		users := ""
		if reservation.Users != nil {
			users = *reservation.Users
		}

		accounts := ""
		if reservation.Accounts != nil {
			accounts = *reservation.Accounts
		}

		// Flags are now []ReservationFlagsValue enum
		flagStrs := make([]string, 0, len(reservation.Flags))
		for _, f := range reservation.Flags {
			flagStrs = append(flagStrs, string(f))
		}
		flags := strings.Join(flagStrs, ",")

		// Calculate state string for info metric
		stateStr := "inactive"
		if isActive {
			stateStr = "active"
		}

		ch <- prometheus.MustNewConstMetric(
			c.reservationInfo,
			prometheus.GaugeValue,
			1,
			name,
			stateStr,
			partition,
			users,
			accounts,
			flags,
		)
	}

	return nil
}
