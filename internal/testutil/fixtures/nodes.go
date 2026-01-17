package fixtures

import (
	"fmt"

	"github.com/jontk/slurm-client"
)

// GetTestNodes returns test node data
func GetTestNodes() []slurm.Node {
	return []slurm.Node{
		{
			Name:         "node01",
			State:        "idle",
			CPUs:         64,
			Memory:       131072, // 128GB in MB
			Partitions:   []string{"compute", "all"},
			Features:     []string{"haswell", "ib"},
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    0,
				"allocated_memory":  0,
				"tmp_disk":          1000000, // 1TB in MB
				"allocated_tmpDisk": 0,
				"gres":              "",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
		{
			Name:         "node02",
			State:        "allocated",
			CPUs:         64,
			Memory:       131072,
			Partitions:   []string{"compute", "all"},
			Features:     []string{"haswell", "ib"},
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    64,
				"allocated_memory":  65536, // 64GB allocated
				"tmp_disk":          1000000,
				"allocated_tmpDisk": 500000,
				"gres":              "",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
		{
			Name:         "node03",
			State:        "down",
			CPUs:         64,
			Memory:       131072,
			Partitions:   []string{"compute", "all"},
			Features:     []string{"haswell", "ib"},
			Reason:       "Hardware failure",
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    0,
				"allocated_memory":  0,
				"tmp_disk":          1000000,
				"allocated_tmpDisk": 0,
				"gres":              "",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
		{
			Name:         "node04",
			State:        "mixed",
			CPUs:         64,
			Memory:       131072,
			Partitions:   []string{"compute", "all"},
			Features:     []string{"haswell", "ib"},
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    32,
				"allocated_memory":  32768, // 32GB allocated
				"tmp_disk":          1000000,
				"allocated_tmpDisk": 250000,
				"gres":              "",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
		{
			Name:         "gpu-node01",
			State:        "idle",
			CPUs:         32,
			Memory:       262144, // 256GB
			Partitions:   []string{"gpu", "all"},
			Features:     []string{"skylake", "ib", "gpu"},
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    0,
				"allocated_memory":  0,
				"tmp_disk":          2000000, // 2TB
				"allocated_tmpDisk": 0,
				"gres":              "gpu:4",
				"gres_used":         "gpu:0",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
		{
			Name:         "gpu-node02",
			State:        "drain",
			CPUs:         32,
			Memory:       262144,
			Partitions:   []string{"gpu", "all"},
			Features:     []string{"skylake", "ib", "gpu"},
			Reason:       "Maintenance",
			Architecture: "x86_64",
			Metadata: map[string]interface{}{
				"allocated_cpus":    0,
				"allocated_memory":  0,
				"tmp_disk":          2000000,
				"allocated_tmpDisk": 0,
				"gres":              "gpu:4",
				"gres_used":         "gpu:0",
				"os":                "Linux",
				"kernel_version":    "5.15.0",
			},
		},
	}
}

// GetTestNodeList returns a complete test node list
func GetTestNodeList() *slurm.NodeList {
	return &slurm.NodeList{
		Nodes: GetTestNodes(),
	}
}

// GetIdleNodeList returns only idle nodes
func GetIdleNodeList() *slurm.NodeList {
	nodes := GetTestNodes()
	idleNodes := []slurm.Node{}
	for _, node := range nodes {
		if node.State == "idle" {
			idleNodes = append(idleNodes, node)
		}
	}
	return &slurm.NodeList{
		Nodes: idleNodes,
	}
}

// GetEmptyNodeList returns an empty node list
func GetEmptyNodeList() *slurm.NodeList {
	return &slurm.NodeList{
		Nodes: []slurm.Node{},
	}
}

// GenerateLargeNodeList generates a large number of nodes for performance testing
func GenerateLargeNodeList(count int) *slurm.NodeList {
	nodes := make([]slurm.Node, count)

	states := []string{"idle", "allocated", "mixed", "down", "drain", "draining"}
	partitionSets := [][]string{
		{"compute", "all"},
		{"gpu", "all"},
		{"bigmem", "all"},
		{"debug", "all"},
		{"interactive"},
	}

	for i := 0; i < count; i++ {
		nodeName := fmt.Sprintf("node%04d", i+1)
		state := states[i%len(states)]
		partitions := partitionSets[i%len(partitionSets)]

		// Create realistic node with varying specifications
		cpus := 16 + (i%9)*8              // 16, 24, 32, 40, 48, 56, 64, 72, 80 CPUs
		memory := 32768 + (i%16)*8192     // 32GB-128GB memory in MB
		tmpDisk := 500000 + (i%20)*100000 // 500GB-2.5TB temp disk

		// Calculate allocation based on state
		var allocatedCPUs int
		var allocatedMemory int
		var allocatedTmpDisk int

		switch state {
		case "allocated":
			allocatedCPUs = cpus
			allocatedMemory = memory
			allocatedTmpDisk = tmpDisk
		case "mixed":
			allocatedCPUs = cpus / 2
			allocatedMemory = memory / 2
			allocatedTmpDisk = tmpDisk / 2
		default:
			allocatedCPUs = 0
			allocatedMemory = 0
			allocatedTmpDisk = 0
		}

		metadata := map[string]interface{}{
			"allocated_cpus":    allocatedCPUs,
			"allocated_memory":  allocatedMemory,
			"tmp_disk":          tmpDisk,
			"allocated_tmpDisk": allocatedTmpDisk,
			"gres":              "",
			"os":                "Linux",
			"kernel_version":    "5.15.0",
		}

		node := slurm.Node{
			Name:         nodeName,
			State:        state,
			CPUs:         cpus,
			Memory:       memory,
			Partitions:   partitions,
			Features:     []string{"haswell", "ib"},
			Architecture: "x86_64",
			Metadata:     metadata,
		}

		// Add GPU nodes for some nodes
		if i%10 == 0 {
			node.Metadata["gres"] = "gpu:4"
			node.Metadata["gres_used"] = fmt.Sprintf("gpu:%d", allocatedCPUs/8) // Assume 8 CPUs per GPU
			node.Features = append(node.Features, "gpu")
		}

		// Add reasons for down/drain nodes
		if state == "down" || state == "drain" || state == "draining" {
			reasons := []string{"Hardware failure", "Maintenance", "Power issue", "Network problem"}
			node.Reason = reasons[i%len(reasons)]
		}

		nodes[i] = node
	}

	return &slurm.NodeList{
		Nodes: nodes,
	}
}
