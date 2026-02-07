package fixtures

import (
	slurm "github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/api"
)

// Helper functions defined in jobs.go
// strPtr, int32Ptr, uint32Ptr, uint64Ptr

func int64Ptr(i int64) *int64 {
	return &i
}

// GetTestNodes returns a list of test nodes for testing
func GetTestNodes() []slurm.Node {
	return []slurm.Node{
		{
			Name:            strPtr("node01"),
			State:           []api.NodeState{api.NodeStateIdle},
			CPUs:            int32Ptr(32),
			AllocCPUs:       int32Ptr(0),
			RealMemory:      int64Ptr(64 * 1024), // 64GB in MB
			AllocMemory:     int64Ptr(0),         // 0MB allocated
			Architecture:    strPtr("x86_64"),
			OperatingSystem: strPtr("Linux"),
			Partitions:      []string{"general", "debug"},
		},
		{
			Name:            strPtr("node02"),
			State:           []api.NodeState{api.NodeStateAllocated},
			CPUs:            int32Ptr(64),
			AllocCPUs:       int32Ptr(64),
			RealMemory:      int64Ptr(128 * 1024), // 128GB in MB
			AllocMemory:     int64Ptr(128 * 1024), // 128GB in MB
			Architecture:    strPtr("x86_64"),
			OperatingSystem: strPtr("Linux"),
			Partitions:      []string{"general"},
		},
		{
			Name:            strPtr("node03"),
			State:           []api.NodeState{api.NodeStateMixed},
			CPUs:            int32Ptr(48),
			AllocCPUs:       int32Ptr(24),
			RealMemory:      int64Ptr(96 * 1024), // 96GB in MB
			AllocMemory:     int64Ptr(48 * 1024), // 48GB in MB (50% allocated)
			Architecture:    strPtr("aarch64"),
			OperatingSystem: strPtr("Linux 5.15"),
			Partitions:      []string{"general", "gpu"},
		},
		{
			Name:            strPtr("node04"),
			State:           []api.NodeState{api.NodeStateDown},
			CPUs:            int32Ptr(32),
			AllocCPUs:       int32Ptr(0),
			RealMemory:      int64Ptr(64 * 1024),
			AllocMemory:     int64Ptr(0),
			Architecture:    strPtr("x86_64"),
			OperatingSystem: strPtr("Linux"),
			Partitions:      []string{"general"},
		},
	}
}

// GetTestNodeList returns a NodeList for testing
func GetTestNodeList() *slurm.NodeList {
	return &slurm.NodeList{
		Nodes: GetTestNodes(),
	}
}
