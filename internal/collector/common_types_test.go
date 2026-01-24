// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"testing"
	"time"
)

func TestAccountHierarchy(t *testing.T) {
	t.Parallel()
	t.Run("EmptyHierarchy", func(t *testing.T) {
		t.Parallel()
		h := &AccountHierarchy{
			AccountMap:       make(map[string]*AccountNode),
			RootAccounts:     []*AccountNode{},
			LastUpdated:      time.Now(),
			HierarchyType:    "account",
			ValidationErrors: []string{},
		}

		if h.TotalAccounts != 0 {
			t.Errorf("Expected 0 total accounts, got %d", h.TotalAccounts)
		}
		if h.BalanceScore < 0 || h.BalanceScore > 1 {
			t.Errorf("BalanceScore should be between 0 and 1, got %f", h.BalanceScore)
		}
	})

	t.Run("HierarchyWithMetrics", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		h := &AccountHierarchy{
			TotalAccounts:       100,
			MaxDepth:            5,
			BalanceScore:        0.85,
			UtilizationScore:    0.75,
			TotalShares:         10000,
			TotalRawUsage:       8000,
			EffectiveUsage:      7500,
			NormalizedShares:    10000.0,
			LeafAccounts:        80,
			BranchAccounts:      20,
			OrphanAccounts:      0,
			CircularRefs:        0,
			ShareGini:           0.3,
			ShareEntropy:        4.5,
			ShareSkewness:       0.2,
			ValidationErrors:    []string{},
			Warnings:            []string{"low usage"},
			LastUpdated:         now,
			EfficiencyScore:     0.8,
			CoverageScore:       0.9,
			DataCompleteness:    0.95,
			DataConsistency:     0.88,
			HierarchyType:       "account",
			TotalUsers:          500,
			AccountMap:          make(map[string]*AccountNode),
			OrganizationalUnits: map[string]*OrganizationalUnit{},
		}

		if h.TotalAccounts != 100 {
			t.Errorf("Expected 100 total accounts, got %d", h.TotalAccounts)
		}
		if h.MaxDepth != 5 {
			t.Errorf("Expected max depth 5, got %d", h.MaxDepth)
		}
		if h.BalanceScore != 0.85 {
			t.Errorf("Expected balance score 0.85, got %f", h.BalanceScore)
		}
		if h.UtilizationScore != 0.75 {
			t.Errorf("Expected utilization score 0.75, got %f", h.UtilizationScore)
		}
		if h.TotalShares != 10000 {
			t.Errorf("Expected total shares 10000, got %d", h.TotalShares)
		}
		if h.LeafAccounts != 80 {
			t.Errorf("Expected 80 leaf accounts, got %d", h.LeafAccounts)
		}
		if h.BranchAccounts != 20 {
			t.Errorf("Expected 20 branch accounts, got %d", h.BranchAccounts)
		}
		if h.TotalUsers != 500 {
			t.Errorf("Expected 500 total users, got %d", h.TotalUsers)
		}
		if h.HierarchyType != "account" {
			t.Errorf("Expected hierarchy type 'account', got '%s'", h.HierarchyType)
		}
	})
}

func TestAccountNode(t *testing.T) {
	t.Parallel()
	t.Run("BasicNode", func(t *testing.T) {
		t.Parallel()
		node := &AccountNode{
			Name:              "root",
			AccountName:       "root",
			Description:       "Root account",
			Parent:            "",
			ParentAccount:     "",
			Children:          []*AccountNode{},
			ChildAccounts:     []*AccountNode{},
			Level:             0,
			Path:              "/root",
			Status:            "active",
			ActiveUsers:       10,
			CPULimit:          1000,
			MemoryLimit:       100 * 1024 * 1024 * 1024, // 100GB
			GPULimit:          10,
			WallTimeLimit:     8 * 60 * 60, // 8 hours
			RawUsage:          800,
			EffectiveUsage:    750,
			NormalizedUsage:   0.8,
			QuotaAllocated:    1000.0,
			QuotaUsed:         800.0,
			QuotaRemaining:    200.0,
			RawShares:         100,
			NormalizedShares:  100.0,
			EffectiveShares:   80.0,
			FairShareValue:    0.8,
			Priority:          1000,
			JobCount:          50,
			UserCount:         10,
			DirectUsers:       5,
			SubAccountCount:   2,
			TotalCPUHours:     1000.5,
			TotalGPUHours:     50.25,
			AvgJobSize:        20.0,
			AvgWaitTime:       5 * time.Minute,
			CPUEfficiency:     0.85,
			MemoryEfficiency:  0.9,
			OverallEfficiency: 0.875,
			IsActive:          true,
			IsDefault:         true,
			IsRoot:            true,
			IsLeaf:            false,
			HasQuota:          true,
			HasDeficit:        false,
			ValidationStatus:  "valid",
			ValidationErrors:  []string{},
			ActiveUserCount:   10,
			DataQuality:       0.95,
			AllUsers:          []string{"user1", "user2", "user3"},
		}

		if node.Name != "root" {
			t.Errorf("Expected name 'root', got '%s'", node.Name)
		}
		if node.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", node.Status)
		}
		if !node.IsActive {
			t.Error("Expected IsActive to be true")
		}
		if !node.IsRoot {
			t.Error("Expected IsRoot to be true")
		}
		if node.IsLeaf {
			t.Error("Expected IsLeaf to be false")
		}
		if node.ActiveUsers != 10 {
			t.Errorf("Expected 10 active users, got %d", node.ActiveUsers)
		}
		if node.CPULimit != 1000 {
			t.Errorf("Expected CPU limit 1000, got %d", node.CPULimit)
		}
		if node.MemoryLimit != 100*1024*1024*1024 {
			t.Errorf("Expected memory limit %d, got %d", 100*1024*1024*1024, node.MemoryLimit)
		}
		if node.GPULimit != 10 {
			t.Errorf("Expected GPU limit 10, got %d", node.GPULimit)
		}
		if node.JobCount != 50 {
			t.Errorf("Expected job count 50, got %d", node.JobCount)
		}
		if node.TotalCPUHours != 1000.5 {
			t.Errorf("Expected total CPU hours 1000.5, got %f", node.TotalCPUHours)
		}
		if node.AvgWaitTime != 5*time.Minute {
			t.Errorf("Expected avg wait time 5m, got %v", node.AvgWaitTime)
		}
	})

	t.Run("NodeWithChildren", func(t *testing.T) {
		t.Parallel()
		child1 := &AccountNode{Name: "child1"}
		child2 := &AccountNode{Name: "child2"}
		parent := &AccountNode{
			Name:            "parent",
			Children:        []*AccountNode{child1, child2},
			ChildAccounts:   []*AccountNode{child1, child2},
			SubAccountCount: 2,
		}

		if len(parent.Children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(parent.Children))
		}
		if parent.SubAccountCount != 2 {
			t.Errorf("Expected sub account count 2, got %d", parent.SubAccountCount)
		}
	})

	t.Run("NodeWithQuotaDeficit", func(t *testing.T) {
		t.Parallel()
		node := &AccountNode{
			Name:           "overused",
			QuotaUsed:      1200.0,
			QuotaAllocated: 1000.0,
			QuotaRemaining: 0.0,
			HasQuota:       true,
			HasDeficit:     true,
		}

		if !node.HasDeficit {
			t.Error("Expected HasDeficit to be true")
		}
		if node.QuotaRemaining != 0.0 {
			t.Errorf("Expected quota remaining 0.0, got %f", node.QuotaRemaining)
		}
	})
}

func TestAccountUsageInfo(t *testing.T) {
	t.Parallel()
	t.Run("BasicUsage", func(t *testing.T) {
		t.Parallel()
		usage := &AccountUsageInfo{
			CPUHours:           1000.5,
			MemoryMBHours:      500000.0,
			ResourceEfficiency: 0.85,
			WasteRatio:         0.15,
			EstimatedCost:      100.50,
		}

		if usage.CPUHours != 1000.5 {
			t.Errorf("Expected CPU hours 1000.5, got %f", usage.CPUHours)
		}
		if usage.MemoryMBHours != 500000.0 {
			t.Errorf("Expected memory MB hours 500000.0, got %f", usage.MemoryMBHours)
		}
		if usage.ResourceEfficiency != 0.85 {
			t.Errorf("Expected resource efficiency 0.85, got %f", usage.ResourceEfficiency)
		}
		if usage.WasteRatio != 0.15 {
			t.Errorf("Expected waste ratio 0.15, got %f", usage.WasteRatio)
		}
		if usage.EstimatedCost != 100.50 {
			t.Errorf("Estimated cost 100.50, got %f", usage.EstimatedCost)
		}
	})

	t.Run("HighEfficiencyUsage", func(t *testing.T) {
		t.Parallel()
		usage := &AccountUsageInfo{
			CPUHours:           500.0,
			MemoryMBHours:      250000.0,
			ResourceEfficiency: 0.95,
			WasteRatio:         0.05,
			EstimatedCost:      50.0,
		}

		if usage.ResourceEfficiency != 0.95 {
			t.Errorf("Expected high efficiency 0.95, got %f", usage.ResourceEfficiency)
		}
		if usage.WasteRatio > 0.10 {
			t.Errorf("Expected low waste ratio <= 0.10, got %f", usage.WasteRatio)
		}
	})
}

func TestQuotaViolation(t *testing.T) {
	t.Parallel()
	t.Run("BasicViolation", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		violation := &QuotaViolation{
			Account:            "test-account",
			UserName:           "user1",
			JobID:              "12345",
			Timestamp:          now,
			Type:               "CPU",
			ViolationType:      "CPU",
			Severity:           "Critical",
			Limit:              1000.0,
			Requested:          1200.0,
			Current:            950.0,
			Message:            "CPU quota exceeded",
			ResourceType:       "CPU",
			QuotaType:          "user",
			ResolutionRequired: true,
			ResolutionActions:  []string{"Increase quota", "Reduce job size"},
			AutoResolved:       false,
			RecurrenceCount:    1,
			FirstOccurrence:    now.Add(-24 * time.Hour),
			LastOccurrence:     now,
		}

		if violation.Account != "test-account" {
			t.Errorf("Expected account 'test-account', got '%s'", violation.Account)
		}
		if violation.UserName != "user1" {
			t.Errorf("Expected username 'user1', got '%s'", violation.UserName)
		}
		if violation.JobID != "12345" {
			t.Errorf("Expected job ID '12345', got '%s'", violation.JobID)
		}
		if violation.Severity != "Critical" {
			t.Errorf("Expected severity 'Critical', got '%s'", violation.Severity)
		}
		if violation.Limit != 1000.0 {
			t.Errorf("Expected limit 1000.0, got %f", violation.Limit)
		}
		if violation.Requested != 1200.0 {
			t.Errorf("Expected requested 1200.0, got %f", violation.Requested)
		}
		if !violation.ResolutionRequired {
			t.Error("Expected ResolutionRequired to be true")
		}
		if violation.RecurrenceCount != 1 {
			t.Errorf("Expected recurrence count 1, got %d", violation.RecurrenceCount)
		}
	})

	t.Run("MemoryViolation", func(t *testing.T) {
		t.Parallel()
		violation := &QuotaViolation{
			Account:            "test-account",
			Type:               "Memory",
			ViolationType:      "Memory",
			Severity:           "Warning",
			Limit:              100 * 1024 * 1024 * 1024, // 100GB
			Requested:          120 * 1024 * 1024 * 1024, // 120GB
			Current:            95 * 1024 * 1024 * 1024,  // 95GB
			Message:            "Memory quota approaching limit",
			ResourceType:       "Memory",
			QuotaType:          "account",
			ResolutionRequired: false,
			RecurrenceCount:    5,
			RecurrencePattern:  "daily",
		}

		if violation.Type != "Memory" {
			t.Errorf("Expected type 'Memory', got '%s'", violation.Type)
		}
		if violation.Severity != "Warning" {
			t.Errorf("Expected severity 'Warning', got '%s'", violation.Severity)
		}
		if violation.RecurrenceCount != 5 {
			t.Errorf("Expected recurrence count 5, got %d", violation.RecurrenceCount)
		}
		if violation.RecurrencePattern != "daily" {
			t.Errorf("Expected recurrence pattern 'daily', got '%s'", violation.RecurrencePattern)
		}
	})
}

func TestQuotaRecommendation(t *testing.T) {
	t.Parallel()
	t.Run("IncreaseQuotaRecommendation", func(t *testing.T) {
		t.Parallel()
		rec := &QuotaRecommendation{
			Account:             "test-account",
			Type:                "CPU",
			Priority:            "High",
			CurrentValue:        1000.0,
			RecommendedValue:    1500.0,
			Reason:              "Average usage exceeds current quota",
			PotentialSavings:    0.0,
			EstimatedImpact:     "Reduce queue wait time by 30%",
			ImplementationSteps: []string{"Update SLURM configuration", "Test new quota"},
			Prerequisites:       []string{"Approval from admin", "Resource availability"},
			RiskAssessment:      "Low",
			Confidence:          0.9,
			HistoricalSuccess:   0.95,
			ExpectedROI:         0.85,
		}

		if rec.Account != "test-account" {
			t.Errorf("Expected account 'test-account', got '%s'", rec.Account)
		}
		if rec.Type != "CPU" {
			t.Errorf("Expected type 'CPU', got '%s'", rec.Type)
		}
		if rec.Priority != "High" {
			t.Errorf("Expected priority 'High', got '%s'", rec.Priority)
		}
		if rec.CurrentValue != 1000.0 {
			t.Errorf("Expected current value 1000.0, got %f", rec.CurrentValue)
		}
		if rec.RecommendedValue != 1500.0 {
			t.Errorf("Expected recommended value 1500.0, got %f", rec.RecommendedValue)
		}
		if rec.Confidence != 0.9 {
			t.Errorf("Expected confidence 0.9, got %f", rec.Confidence)
		}
		if rec.HistoricalSuccess != 0.95 {
			t.Errorf("Expected historical success 0.95, got %f", rec.HistoricalSuccess)
		}
	})

	t.Run("ReduceQuotaRecommendation", func(t *testing.T) {
		t.Parallel()
		rec := &QuotaRecommendation{
			Account:             "over-provisioned-account",
			Type:                "Memory",
			Priority:            "Low",
			CurrentValue:        500 * 1024 * 1024 * 1024, // 500GB
			RecommendedValue:    250 * 1024 * 1024 * 1024, // 250GB
			Reason:              "Historical usage is significantly below quota",
			PotentialSavings:    100.0,
			EstimatedImpact:     "Reduce wasted allocation",
			ImplementationSteps: []string{"Reduce quota in SLURM", "Monitor usage"},
			RiskAssessment:      "Very Low",
			Confidence:          0.95,
			HistoricalSuccess:   0.98,
			ExpectedROI:         0.9,
		}

		if rec.CurrentValue != float64(500*1024*1024*1024) {
			t.Errorf("Expected current value %f, got %f", float64(500*1024*1024*1024), rec.CurrentValue)
		}
		if rec.RecommendedValue != float64(250*1024*1024*1024) {
			t.Errorf("Expected recommended value %f, got %f", float64(250*1024*1024*1024), rec.RecommendedValue)
		}
		if rec.PotentialSavings != 100.0 {
			t.Errorf("Expected potential savings 100.0, got %f", rec.PotentialSavings)
		}
	})
}

func TestOptimizationOpportunity(t *testing.T) {
	t.Parallel()
	t.Run("JobSizeOptimization", func(t *testing.T) {
		t.Parallel()
		opp := &OptimizationOpportunity{
			Type:             "Job Size",
			Description:      "Many jobs are using more resources than necessary",
			Account:          "test-account",
			PotentialSavings: 50.0,
			Implementation:   "Educate users about efficient resource allocation",
			Priority:         4,
			Effort:           "Medium",
			Actions:          []string{"Create job size guidelines", "Monitor efficiency"},
			ResourceImpact:   map[string]float64{"CPU": 30.0, "Memory": 40.0},
			Confidence:       0.85,
			DataPoints:       1000,
			LastValidated:    time.Now(),
		}

		if opp.Type != "Job Size" {
			t.Errorf("Expected type 'Job Size', got '%s'", opp.Type)
		}
		if opp.PotentialSavings != 50.0 {
			t.Errorf("Expected potential savings 50.0, got %f", opp.PotentialSavings)
		}
		if opp.Priority != 4 {
			t.Errorf("Expected priority 4, got %d", opp.Priority)
		}
		if opp.Effort != "Medium" {
			t.Errorf("Expected effort 'Medium', got '%s'", opp.Effort)
		}
		if opp.Confidence != 0.85 {
			t.Errorf("Expected confidence 0.85, got %f", opp.Confidence)
		}
		if opp.DataPoints != 1000 {
			t.Errorf("Expected data points 1000, got %d", opp.DataPoints)
		}
	})
}

func TestEfficiencyMetrics(t *testing.T) {
	t.Parallel()
	t.Run("BasicMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := &EfficiencyMetrics{
			OverallEfficiency:    0.85,
			CPUEfficiency:        0.9,
			MemoryEfficiency:     0.88,
			GPUEfficiency:        0.92,
			ResourceEfficiency:   0.89,
			ThroughputEfficiency: 0.82,
			WasteRatio:           0.15,
			OptimizationScore:    0.75,
			OptimalityScore:      0.8,
			ImprovementPotential: 0.2,
			EfficiencyGrade:      "B",
			EfficiencyCategory:   "Good",
			Recommendations:      []string{"Reduce CPU waste", "Optimize memory usage"},
		}

		if metrics.OverallEfficiency != 0.85 {
			t.Errorf("Expected overall efficiency 0.85, got %f", metrics.OverallEfficiency)
		}
		if metrics.CPUEfficiency != 0.9 {
			t.Errorf("Expected CPU efficiency 0.9, got %f", metrics.CPUEfficiency)
		}
		if metrics.MemoryEfficiency != 0.88 {
			t.Errorf("Expected memory efficiency 0.88, got %f", metrics.MemoryEfficiency)
		}
		if metrics.EfficiencyGrade != "B" {
			t.Errorf("Expected grade 'B', got '%s'", metrics.EfficiencyGrade)
		}
		if metrics.EfficiencyCategory != "Good" {
			t.Errorf("Expected category 'Good', got '%s'", metrics.EfficiencyCategory)
		}
	})

	t.Run("WithResourceWaste", func(t *testing.T) {
		t.Parallel()
		waste := &ResourceWaste{
			WastedCPUHours:      100.5,
			WastedMemoryGBHours: 50.0,
			WastedMemoryGB:      100.0,
			WastedGPUHours:      10.25,
			TotalWasteCost:      25.50,
			WastePercentage:     0.15,
			WasteCost:           25.50,
			IdleTime:            500.0,
		}

		metrics := &EfficiencyMetrics{
			OverallEfficiency:  0.8,
			CPUEfficiency:      0.85,
			MemoryEfficiency:   0.82,
			ResourceEfficiency: 0.835,
			ResourceWaste:      waste,
			WasteRatio:         0.2,
			OptimizationScore:  0.7,
			EfficiencyGrade:    "C",
			EfficiencyCategory: "Fair",
		}

		if metrics.ResourceWaste == nil {
			t.Error("Expected ResourceWaste to be set")
		}
		if metrics.ResourceWaste.WastedCPUHours != 100.5 {
			t.Errorf("Expected wasted CPU hours 100.5, got %f", metrics.ResourceWaste.WastedCPUHours)
		}
	})
}

func TestTimeSeriesPoint(t *testing.T) {
	t.Parallel()
	t.Run("BasicPoint", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		point := &TimeSeriesPoint{
			Timestamp:       now,
			CPUHours:        1000.5,
			MemoryGB:        500.0,
			GPUHours:        50.25,
			JobCount:        100,
			UserCount:       10,
			SubmissionCount: 150,
			WaitTime:        5.5,
			RunTime:         10.0,
			CostUSD:         100.50,
		}

		if point.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
		if point.CPUHours != 1000.5 {
			t.Errorf("Expected CPU hours 1000.5, got %f", point.CPUHours)
		}
		if point.MemoryGB != 500.0 {
			t.Errorf("Expected memory GB 500.0, got %f", point.MemoryGB)
		}
		if point.GPUHours != 50.25 {
			t.Errorf("Expected GPU hours 50.25, got %f", point.GPUHours)
		}
		if point.JobCount != 100 {
			t.Errorf("Expected job count 100, got %d", point.JobCount)
		}
		if point.UserCount != 10 {
			t.Errorf("Expected user count 10, got %d", point.UserCount)
		}
		if point.WaitTime != 5.5 {
			t.Errorf("Expected wait time 5.5, got %f", point.WaitTime)
		}
		if point.RunTime != 10.0 {
			t.Errorf("Expected run time 10.0, got %f", point.RunTime)
		}
		if point.CostUSD != 100.50 {
			t.Errorf("Expected cost USD 100.50, got %f", point.CostUSD)
		}
	})
}

func TestPerformanceMetrics(t *testing.T) {
	t.Parallel()
	t.Run("BasicMetrics", func(t *testing.T) {
		t.Parallel()
		metrics := &PerformanceMetrics{
			Throughput:     100.5,
			Latency:        5.2,
			QueueTime:      10.0,
			ExecutionTime:  15.5,
			TurnaroundTime: 25.5,
			SuccessRate:    0.95,
			ErrorRate:      0.05,
		}

		if metrics.Throughput != 100.5 {
			t.Errorf("Expected throughput 100.5, got %f", metrics.Throughput)
		}
		if metrics.Latency != 5.2 {
			t.Errorf("Expected latency 5.2, got %f", metrics.Latency)
		}
		if metrics.QueueTime != 10.0 {
			t.Errorf("Expected queue time 10.0, got %f", metrics.QueueTime)
		}
		if metrics.ExecutionTime != 15.5 {
			t.Errorf("Expected execution time 15.5, got %f", metrics.ExecutionTime)
		}
		if metrics.TurnaroundTime != 25.5 {
			t.Errorf("Expected turnaround time 25.5, got %f", metrics.TurnaroundTime)
		}
		if metrics.SuccessRate != 0.95 {
			t.Errorf("Expected success rate 0.95, got %f", metrics.SuccessRate)
		}
		if metrics.ErrorRate != 0.05 {
			t.Errorf("Expected error rate 0.05, got %f", metrics.ErrorRate)
		}
	})
}

func TestResourceUsagePattern(t *testing.T) {
	t.Parallel()
	t.Run("BasicPattern", func(t *testing.T) {
		t.Parallel()
		pattern := &ResourceUsagePattern{
			PeakHours:             []int{9, 10, 11, 14, 15, 16},
			OffPeakHours:          []int{0, 1, 2, 3, 4, 22, 23},
			AverageUsage:          0.75,
			PeakUsage:             0.95,
			Variability:           0.2,
			Predictability:        0.85,
			SeasonalTrends:        []string{"Weekly cycle", "Monthly spike"},
			EfficiencyScore:       0.8,
			WastePercentage:       0.15,
			OptimizationPotential: 0.2,
			RecommendedActions:    []string{"Shift workloads", "Increase capacity during peak"},
		}

		if len(pattern.PeakHours) != 6 {
			t.Errorf("Expected 6 peak hours, got %d", len(pattern.PeakHours))
		}
		if len(pattern.OffPeakHours) != 7 {
			t.Errorf("Expected 7 off-peak hours, got %d", len(pattern.OffPeakHours))
		}
		if pattern.AverageUsage != 0.75 {
			t.Errorf("Expected average usage 0.75, got %f", pattern.AverageUsage)
		}
		if pattern.PeakUsage != 0.95 {
			t.Errorf("Expected peak usage 0.95, got %f", pattern.PeakUsage)
		}
		if pattern.EfficiencyScore != 0.8 {
			t.Errorf("Expected efficiency score 0.8, got %f", pattern.EfficiencyScore)
		}
	})
}
