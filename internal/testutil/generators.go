// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package testutil

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	slurm "github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/api"
)

// Helper functions for pointer types
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func uint32Ptr(i uint32) *uint32 {
	return &i
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}

// TestDataGenerator provides functions to generate test data
type TestDataGenerator struct {
	rand *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateJobs generates test jobs
func (g *TestDataGenerator) GenerateJobs(count int) []slurm.Job {
	jobs := make([]slurm.Job, count)

	states := []api.JobState{api.JobStateRunning, api.JobStatePending, api.JobStateCompleted, api.JobStateFailed, api.JobStateCancelled}
	partitions := []string{"gpu", "cpu", "highmem", "debug"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i := 0; i < count; i++ {
		startTime := time.Now().Add(-time.Duration(g.rand.Intn(7*24)) * time.Hour)
		jobID := int32(10000 + i)
		partition := partitions[g.rand.Intn(len(partitions))]
		user := users[g.rand.Intn(len(users))]
		state := states[g.rand.Intn(len(states))]

		job := slurm.Job{
			JobID:                   &jobID,
			Name:                    strPtr(fmt.Sprintf("job-%d", i)),
			JobState:                []api.JobState{state},
			UserID:                  int32Ptr(int32(1000 + g.rand.Intn(100))),
			UserName:                &user,
			GroupID:                 int32Ptr(int32(1000 + g.rand.Intn(5))),
			GroupName:               strPtr(fmt.Sprintf("group%d", g.rand.Intn(5)+1)),
			Partition:               &partition,
			Priority:                uint32Ptr(uint32(g.rand.Intn(1000))),
			SubmitTime:              startTime.Add(-time.Duration(g.rand.Intn(60)) * time.Minute),
			StartTime:               startTime,
			CPUs:                    uint32Ptr(uint32(g.rand.Intn(64) + 1)),
			MemoryPerNode:           uint64Ptr(uint64(g.rand.Intn(128*1024) + 1024)), // 1GB to 128GB in MB
			TimeLimit:               uint32Ptr(uint32(g.rand.Intn(72*60) + 60)),      // 1 to 72 hours in minutes
			CurrentWorkingDirectory: strPtr(fmt.Sprintf("/home/%s/work", user)),
			Command:                 strPtr(fmt.Sprintf("simulation_%d.sh", i)),
			Nodes:                   strPtr(strings.Join(g.generateNodeList(g.rand.Intn(4)+1), ",")),
		}

		// Set end time for completed jobs
		if state == api.JobStateCompleted || state == api.JobStateFailed || state == api.JobStateCancelled {
			var timeLimit int
			if job.TimeLimit != nil && *job.TimeLimit > 0 {
				timeLimit = int(*job.TimeLimit)
			}
			if timeLimit == 0 {
				timeLimit = 60
			}
			endTime := startTime.Add(time.Duration(g.rand.Intn(timeLimit)) * time.Minute)
			job.EndTime = endTime
		}

		jobs[i] = job
	}

	return jobs
}

// GenerateNodes generates test nodes  
func (g *TestDataGenerator) GenerateNodes(count int) []slurm.Node {
	nodes := make([]slurm.Node, count)

	states := []api.NodeState{api.NodeStateIdle, api.NodeStateAllocated, api.NodeStateMixed, api.NodeStateDown, api.NodeStateDrain}
	
	for i := 0; i < count; i++ {
		cpus := int32([]int{16, 32, 64, 128}[g.rand.Intn(4)])
		memory := int64([]int{64 * 1024, 128 * 1024, 256 * 1024, 512 * 1024}[g.rand.Intn(4)]) // GB in MB

		node := slurm.Node{
			Name:       strPtr(fmt.Sprintf("node%03d", i+1)),
			State:      []api.NodeState{states[g.rand.Intn(len(states))]},
			CPUs:       &cpus,
			RealMemory: &memory,
		}

		nodes[i] = node
	}

	return nodes
}

// GeneratePartitions generates test partitions
func (g *TestDataGenerator) GeneratePartitions(count int) []slurm.Partition {
	partitions := make([]slurm.Partition, count)

	names := []string{"gpu", "cpu", "highmem", "debug", "interactive"}

	for i := 0; i < count && i < len(names); i++ {
		partition := slurm.Partition{
			Name: strPtr(names[i]),
		}

		partitions[i] = partition
	}

	return partitions[:count]
}

// generateNodeList generates a list of node names
func (g *TestDataGenerator) generateNodeList(count int) []string {
	nodes := make([]string, count)
	for i := 0; i < count; i++ {
		nodes[i] = fmt.Sprintf("node%03d", g.rand.Intn(1000)+1)
	}
	return nodes
}


// GenerateJobStates generates realistic job state distributions
func (g *TestDataGenerator) GenerateJobStates() map[string]int {
	total := g.rand.Intn(1000) + 100

	// Realistic distribution
	running := int(float64(total) * 0.4)                        // 40% running
	pending := int(float64(total) * 0.3)                        // 30% pending
	completed := int(float64(total) * 0.25)                     // 25% completed
	failed := int(float64(total) * 0.04)                        // 4% failed
	cancelled := total - running - pending - completed - failed // remainder

	return map[string]int{
		"RUNNING":   running,
		"PENDING":   pending,
		"COMPLETED": completed,
		"FAILED":    failed,
		"CANCELLED": cancelled,
	}
}

// GenerateClusterLoad generates realistic cluster load data
func (g *TestDataGenerator) GenerateClusterLoad() map[string]interface{} {
	totalNodes := g.rand.Intn(500) + 100
	allocatedNodes := int(float64(totalNodes) * (0.6 + g.rand.Float64()*0.3)) // 60-90% allocated

	totalCPUs := totalNodes * 32                                            // Average 32 CPUs per node
	allocatedCPUs := int(float64(totalCPUs) * (0.7 + g.rand.Float64()*0.2)) // 70-90% allocated

	return map[string]interface{}{
		"total_nodes":     totalNodes,
		"allocated_nodes": allocatedNodes,
		"idle_nodes":      totalNodes - allocatedNodes,
		"total_cpus":      totalCPUs,
		"allocated_cpus":  allocatedCPUs,
		"idle_cpus":       totalCPUs - allocatedCPUs,
		"load_factor":     float64(allocatedCPUs) / float64(totalCPUs),
	}
}

// GenerateTimeSeriesData generates time series data for testing
func (g *TestDataGenerator) GenerateTimeSeriesData(points int, interval time.Duration) []TimePoint {
	data := make([]TimePoint, points)
	baseTime := time.Now().Add(-time.Duration(points) * interval)

	for i := 0; i < points; i++ {
		data[i] = TimePoint{
			Time:  baseTime.Add(time.Duration(i) * interval),
			Value: g.rand.Float64() * 100,
		}
	}

	return data
}

// TimePoint represents a point in time series data
type TimePoint struct {
	Time  time.Time
	Value float64
}

// Global test data generator
var Generator = NewTestDataGenerator()
