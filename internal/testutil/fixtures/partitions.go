package fixtures

import (
	slurm "github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/api"
)

// GetTestPartitions returns a list of test partitions for testing
func GetTestPartitions() []slurm.Partition {
	return []slurm.Partition{
		{
			Name: strPtr("compute"),
		},
		{
			Name: strPtr("gpu"),
		},
		{
			Name: strPtr("highmem"),
		},
	}
}

// GetTestPartitionList returns a PartitionList for testing
func GetTestPartitionList() *slurm.PartitionList {
	int32Ptr := func(i int32) *int32 { return &i }

	return &slurm.PartitionList{
		Partitions: []slurm.Partition{
			{
				Name: strPtr("compute"),
				Partition: &slurm.PartitionPartition{
					State: []slurm.StateValue{api.StateUp},
				},
				Nodes: &slurm.PartitionNodes{
					Total: int32Ptr(10),
				},
				CPUs: &slurm.PartitionCPUs{
					Total: int32Ptr(80),
				},
			},
			{
				Name: strPtr("gpu"),
				Partition: &slurm.PartitionPartition{
					State: []slurm.StateValue{api.StateUp},
				},
				Nodes: &slurm.PartitionNodes{
					Total: int32Ptr(4),
				},
				CPUs: &slurm.PartitionCPUs{
					Total: int32Ptr(32),
				},
			},
		},
	}
}

// GetEmptyPartitionList returns an empty PartitionList for testing
func GetEmptyPartitionList() *slurm.PartitionList {
	return &slurm.PartitionList{
		Partitions: []slurm.Partition{},
	}
}

// GetActivePartitionList returns a PartitionList with active partitions for testing
func GetActivePartitionList() *slurm.PartitionList {
	int32Ptr := func(i int32) *int32 { return &i }

	return &slurm.PartitionList{
		Partitions: []slurm.Partition{
			{
				Name: strPtr("compute"),
				Partition: &slurm.PartitionPartition{
					State: []slurm.StateValue{api.StateUp},
				},
				Nodes: &slurm.PartitionNodes{
					Total: int32Ptr(20),
				},
				CPUs: &slurm.PartitionCPUs{
					Total: int32Ptr(160),
				},
			},
		},
	}
}
