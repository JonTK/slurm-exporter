package fixtures

import (
	"github.com/jontk/slurm-client"
)

// GetTestPartitions returns test partition data
func GetTestPartitions() []slurm.Partition {
	return []slurm.Partition{
		{
			Name:           "compute",
			State:          "up",
			TotalNodes:     4,
			AvailableNodes: 2,
			TotalCPUs:      256,
			IdleCPUs:       64,
			DefaultTime:    24 * 60,     // 24 hours in minutes
			MaxTime:        7 * 24 * 60, // 7 days in minutes
			DefaultMemory:  8192,        // 8GB default memory
			MaxMemory:      32768,       // 32GB max memory
			Nodes:          []string{"node01", "node02", "node03", "node04"},
			AllowedGroups:  []string{"all"},
			DeniedGroups:   []string{},
			Priority:       100,
		},
		{
			Name:           "gpu",
			State:          "up",
			TotalNodes:     2,
			AvailableNodes: 1,
			TotalCPUs:      64,
			IdleCPUs:       32,
			DefaultTime:    12 * 60,     // 12 hours in minutes
			MaxTime:        3 * 24 * 60, // 3 days in minutes
			DefaultMemory:  16384,       // 16GB default memory
			MaxMemory:      65536,       // 64GB max memory
			Nodes:          []string{"gpu-node01", "gpu-node02"},
			AllowedGroups:  []string{"gpu-users"},
			DeniedGroups:   []string{},
			Priority:       200,
		},
		{
			Name:           "all",
			State:          "up",
			TotalNodes:     6,
			AvailableNodes: 3,
			TotalCPUs:      320,
			IdleCPUs:       96,
			DefaultTime:    1 * 60,  // 1 hour in minutes
			MaxTime:        24 * 60, // 24 hours in minutes
			DefaultMemory:  4096,    // 4GB default memory
			MaxMemory:      16384,   // 16GB max memory
			Nodes:          []string{"node01", "node02", "node03", "node04", "gpu-node01", "gpu-node02"},
			AllowedGroups:  []string{"all"},
			DeniedGroups:   []string{},
			Priority:       50,
		},
		{
			Name:           "maintenance",
			State:          "down",
			TotalNodes:     0,
			AvailableNodes: 0,
			TotalCPUs:      0,
			IdleCPUs:       0,
			DefaultTime:    1 * 60, // 1 hour in minutes
			MaxTime:        4 * 60, // 4 hours in minutes
			DefaultMemory:  4096,
			MaxMemory:      8192,
			Nodes:          []string{},
			AllowedGroups:  []string{"admin"},
			DeniedGroups:   []string{},
			Priority:       1000,
		},
		{
			Name:           "bigmem",
			State:          "up",
			TotalNodes:     1,
			AvailableNodes: 1,
			TotalCPUs:      128,
			IdleCPUs:       128,
			DefaultTime:    48 * 60,      // 48 hours in minutes
			MaxTime:        14 * 24 * 60, // 14 days in minutes
			DefaultMemory:  65536,        // 64GB default memory
			MaxMemory:      131072,       // 128GB max memory
			Nodes:          []string{"bigmem01"},
			AllowedGroups:  []string{"bigmem-users"},
			DeniedGroups:   []string{},
			Priority:       150,
		},
	}
}

// GetTestPartitionList returns a complete test partition list
func GetTestPartitionList() *slurm.PartitionList {
	return &slurm.PartitionList{
		Partitions: GetTestPartitions(),
	}
}

// GetActivePartitionList returns only active (up) partitions
func GetActivePartitionList() *slurm.PartitionList {
	partitions := GetTestPartitions()
	activePartitions := []slurm.Partition{}
	for _, partition := range partitions {
		if partition.State == "up" {
			activePartitions = append(activePartitions, partition)
		}
	}
	return &slurm.PartitionList{
		Partitions: activePartitions,
	}
}

// GetEmptyPartitionList returns an empty partition list
func GetEmptyPartitionList() *slurm.PartitionList {
	return &slurm.PartitionList{
		Partitions: []slurm.Partition{},
	}
}
