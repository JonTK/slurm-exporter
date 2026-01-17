package fixtures

import (
	"fmt"
	"time"

	"github.com/jontk/slurm-client"
)

// GetTestJobs returns test job data
func GetTestJobs() []slurm.Job {
	now := time.Now()
	startTime1 := now.Add(-1 * time.Hour)
	startTime2 := now.Add(-3 * time.Hour)
	endTime1 := now.Add(-1 * time.Hour)
	startTime3 := now.Add(-90 * time.Minute)
	endTime2 := now.Add(-30 * time.Minute)
	startTime4 := now.Add(-45 * time.Minute)

	return []slurm.Job{
		{
			ID:         "12345",
			Name:       "test-job-1",
			State:      "RUNNING",
			UserID:     "user1",
			GroupID:    "group1",
			Partition:  "compute",
			Priority:   1000,
			SubmitTime: now.Add(-2 * time.Hour),
			StartTime:  &startTime1,
			CPUs:       16,
			Memory:     32768,   // 32GB in MB
			TimeLimit:  24 * 60, // 24 hours in minutes
			WorkingDir: "/home/user1",
			Command:    "test-job.sh",
			Nodes:      []string{"node01", "node02"},
			Metadata: map[string]interface{}{
				"account": "research",
				"qos":     "normal",
			},
		},
		{
			ID:         "12346",
			Name:       "test-job-2",
			State:      "PENDING",
			UserID:     "user2",
			GroupID:    "group2",
			Partition:  "compute",
			Priority:   500,
			SubmitTime: now.Add(-30 * time.Minute),
			CPUs:       32,
			Memory:     65536,   // 64GB in MB
			TimeLimit:  48 * 60, // 48 hours in minutes
			WorkingDir: "/home/user2",
			Command:    "test-job.sh",
			Metadata: map[string]interface{}{
				"account": "engineering",
				"qos":     "high",
				"reason":  "Resources",
			},
		},
		{
			ID:         "12347",
			Name:       "test-job-3",
			State:      "COMPLETED",
			UserID:     "user1",
			GroupID:    "group1",
			Partition:  "gpu",
			Priority:   750,
			SubmitTime: now.Add(-4 * time.Hour),
			StartTime:  &startTime2,
			EndTime:    &endTime1,
			CPUs:       8,
			Memory:     16384,  // 16GB in MB
			TimeLimit:  4 * 60, // 4 hours in minutes
			WorkingDir: "/home/user1",
			Command:    "gpu-job.sh",
			Nodes:      []string{"gpu-node01"},
			ExitCode:   0,
			Metadata: map[string]interface{}{
				"account": "research",
				"qos":     "normal",
				"gpus":    2,
			},
		},
		{
			ID:         "12348",
			Name:       "test-job-4",
			State:      "FAILED",
			UserID:     "user3",
			GroupID:    "group3",
			Partition:  "compute",
			Priority:   250,
			SubmitTime: now.Add(-2 * time.Hour),
			StartTime:  &startTime3,
			EndTime:    &endTime2,
			CPUs:       4,
			Memory:     8192,   // 8GB in MB
			TimeLimit:  2 * 60, // 2 hours in minutes
			WorkingDir: "/home/user3",
			Command:    "test-job.sh",
			Nodes:      []string{"node03"},
			ExitCode:   1,
			Metadata: map[string]interface{}{
				"account": "finance",
				"qos":     "low",
				"reason":  "NonZeroExitCode",
			},
		},
		{
			ID:         "12349",
			Name:       "array-job",
			State:      "RUNNING",
			UserID:     "user2",
			GroupID:    "group2",
			Partition:  "compute",
			Priority:   600,
			SubmitTime: now.Add(-1 * time.Hour),
			StartTime:  &startTime4,
			CPUs:       2,
			Memory:     4096,   // 4GB in MB
			TimeLimit:  1 * 60, // 1 hour in minutes
			WorkingDir: "/home/user2",
			Command:    "array-job.sh",
			Nodes:      []string{"node04"},
			Metadata: map[string]interface{}{
				"account":        "engineering",
				"qos":            "normal",
				"array_job_id":   12349,
				"array_task_id":  1,
				"array_task_str": "1-10",
			},
		},
	}
}

// GetEmptyJobList returns an empty job list
func GetEmptyJobList() *slurm.JobList {
	return &slurm.JobList{
		Jobs: []slurm.Job{},
	}
}

// GetRunningJobList returns only running jobs
func GetRunningJobList() *slurm.JobList {
	jobs := GetTestJobs()
	runningJobs := []slurm.Job{}
	for _, job := range jobs {
		if job.State == "RUNNING" {
			runningJobs = append(runningJobs, job)
		}
	}
	return &slurm.JobList{
		Jobs: runningJobs,
	}
}

// GetTestJobList returns a complete test job list
func GetTestJobList() *slurm.JobList {
	return &slurm.JobList{
		Jobs: GetTestJobs(),
	}
}

// GenerateLargeJobList generates a large number of jobs for performance testing
func GenerateLargeJobList(count int) *slurm.JobList {
	jobs := make([]slurm.Job, count)

	states := []string{"PENDING", "RUNNING", "COMPLETED", "FAILED", "CANCELLED"}
	partitions := []string{"compute", "gpu", "bigmem", "debug", "interactive"}
	users := []string{"user1", "user2", "user3", "user4", "user5", "admin", "researcher"}

	for i := 0; i < count; i++ {
		jobID := fmt.Sprintf("%d", 100000+i)
		state := states[i%len(states)]
		partition := partitions[i%len(partitions)]
		user := users[i%len(users)]

		// Create realistic job with varying resource requirements
		cpus := 1 + (i % 64)                           // 1-64 CPUs
		memory := 1024 + (i%32)*1024                   // 1GB-32GB memory
		runtime := time.Duration(i%3600) * time.Second // 0-1 hour runtime

		startTime := time.Now().Add(-runtime)
		endTime := time.Now()

		job := slurm.Job{
			ID:         jobID,
			Name:       fmt.Sprintf("job_%d", i),
			State:      state,
			Partition:  partition,
			UserID:     fmt.Sprintf("%d", 1000+i%100),
			GroupID:    fmt.Sprintf("group%d", i%10),
			Priority:   1000 + (i % 1000),
			CPUs:       cpus,
			Memory:     memory,
			TimeLimit:  24 * 60, // 24 hours in minutes
			SubmitTime: time.Now().Add(-time.Duration(i) * time.Minute),
			WorkingDir: fmt.Sprintf("/home/%s", user),
			Command:    fmt.Sprintf("job_%d.sh", i),
			Nodes:      []string{fmt.Sprintf("node%02d", (i%10)+1)},
			ExitCode:   0,
			Metadata: map[string]interface{}{
				"account": "default",
				"qos":     "normal",
			},
		}

		// Add start/end times for non-pending jobs
		if state != "PENDING" {
			job.StartTime = &startTime
		}
		if state == "COMPLETED" || state == "FAILED" || state == "CANCELLED" {
			job.EndTime = &endTime
		}

		// Add some variety to failed jobs
		if state == "FAILED" {
			job.ExitCode = 1
		}

		// Add GPU requirements for GPU partition
		if partition == "gpu" {
			job.Metadata["gpus"] = 1 + (i % 4)
		}

		// Add array job information for some jobs
		if i%10 == 0 {
			job.Metadata["array_job_id"] = 100000 + i
			job.Metadata["array_task_id"] = i % 5
		}

		jobs[i] = job
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}

// GenerateJobBatch generates a batch of jobs with specific characteristics
func GenerateJobBatch(count int, stateFilter string, partitionFilter string) *slurm.JobList {
	jobs := make([]slurm.Job, count)

	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i := 0; i < count; i++ {
		jobID := fmt.Sprintf("%d", 200000+i)
		user := users[i%len(users)]

		startTime := time.Now().Add(-time.Duration(i) * time.Second)
		endTime := time.Now()

		job := slurm.Job{
			ID:         jobID,
			Name:       fmt.Sprintf("batch_job_%d", i),
			State:      stateFilter,
			Partition:  partitionFilter,
			UserID:     fmt.Sprintf("%d", 2000+i%50),
			GroupID:    fmt.Sprintf("group%d", i%5),
			Priority:   500 + i,
			CPUs:       1 + (i % 32),
			Memory:     2048 + (i%16)*1024,
			TimeLimit:  12 * 60, // 12 hours in minutes
			SubmitTime: time.Now().Add(-time.Duration(i) * time.Second),
			WorkingDir: fmt.Sprintf("/home/%s", user),
			Command:    fmt.Sprintf("batch_job_%d.sh", i),
			Nodes:      []string{fmt.Sprintf("node%02d", (i%20)+1)},
			ExitCode:   0,
			Metadata: map[string]interface{}{
				"account": "batch",
				"qos":     "normal",
			},
		}

		// Add start/end times based on state
		if stateFilter != "PENDING" {
			job.StartTime = &startTime
		}
		if stateFilter == "COMPLETED" || stateFilter == "FAILED" || stateFilter == "CANCELLED" {
			job.EndTime = &endTime
		}

		jobs[i] = job
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}
