package fixtures

import (
	"fmt"
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

// createRunningTestJob creates a running test job
func createRunningTestJob(now time.Time, startTime time.Time) slurm.Job {
	return slurm.Job{
		JobID:                   int32Ptr(12345),
		Name:                    strPtr("test-job-1"),
		JobState:                []api.JobState{api.JobStateRunning},
		UserID:                  int32Ptr(1001),
		UserName:                strPtr("user1"),
		GroupID:                 int32Ptr(1001),
		GroupName:               strPtr("group1"),
		Partition:               strPtr("compute"),
		Priority:                uint32Ptr(1000),
		SubmitTime:              now.Add(-2 * time.Hour),
		StartTime:               startTime,
		CPUs:                    uint32Ptr(16),
		MemoryPerNode:           uint64Ptr(32768), // 32GB in MB
		TimeLimit:               uint32Ptr(24 * 60), // 24 hours in minutes
		CurrentWorkingDirectory: strPtr("/home/user1"),
		Command:                 strPtr("test-job.sh"),
		Nodes:                   strPtr("node01,node02"),
	}
}

// createPendingTestJob creates a pending test job
func createPendingTestJob(now time.Time) slurm.Job {
	return slurm.Job{
		JobID:                   int32Ptr(12346),
		Name:                    strPtr("test-job-2"),
		JobState:                []api.JobState{api.JobStatePending},
		UserID:                  int32Ptr(1002),
		UserName:                strPtr("user2"),
		GroupID:                 int32Ptr(1002),
		GroupName:               strPtr("group2"),
		Partition:               strPtr("compute"),
		Priority:                uint32Ptr(500),
		SubmitTime:              now.Add(-30 * time.Minute),
		CPUs:                    uint32Ptr(32),
		MemoryPerNode:           uint64Ptr(65536), // 64GB in MB
		TimeLimit:               uint32Ptr(48 * 60), // 48 hours in minutes
		CurrentWorkingDirectory: strPtr("/home/user2"),
		Command:                 strPtr("test-job.sh"),
	}
}

// createCompletedTestJob creates a completed test job
func createCompletedTestJob(now time.Time, startTime, endTime time.Time) slurm.Job {
	exitCode := api.ExitCode{
		ReturnCode: uint32Ptr(0),
	}
	return slurm.Job{
		JobID:                   int32Ptr(12347),
		Name:                    strPtr("test-job-3"),
		JobState:                []api.JobState{api.JobStateCompleted},
		UserID:                  int32Ptr(1001),
		UserName:                strPtr("user1"),
		GroupID:                 int32Ptr(1001),
		GroupName:               strPtr("group1"),
		Partition:               strPtr("gpu"),
		Priority:                uint32Ptr(750),
		SubmitTime:              now.Add(-4 * time.Hour),
		StartTime:               startTime,
		EndTime:                 endTime,
		CPUs:                    uint32Ptr(8),
		MemoryPerNode:           uint64Ptr(16384), // 16GB in MB
		TimeLimit:               uint32Ptr(4 * 60), // 4 hours in minutes
		CurrentWorkingDirectory: strPtr("/home/user1"),
		Command:                 strPtr("gpu-job.sh"),
		Nodes:                   strPtr("gpu-node01"),
		ExitCode:                &exitCode,
	}
}

// createFailedTestJob creates a failed test job
func createFailedTestJob(now time.Time, startTime, endTime time.Time) slurm.Job {
	exitCode := api.ExitCode{
		ReturnCode: uint32Ptr(1),
	}
	return slurm.Job{
		JobID:                   int32Ptr(12348),
		Name:                    strPtr("test-job-4"),
		JobState:                []api.JobState{api.JobStateFailed},
		UserID:                  int32Ptr(1003),
		UserName:                strPtr("user3"),
		GroupID:                 int32Ptr(1003),
		GroupName:               strPtr("group3"),
		Partition:               strPtr("compute"),
		Priority:                uint32Ptr(250),
		SubmitTime:              now.Add(-2 * time.Hour),
		StartTime:               startTime,
		EndTime:                 endTime,
		CPUs:                    uint32Ptr(4),
		MemoryPerNode:           uint64Ptr(8192), // 8GB in MB
		TimeLimit:               uint32Ptr(2 * 60), // 2 hours in minutes
		CurrentWorkingDirectory: strPtr("/home/user3"),
		Command:                 strPtr("test-job.sh"),
		Nodes:                   strPtr("node03"),
		ExitCode:                &exitCode,
	}
}

// createArrayTestJob creates an array test job
func createArrayTestJob(now time.Time, startTime time.Time) slurm.Job {
	return slurm.Job{
		JobID:                   int32Ptr(12349),
		Name:                    strPtr("array-job"),
		JobState:                []api.JobState{api.JobStateRunning},
		UserID:                  int32Ptr(1002),
		UserName:                strPtr("user2"),
		GroupID:                 int32Ptr(1002),
		GroupName:               strPtr("group2"),
		Partition:               strPtr("compute"),
		Priority:                uint32Ptr(600),
		SubmitTime:              now.Add(-1 * time.Hour),
		StartTime:               startTime,
		CPUs:                    uint32Ptr(2),
		MemoryPerNode:           uint64Ptr(4096), // 4GB in MB
		TimeLimit:               uint32Ptr(1 * 60), // 1 hour in minutes
		CurrentWorkingDirectory: strPtr("/home/user2"),
		Command:                 strPtr("array-job.sh"),
		Nodes:                   strPtr("node04"),
		ArrayJobID:              uint32Ptr(12349),
		ArrayTaskID:             uint32Ptr(1),
		ArrayTaskString:         strPtr("1-10"),
	}
}

// GetTestJobs returns test job data
func GetTestJobs() []slurm.Job {
	now := time.Now()
	return []slurm.Job{
		createRunningTestJob(now, now.Add(-1*time.Hour)),
		createPendingTestJob(now),
		createCompletedTestJob(now, now.Add(-3*time.Hour), now.Add(-1*time.Hour)),
		createFailedTestJob(now, now.Add(-90*time.Minute), now.Add(-30*time.Minute)),
		createArrayTestJob(now, now.Add(-45*time.Minute)),
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
		if len(job.JobState) > 0 && job.JobState[0] == api.JobStateRunning {
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

	states := []api.JobState{api.JobStatePending, api.JobStateRunning, api.JobStateCompleted, api.JobStateFailed, api.JobStateCancelled}
	partitions := []string{"compute", "gpu", "bigmem", "debug", "interactive"}
	users := []string{"user1", "user2", "user3", "user4", "user5", "admin", "researcher"}

	for i := 0; i < count; i++ {
		jobID := int32(100000 + i)
		state := states[i%len(states)]
		partition := partitions[i%len(partitions)]
		user := users[i%len(users)]

		// Create realistic job with varying resource requirements
		cpus := uint32(1 + (i % 64))                     // 1-64 CPUs
		memory := uint64(1024 + (i%32)*1024)             // 1GB-32GB memory
		runtime := time.Duration(i%3600) * time.Second   // 0-1 hour runtime

		startTime := time.Now().Add(-runtime)
		endTime := time.Now()

		job := slurm.Job{
			JobID:                   &jobID,
			Name:                    strPtr(fmt.Sprintf("job_%d", i)),
			JobState:                []api.JobState{state},
			Partition:               &partition,
			UserID:                  int32Ptr(int32(1000 + i%100)),
			UserName:                &user,
			GroupID:                 int32Ptr(int32(1000 + i%10)),
			GroupName:               strPtr(fmt.Sprintf("group%d", i%10)),
			Priority:                uint32Ptr(uint32(1000 + (i % 1000))),
			CPUs:                    &cpus,
			MemoryPerNode:           &memory,
			TimeLimit:               uint32Ptr(24 * 60), // 24 hours in minutes
			SubmitTime:              time.Now().Add(-time.Duration(i) * time.Minute),
			CurrentWorkingDirectory: strPtr(fmt.Sprintf("/home/%s", user)),
			Command:                 strPtr(fmt.Sprintf("job_%d.sh", i)),
			Nodes:                   strPtr(fmt.Sprintf("node%02d", (i%10)+1)),
		}

		// Add start/end times for non-pending jobs
		if state != api.JobStatePending {
			job.StartTime = startTime
		}
		if state == api.JobStateCompleted || state == api.JobStateFailed || state == api.JobStateCancelled {
			job.EndTime = endTime
		}

		// Add some variety to failed jobs
		if state == api.JobStateFailed {
			job.ExitCode = &api.ExitCode{ReturnCode: uint32Ptr(1)}
		} else {
			job.ExitCode = &api.ExitCode{ReturnCode: uint32Ptr(0)}
		}

		// Add array job information for some jobs
		if i%10 == 0 {
			arrayJobID := uint32(100000 + i)
			job.ArrayJobID = &arrayJobID
			job.ArrayTaskID = uint32Ptr(uint32(i % 5))
		}

		jobs[i] = job
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}

// GenerateJobBatch generates a batch of jobs with specific characteristics
func GenerateJobBatch(count int, stateFilter api.JobState, partitionFilter string) *slurm.JobList {
	jobs := make([]slurm.Job, count)

	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i := 0; i < count; i++ {
		jobID := int32(200000 + i)
		user := users[i%len(users)]

		startTime := time.Now().Add(-time.Duration(i) * time.Second)
		endTime := time.Now()

		cpus := uint32(1 + (i % 32))
		memory := uint64(2048 + (i%16)*1024)

		job := slurm.Job{
			JobID:                   &jobID,
			Name:                    strPtr(fmt.Sprintf("batch_job_%d", i)),
			JobState:                []api.JobState{stateFilter},
			Partition:               &partitionFilter,
			UserID:                  int32Ptr(int32(2000 + i%50)),
			UserName:                &user,
			GroupID:                 int32Ptr(int32(2000 + i%5)),
			GroupName:               strPtr(fmt.Sprintf("group%d", i%5)),
			Priority:                uint32Ptr(uint32(500 + i)),
			CPUs:                    &cpus,
			MemoryPerNode:           &memory,
			TimeLimit:               uint32Ptr(12 * 60), // 12 hours in minutes
			SubmitTime:              time.Now().Add(-time.Duration(i) * time.Second),
			CurrentWorkingDirectory: strPtr(fmt.Sprintf("/home/%s", user)),
			Command:                 strPtr(fmt.Sprintf("batch_job_%d.sh", i)),
			Nodes:                   strPtr(fmt.Sprintf("node%02d", (i%20)+1)),
			ExitCode:                &api.ExitCode{ReturnCode: uint32Ptr(0)},
		}

		// Add start/end times based on state
		if stateFilter != api.JobStatePending {
			job.StartTime = startTime
		}
		if stateFilter == api.JobStateCompleted || stateFilter == api.JobStateFailed || stateFilter == api.JobStateCancelled {
			job.EndTime = endTime
		}

		jobs[i] = job
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}
