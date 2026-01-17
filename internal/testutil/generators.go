package testutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jontk/slurm-client"
)

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

	states := []string{"RUNNING", "PENDING", "COMPLETED", "FAILED", "CANCELLED"}
	partitions := []string{"gpu", "cpu", "highmem", "debug"}
	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i := 0; i < count; i++ {
		startTime := time.Now().Add(-time.Duration(g.rand.Intn(7*24)) * time.Hour)

		job := slurm.Job{
			ID:         fmt.Sprintf("%d", 10000+i),
			Name:       fmt.Sprintf("job-%d", i),
			State:      states[g.rand.Intn(len(states))],
			UserID:     users[g.rand.Intn(len(users))],
			GroupID:    fmt.Sprintf("group%d", g.rand.Intn(5)+1),
			Partition:  partitions[g.rand.Intn(len(partitions))],
			Priority:   g.rand.Intn(1000),
			SubmitTime: startTime.Add(-time.Duration(g.rand.Intn(60)) * time.Minute),
			StartTime:  &startTime,
			CPUs:       g.rand.Intn(64) + 1,
			Memory:     g.rand.Intn(128*1024) + 1024, // 1GB to 128GB in MB
			TimeLimit:  g.rand.Intn(72*60) + 60,      // 1 to 72 hours in minutes
			WorkingDir: fmt.Sprintf("/home/%s/work", users[g.rand.Intn(len(users))]),
			Command:    fmt.Sprintf("simulation_%d.sh", i),
			Nodes:      g.generateNodeList(g.rand.Intn(4) + 1),
			Metadata: map[string]interface{}{
				"account": fmt.Sprintf("account%d", g.rand.Intn(10)+1),
				"qos":     []string{"normal", "high", "low"}[g.rand.Intn(3)],
			},
		}

		// Set end time for completed jobs
		if job.State == "COMPLETED" || job.State == "FAILED" || job.State == "CANCELLED" {
			endTime := startTime.Add(time.Duration(g.rand.Intn(job.TimeLimit)) * time.Minute)
			job.EndTime = &endTime
		}

		jobs[i] = job
	}

	return jobs
}

// GenerateNodes generates test nodes
func (g *TestDataGenerator) GenerateNodes(count int) []slurm.Node {
	nodes := make([]slurm.Node, count)

	states := []string{"idle", "allocated", "mixed", "down", "drain"}
	partitions := []string{"gpu", "cpu", "highmem", "debug"}

	for i := 0; i < count; i++ {
		node := slurm.Node{
			Name:   fmt.Sprintf("node%03d", i+1),
			State:  states[g.rand.Intn(len(states))],
			CPUs:   []int{16, 32, 64, 128}[g.rand.Intn(4)],
			Memory: []int{64 * 1024, 128 * 1024, 256 * 1024, 512 * 1024}[g.rand.Intn(4)], // GB in MB
			Metadata: map[string]interface{}{
				"arch":      []string{"x86_64", "aarch64"}[g.rand.Intn(2)],
				"features":  g.generateFeatures(),
				"partition": partitions[g.rand.Intn(len(partitions))],
			},
		}

		// Set allocated resources for allocated/mixed nodes
		if node.State == "allocated" || node.State == "mixed" {
			allocatedCPUs := g.rand.Intn(node.CPUs)
			allocatedMemory := g.rand.Intn(node.Memory)

			node.Metadata["allocated_cpus"] = allocatedCPUs
			node.Metadata["allocated_memory"] = allocatedMemory
		}

		nodes[i] = node
	}

	return nodes
}

// GeneratePartitions generates test partitions
func (g *TestDataGenerator) GeneratePartitions(count int) []slurm.Partition {
	partitions := make([]slurm.Partition, count)

	names := []string{"gpu", "cpu", "highmem", "debug", "interactive"}
	states := []string{"UP", "DOWN", "DRAIN"}

	for i := 0; i < count && i < len(names); i++ {
		totalNodes := g.rand.Intn(50) + 10
		availableNodes := g.rand.Intn(totalNodes)
		totalCPUs := totalNodes * (g.rand.Intn(64) + 16)
		idleCPUs := g.rand.Intn(totalCPUs)

		partition := slurm.Partition{
			Name:           names[i],
			State:          states[g.rand.Intn(len(states))],
			TotalNodes:     totalNodes,
			AvailableNodes: availableNodes,
			TotalCPUs:      totalCPUs,
			IdleCPUs:       idleCPUs,
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

// generateFeatures generates node features
func (g *TestDataGenerator) generateFeatures() []string {
	allFeatures := []string{"gpu", "ssd", "infiniband", "large_mem", "nvme", "fpga"}
	count := g.rand.Intn(3) + 1 // 1-3 features

	features := make([]string, 0, count)
	used := make(map[string]bool)

	for len(features) < count {
		feature := allFeatures[g.rand.Intn(len(allFeatures))]
		if !used[feature] {
			features = append(features, feature)
			used[feature] = true
		}
	}

	return features
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
