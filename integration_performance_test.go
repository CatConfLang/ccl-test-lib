package ccl_test_lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/generator"
	"github.com/tylerbu/ccl-test-lib/loader"
	"github.com/tylerbu/ccl-test-lib/types"
)

// Performance and scalability integration tests
// These tests verify the library can handle larger datasets and concurrent operations

func TestIntegration_LargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create large dataset (1000 source tests, each with 5 validations = 5000 flat tests)
	const numTests = 1000
	const validationsPerTest = 5

	t.Logf("Creating %d source tests with %d validations each", numTests, validationsPerTest)

	sourceTests := make([]loader.CompactTest, numTests)
	for i := 0; i < numTests; i++ {
		sourceTests[i] = loader.CompactTest{
			Name:     fmt.Sprintf("large_test_%d", i),
			Input:    fmt.Sprintf("key_%d = value_%d\ncount_%d = %d\nflag_%d = true", i, i, i, i, i),
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect: []map[string]interface{}{
						{"key": fmt.Sprintf("key_%d", i), "value": fmt.Sprintf("value_%d", i)},
						{"key": fmt.Sprintf("count_%d", i), "value": fmt.Sprintf("%d", i)},
						{"key": fmt.Sprintf("flag_%d", i), "value": "true"},
					},
				},
				{
					Function: "build_hierarchy",
					Expect: map[string]interface{}{
						fmt.Sprintf("key_%d", i):   fmt.Sprintf("value_%d", i),
						fmt.Sprintf("count_%d", i): fmt.Sprintf("%d", i),
						fmt.Sprintf("flag_%d", i):  "true",
					},
				},
				{
					Function: "get_string",
					Args:     []string{fmt.Sprintf("key_%d", i)},
					Expect:   fmt.Sprintf("value_%d", i),
				},
				{
					Function: "get_int",
					Args:     []string{fmt.Sprintf("count_%d", i)},
					Expect:   i,
				},
				{
					Function: "get_bool",
					Args:     []string{fmt.Sprintf("flag_%d", i)},
					Expect:   true,
				},
			},
		}
	}

	// Write source data in chunks to avoid memory issues
	const chunkSize = 100
	for chunk := 0; chunk < numTests; chunk += chunkSize {
		end := chunk + chunkSize
		if end > numTests {
			end = numTests
		}

		chunkData := sourceTests[chunk:end]
		// Wrap chunk in CompactTestFile structure for correct parsing
		chunkTestFile := loader.CompactTestFile{
			Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
			Tests:  chunkData,
		}
		sourceData, _ := json.MarshalIndent(chunkTestFile, "", "  ")
		sourceFile := filepath.Join(sourceDir, fmt.Sprintf("large_chunk_%d.json", chunk/chunkSize))
		if err := os.WriteFile(sourceFile, sourceData, 0644); err != nil {
			t.Fatalf("Failed to write source chunk %d: %v", chunk/chunkSize, err)
		}
	}

	// Measure generation performance
	start := time.Now()
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
		Verbose:      false, // Disable verbose to avoid spam
	})

	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}
	generationTime := time.Since(start)

	t.Logf("Generation of %d tests took %v", numTests*validationsPerTest, generationTime)

	// Verify generation results
	outputFiles, err := filepath.Glob(filepath.Join(generatedDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list output files: %v", err)
	}

	expectedFiles := (numTests / chunkSize)
	if numTests%chunkSize != 0 {
		expectedFiles++
	}

	if len(outputFiles) != expectedFiles {
		t.Errorf("Expected %d output files, got %d", expectedFiles, len(outputFiles))
	}

	// Measure loading performance
	cfg := config.ImplementationConfig{
		Name:    "large-dataset-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
			config.FunctionGetInt,
			config.FunctionGetBool,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{},
		VariantChoice:   config.VariantProposed,
	}

	start = time.Now()
	tests, err := LoadCompatibleTests(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Loading failed: %v", err)
	}
	loadingTime := time.Since(start)

	t.Logf("Loading of %d tests took %v", len(tests), loadingTime)

	// Verify loading results
	expectedTestCount := numTests * validationsPerTest
	if len(tests) != expectedTestCount {
		t.Errorf("Expected %d loaded tests, got %d", expectedTestCount, len(tests))
	}

	// Measure statistics performance
	start = time.Now()
	stats, err := GetTestStats(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Statistics failed: %v", err)
	}
	statsTime := time.Since(start)

	t.Logf("Statistics calculation took %v", statsTime)

	// Verify statistics
	if stats.TotalTests != expectedTestCount {
		t.Errorf("Expected %d total tests in stats, got %d", expectedTestCount, stats.TotalTests)
	}

	// Performance benchmarks (adjust these thresholds based on actual performance)
	if generationTime > 30*time.Second {
		t.Logf("Warning: Generation took %v, which may be slow for %d tests", generationTime, expectedTestCount)
	}
	if loadingTime > 10*time.Second {
		t.Logf("Warning: Loading took %v, which may be slow for %d tests", loadingTime, len(tests))
	}
	if statsTime > 5*time.Second {
		t.Logf("Warning: Statistics took %v, which may be slow for %d tests", statsTime, len(tests))
	}
}

func TestIntegration_MemoryUsagePattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create tests with large string content to test memory handling
	largeContent := make([]byte, 1024*10) // 10KB string
	for i := range largeContent {
		largeContent[i] = byte('a' + (i % 26))
	}
	largeString := string(largeContent)

	sourceTests := []loader.CompactTest{
		{
			Name:     "large_content_test",
			Input:    fmt.Sprintf("large_key = %s\nother_key = small_value", largeString),
			Features: []string{},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect: []map[string]interface{}{
						{"key": "large_key", "value": largeString},
						{"key": "other_key", "value": "small_value"},
					},
				},
				{
					Function: "get_string",
					Args:     []string{"large_key"},
					Expect:   largeString,
				},
				{
					Function: "build_hierarchy",
					Expect: map[string]interface{}{
						"large_key": largeString,
						"other_key": "small_value",
					},
				},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  sourceTests,
	}
	sourceData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "large_content.json"), sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Test generation with large content
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})

	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Test loading with large content
	cfg := config.ImplementationConfig{
		Name:    "memory-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
		},
		SupportedFeatures: []config.CCLFeature{},
		BehaviorChoices:   []config.CCLBehavior{},
		VariantChoice:     config.VariantProposed,
	}

	tests, err := LoadCompatibleTests(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Loading failed: %v", err)
	}

	// Verify large content is preserved
	if len(tests) != 3 {
		t.Errorf("Expected 3 tests, got %d", len(tests))
	}

	// Find the get_string test and verify large content
	for _, test := range tests {
		if test.Validation == "get_string" {
			if test.Expected != largeString {
				t.Error("Large string content was not preserved correctly")
			}
			if len(test.Args) != 1 || test.Args[0] != "large_key" {
				t.Error("Args not preserved correctly for large content test")
			}
		}
	}
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create test data
	sourceTests := []loader.CompactTest{
		{
			Name:     "concurrent_test",
			Input:    "key = value",
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "key", "value": "value"}}},
				{Function: "get_string", Args: []string{"key"}, Expect: "value"},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  sourceTests,
	}
	sourceData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "concurrent.json"), sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Generate flat format
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	cfg := config.ImplementationConfig{
		Name:    "concurrent-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionGetString,
		},
		SupportedFeatures: []config.CCLFeature{},
		BehaviorChoices:   []config.CCLBehavior{},
		VariantChoice:     config.VariantProposed,
	}

	// Test concurrent loading operations
	const numConcurrent = 10
	resultChan := make(chan []types.TestCase, numConcurrent)
	errorChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			tests, err := LoadCompatibleTests(tmpDir, cfg)
			if err != nil {
				errorChan <- fmt.Errorf("goroutine %d failed: %v", id, err)
				return
			}
			resultChan <- tests
		}(i)
	}

	// Collect results
	var allResults [][]types.TestCase
	for i := 0; i < numConcurrent; i++ {
		select {
		case result := <-resultChan:
			allResults = append(allResults, result)
		case err := <-errorChan:
			t.Fatalf("Concurrent operation failed: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatalf("Concurrent operation timed out")
		}
	}

	// Verify all results are consistent
	if len(allResults) != numConcurrent {
		t.Errorf("Expected %d results, got %d", numConcurrent, len(allResults))
	}

	expectedCount := 2 // parse + get_string
	for i, result := range allResults {
		if len(result) != expectedCount {
			t.Errorf("Result %d has %d tests, expected %d", i, len(result), expectedCount)
		}
	}

	// Test concurrent statistics operations
	statsChan := make(chan types.TestStatistics, numConcurrent)
	statsErrorChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			stats, err := GetTestStats(tmpDir, cfg)
			if err != nil {
				statsErrorChan <- fmt.Errorf("stats goroutine %d failed: %v", id, err)
				return
			}
			statsChan <- stats
		}(i)
	}

	// Collect statistics results
	var allStats []types.TestStatistics
	for i := 0; i < numConcurrent; i++ {
		select {
		case stats := <-statsChan:
			allStats = append(allStats, stats)
		case err := <-statsErrorChan:
			t.Fatalf("Concurrent statistics failed: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatalf("Concurrent statistics timed out")
		}
	}

	// Verify all statistics are consistent
	for i, stats := range allStats {
		if stats.TotalTests != expectedCount {
			t.Errorf("Stats %d has %d total tests, expected %d", i, stats.TotalTests, expectedCount)
		}
		if stats.CompatibleTests != expectedCount {
			t.Errorf("Stats %d has %d compatible tests, expected %d", i, stats.CompatibleTests, expectedCount)
		}
	}
}

func TestIntegration_MixedFormatHandling(t *testing.T) {
	tmpDir := t.TempDir()
	testsDir := filepath.Join(tmpDir, "tests")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create compact format data
	compactTests := []loader.CompactTest{
		{
			Name:     "compact_format_test",
			Input:    "compact_key = compact_value",
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "compact_key", "value": "compact_value"}}},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  compactTests,
	}
	compactData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(testsDir, "compact.json"), compactData, 0644); err != nil {
		t.Fatalf("Failed to write compact file: %v", err)
	}

	// Create flat format data directly
	flatTests := []types.TestCase{
		{
			Name:       "flat_format_test",
			Input:      "flat_key = flat_value",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "flat_key", "value": "flat_value"}},
			Functions:  []string{"parse"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "flat_format_test",
		},
	}

	flatData, _ := json.MarshalIndent(flatTests, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "flat.json"), flatData, 0644); err != nil {
		t.Fatalf("Failed to write flat file: %v", err)
	}

	cfg := config.ImplementationConfig{
		Name:               "mixed-format-test",
		Version:            "v1.0.0",
		SupportedFunctions: []config.CCLFunction{config.FunctionParse},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	// Test loading from mixed directory structure
	testLoader := loader.NewTestLoader(tmpDir, cfg)

	// Load compact format tests
	compactLoadedTests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatCompact,
		FilterMode: loader.FilterAll,
	})
	if err != nil {
		t.Fatalf("Failed to load compact tests: %v", err)
	}

	// Load flat format tests
	flatLoadedTests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterAll,
	})
	if err != nil {
		t.Fatalf("Failed to load flat tests: %v", err)
	}

	// Verify we get different results from different formats
	if len(compactLoadedTests) != 1 {
		t.Errorf("Expected 1 compact test, got %d", len(compactLoadedTests))
	}
	if len(flatLoadedTests) != 1 {
		t.Errorf("Expected 1 flat test, got %d", len(flatLoadedTests))
	}

	// Verify test content differences
	if compactLoadedTests[0].Name != "compact_format_test" {
		t.Errorf("Expected compact test name 'compact_format_test', got %s", compactLoadedTests[0].Name)
	}
	if flatLoadedTests[0].Name != "flat_format_test" {
		t.Errorf("Expected flat test name 'flat_format_test', got %s", flatLoadedTests[0].Name)
	}

	// Test convenience function behavior (should prefer flat format)
	allTests, err := LoadCompatibleTests(tmpDir, cfg)
	if err != nil {
		t.Fatalf("LoadCompatibleTests failed: %v", err)
	}

	// Should find flat format test
	if len(allTests) != 1 {
		t.Errorf("Expected 1 test from LoadCompatibleTests, got %d", len(allTests))
	}
	if allTests[0].Name != "flat_format_test" {
		t.Errorf("LoadCompatibleTests should prefer flat format, got test %s", allTests[0].Name)
	}
}

// Benchmark tests for performance regression detection

func BenchmarkIntegration_LoadingPerformance(b *testing.B) {
	tmpDir := b.TempDir()
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		b.Fatalf("Failed to create directory: %v", err)
	}

	// Create moderate-sized test dataset
	const numTests = 100
	flatTests := make([]types.TestCase, numTests)
	for i := 0; i < numTests; i++ {
		flatTests[i] = types.TestCase{
			Name:       fmt.Sprintf("bench_test_%d", i),
			Input:      fmt.Sprintf("key_%d = value_%d", i, i),
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": fmt.Sprintf("key_%d", i), "value": fmt.Sprintf("value_%d", i)}},
			Functions:  []string{"parse"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: fmt.Sprintf("bench_test_%d", i),
		}
	}

	flatData, _ := json.MarshalIndent(flatTests, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "bench.json"), flatData, 0644); err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	cfg := config.ImplementationConfig{
		Name:               "bench-test",
		Version:            "v1.0.0",
		SupportedFunctions: []config.CCLFunction{config.FunctionParse},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tests, err := LoadCompatibleTests(tmpDir, cfg)
		if err != nil {
			b.Fatalf("LoadCompatibleTests failed: %v", err)
		}
		if len(tests) != numTests {
			b.Fatalf("Expected %d tests, got %d", numTests, len(tests))
		}
	}
}

func BenchmarkIntegration_StatisticsPerformance(b *testing.B) {
	tmpDir := b.TempDir()
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		b.Fatalf("Failed to create directory: %v", err)
	}

	// Create test dataset for statistics
	const numTests = 50
	flatTests := make([]types.TestCase, numTests)
	for i := 0; i < numTests; i++ {
		flatTests[i] = types.TestCase{
			Name:       fmt.Sprintf("stats_test_%d", i),
			Input:      fmt.Sprintf("key_%d = value_%d", i, i),
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": fmt.Sprintf("key_%d", i), "value": fmt.Sprintf("value_%d", i)}},
			Functions:  []string{"parse"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: fmt.Sprintf("stats_test_%d", i),
		}
	}

	flatData, _ := json.MarshalIndent(flatTests, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "stats.json"), flatData, 0644); err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	cfg := config.ImplementationConfig{
		Name:               "stats-bench-test",
		Version:            "v1.0.0",
		SupportedFunctions: []config.CCLFunction{config.FunctionParse},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats, err := GetTestStats(tmpDir, cfg)
		if err != nil {
			b.Fatalf("GetTestStats failed: %v", err)
		}
		if stats.TotalTests != numTests {
			b.Fatalf("Expected %d total tests, got %d", numTests, stats.TotalTests)
		}
	}
}
