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
			Name:       strPtr("node01"),
			State:      []api.NodeState{api.NodeStateIdle},
			CPUs:       int32Ptr(32),
			RealMemory: int64Ptr(64 * 1024), // 64GB in MB
			Architecture: strPtr("x86_64"),
		},
		{
			Name:       strPtr("node02"),
			State:      []api.NodeState{api.NodeStateAllocated},
			CPUs:       int32Ptr(64),
			RealMemory: int64Ptr(128 * 1024), // 128GB in MB
		},
		{
			Name:       strPtr("node03"),
			State:      []api.NodeState{api.NodeStateMixed},
			CPUs:       int32Ptr(48),
			RealMemory: int64Ptr(96 * 1024), // 96GB in MB
		},
	}
}
