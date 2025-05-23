package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceBenchmarks defines performance requirements
type PerformanceBenchmarks struct {
	StartupTime        time.Duration // Maximum startup time
	ListResponseTime   time.Duration // Maximum time to list notifications
	FilterResponseTime time.Duration // Maximum time to filter notifications
	SearchResponseTime time.Duration // Maximum time to search notifications
	MemoryUsage        int64         // Maximum memory usage in MB
	CPUUsage           float64       // Maximum CPU usage percentage
}

// DefaultBenchmarks returns the default performance benchmarks
func DefaultBenchmarks() PerformanceBenchmarks {
	return PerformanceBenchmarks{
		StartupTime:        2 * time.Second,
		ListResponseTime:   5 * time.Second,
		FilterResponseTime: 3 * time.Second,
		SearchResponseTime: 4 * time.Second,
		MemoryUsage:        100, // 100MB
		CPUUsage:           80.0, // 80%
	}
}

// TestPerformanceBenchmarks tests performance against defined benchmarks
func TestPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN environment variable required for performance tests")
	}

	benchmarks := DefaultBenchmarks()
	tmpDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	// Authenticate first
	authenticateForTesting(t, binaryPath, token)

	t.Run("Startup Performance", func(t *testing.T) {
		testStartupPerformance(t, binaryPath, benchmarks)
	})

	t.Run("List Performance", func(t *testing.T) {
		testListPerformance(t, binaryPath, benchmarks)
	})

	t.Run("Filter Performance", func(t *testing.T) {
		testFilterPerformance(t, binaryPath, benchmarks)
	})

	t.Run("Search Performance", func(t *testing.T) {
		testSearchPerformance(t, binaryPath, benchmarks)
	})

	t.Run("Memory Usage", func(t *testing.T) {
		testMemoryUsage(t, binaryPath, benchmarks)
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		testConcurrentOperations(t, binaryPath, benchmarks)
	})

	t.Run("Large Dataset Performance", func(t *testing.T) {
		testLargeDatasetPerformance(t, binaryPath, benchmarks)
	})
}

func authenticateForTesting(t *testing.T, binaryPath, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", token)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Authentication should succeed: %s", output)
}

func testStartupPerformance(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	// Test cold start performance
	measurements := make([]time.Duration, 5)
	
	for i := 0; i < 5; i++ {
		start := time.Now()
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, binaryPath, "--version")
		
		err := cmd.Run()
		cancel()
		
		elapsed := time.Since(start)
		measurements[i] = elapsed
		
		require.NoError(t, err, "Version command should succeed")
	}

	// Calculate average startup time
	var total time.Duration
	for _, measurement := range measurements {
		total += measurement
	}
	avgStartupTime := total / time.Duration(len(measurements))

	t.Logf("Average startup time: %v", avgStartupTime)
	assert.LessOrEqual(t, avgStartupTime, benchmarks.StartupTime, 
		"Startup time should be less than %v, got %v", benchmarks.StartupTime, avgStartupTime)
}

func testListPerformance(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	measurements := make([]time.Duration, 3)
	
	for i := 0; i < 3; i++ {
		start := time.Now()
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", "100")
		
		err := cmd.Run()
		cancel()
		
		elapsed := time.Since(start)
		measurements[i] = elapsed
		
		require.NoError(t, err, "List command should succeed")
	}

	// Calculate average response time
	var total time.Duration
	for _, measurement := range measurements {
		total += measurement
	}
	avgResponseTime := total / time.Duration(len(measurements))

	t.Logf("Average list response time: %v", avgResponseTime)
	assert.LessOrEqual(t, avgResponseTime, benchmarks.ListResponseTime,
		"List response time should be less than %v, got %v", benchmarks.ListResponseTime, avgResponseTime)
}

func testFilterPerformance(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	filters := []string{
		"is:unread",
		"type:PullRequest",
		"is:unread AND type:PullRequest",
		"repo:owner/repo OR type:Issue",
	}

	for _, filter := range filters {
		t.Run(fmt.Sprintf("Filter_%s", strings.ReplaceAll(filter, " ", "_")), func(t *testing.T) {
			start := time.Now()
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, binaryPath, "list", "--filter", filter, "--limit", "50")
			
			err := cmd.Run()
			cancel()
			
			elapsed := time.Since(start)
			
			require.NoError(t, err, "Filter command should succeed")
			
			t.Logf("Filter '%s' response time: %v", filter, elapsed)
			assert.LessOrEqual(t, elapsed, benchmarks.FilterResponseTime,
				"Filter response time should be less than %v, got %v", benchmarks.FilterResponseTime, elapsed)
		})
	}
}

func testSearchPerformance(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	searchTerms := []string{
		"bug",
		"feature",
		"fix",
		"update",
	}

	for _, term := range searchTerms {
		t.Run(fmt.Sprintf("Search_%s", term), func(t *testing.T) {
			start := time.Now()
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, binaryPath, "search", term, "--limit", "20")
			
			err := cmd.Run()
			cancel()
			
			elapsed := time.Since(start)
			
			require.NoError(t, err, "Search command should succeed")
			
			t.Logf("Search '%s' response time: %v", term, elapsed)
			assert.LessOrEqual(t, elapsed, benchmarks.SearchResponseTime,
				"Search response time should be less than %v, got %v", benchmarks.SearchResponseTime, elapsed)
		})
	}
}

func testMemoryUsage(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	// Start the profiling server
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "profile", "--memory", "--duration", "30")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("Memory profiling not available: %s", output)
		t.Skip("Memory profiling not available")
	}

	// Parse memory usage from output
	lines := strings.Split(string(output), "\n")
	var memoryUsage int64
	
	for _, line := range lines {
		if strings.Contains(line, "Memory usage:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if usage, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
					memoryUsage = usage / (1024 * 1024) // Convert to MB
					break
				}
			}
		}
	}

	if memoryUsage > 0 {
		t.Logf("Memory usage: %d MB", memoryUsage)
		assert.LessOrEqual(t, memoryUsage, benchmarks.MemoryUsage,
			"Memory usage should be less than %d MB, got %d MB", benchmarks.MemoryUsage, memoryUsage)
	} else {
		t.Log("Could not parse memory usage from profiling output")
	}
}

func testConcurrentOperations(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	// Test concurrent list operations
	concurrency := 5
	results := make(chan error, concurrency)

	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", "10")
			err := cmd.Run()
			results <- err
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < concurrency; i++ {
		err := <-results
		assert.NoError(t, err, "Concurrent operation %d should succeed", i)
	}

	elapsed := time.Since(start)
	t.Logf("Concurrent operations completed in: %v", elapsed)
	
	// Should complete within reasonable time even with concurrency
	assert.LessOrEqual(t, elapsed, benchmarks.ListResponseTime*2,
		"Concurrent operations should complete within %v, got %v", benchmarks.ListResponseTime*2, elapsed)
}

func testLargeDatasetPerformance(t *testing.T, binaryPath string, benchmarks PerformanceBenchmarks) {
	// Test with larger datasets
	limits := []int{100, 500, 1000}
	
	for _, limit := range limits {
		t.Run(fmt.Sprintf("Limit_%d", limit), func(t *testing.T) {
			start := time.Now()
			
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", strconv.Itoa(limit), "--format", "json")
			
			output, err := cmd.CombinedOutput()
			cancel()
			
			elapsed := time.Since(start)
			
			require.NoError(t, err, "Large dataset command should succeed")
			
			// Verify we got JSON output
			var notifications []map[string]interface{}
			err = json.Unmarshal(output, &notifications)
			assert.NoError(t, err, "Output should be valid JSON")
			
			t.Logf("Processed %d notifications in %v", len(notifications), elapsed)
			
			// Performance should scale reasonably
			expectedTime := benchmarks.ListResponseTime * time.Duration(limit/100)
			if expectedTime > 30*time.Second {
				expectedTime = 30 * time.Second
			}
			
			assert.LessOrEqual(t, elapsed, expectedTime,
				"Large dataset processing should complete within %v, got %v", expectedTime, elapsed)
		})
	}
}

// TestResourceUsage tests resource usage under various conditions
func TestResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource usage tests in short mode")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN environment variable required for resource tests")
	}

	tmpDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)
	authenticateForTesting(t, binaryPath, token)

	t.Run("Cache Performance", func(t *testing.T) {
		testCachePerformance(t, binaryPath)
	})

	t.Run("Network Efficiency", func(t *testing.T) {
		testNetworkEfficiency(t, binaryPath)
	})

	t.Run("File I/O Performance", func(t *testing.T) {
		testFileIOPerformance(t, binaryPath)
	})
}

func testCachePerformance(t *testing.T, binaryPath string) {
	// First run (cold cache)
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", "50")
	err := cmd.Run()
	cancel()
	coldTime := time.Since(start)
	
	require.NoError(t, err, "Cold cache run should succeed")

	// Second run (warm cache)
	start = time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--limit", "50")
	err = cmd.Run()
	cancel()
	warmTime := time.Since(start)
	
	require.NoError(t, err, "Warm cache run should succeed")

	t.Logf("Cold cache time: %v, Warm cache time: %v", coldTime, warmTime)
	
	// Warm cache should be significantly faster
	assert.Less(t, warmTime, coldTime, "Warm cache should be faster than cold cache")
	assert.Less(t, warmTime, coldTime/2, "Warm cache should be at least 50%% faster")
}

func testNetworkEfficiency(t *testing.T, binaryPath string) {
	// Test with different page sizes to verify efficient API usage
	pageSizes := []int{10, 50, 100}
	
	for _, size := range pageSizes {
		t.Run(fmt.Sprintf("PageSize_%d", size), func(t *testing.T) {
			start := time.Now()
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", strconv.Itoa(size))
			
			err := cmd.Run()
			cancel()
			
			elapsed := time.Since(start)
			
			require.NoError(t, err, "Network request should succeed")
			
			t.Logf("Page size %d completed in %v", size, elapsed)
			
			// Larger page sizes should not be proportionally slower
			// (indicating efficient batching)
			maxExpectedTime := 10 * time.Second
			assert.LessOrEqual(t, elapsed, maxExpectedTime,
				"Network request should complete within %v, got %v", maxExpectedTime, elapsed)
		})
	}
}

func testFileIOPerformance(t *testing.T, binaryPath string) {
	tmpDir, err := os.MkdirTemp("", "gh-notif-io-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "output.json")
	
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, binaryPath, "list", "--limit", "100", "--format", "json", "--output", outputFile)
	
	err = cmd.Run()
	cancel()
	
	elapsed := time.Since(start)
	
	require.NoError(t, err, "File output should succeed")
	
	// Verify file was created and has content
	info, err := os.Stat(outputFile)
	require.NoError(t, err, "Output file should exist")
	assert.Greater(t, info.Size(), int64(0), "Output file should have content")
	
	t.Logf("File I/O completed in %v, file size: %d bytes", elapsed, info.Size())
	
	// File I/O should be reasonably fast
	assert.LessOrEqual(t, elapsed, 10*time.Second,
		"File I/O should complete within 10 seconds, got %v", elapsed)
}
