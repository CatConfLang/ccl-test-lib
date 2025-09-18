package ccl_test_lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tylerbu/ccl-test-lib/config"
)

// Integration tests that work with real CCL test data structure
// These tests assume the ccl-test-data repository exists as a sibling directory

func getTestDataPath() string {
	// Try to find ccl-test-data as a sibling directory
	possiblePaths := []string{
		"../ccl-test-data",
		"../../ccl-test-data", // In case we're run from a subdirectory
		"../../../ccl-test-data",
		"/Volumes/Code/claude-workspace-ccl/ccl-test-data", // Absolute path for Claude Code workspace
	}

	for _, path := range possiblePaths {
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			return path
		}
	}
	return ""
}

func TestIntegration_RealCCLTestData(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping integration tests")
	}

	// Test with a realistic implementation configuration
	cfg := config.ImplementationConfig{
		Name:    "integration-test-impl",
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
			config.FeatureMultiline,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorBooleanLenient,
		},
		VariantChoice: config.VariantProposed,
	}

	// Test loading compatible tests
	tests, err := LoadCompatibleTests(testDataPath, cfg)
	if err != nil {
		t.Fatalf("Failed to load compatible tests from real data: %v", err)
	}

	t.Logf("Loaded %d compatible tests from real CCL test data", len(tests))

	// Verify test structure
	if len(tests) > 0 {
		// Check first test has expected flat format structure
		test := tests[0]
		if test.Validation == "" {
			t.Error("Real test data should have validation field (flat format)")
		}
		if test.Expected == nil {
			t.Error("Real test data should have expected field")
		}
		if test.Input == "" {
			t.Error("Real test data should have input field")
		}

		t.Logf("Sample test: %s, validation: %s", test.Name, test.Validation)
	}

	// Test statistics generation
	stats, err := GetTestStats(testDataPath, cfg)
	if err != nil {
		t.Fatalf("Failed to get test statistics from real data: %v", err)
	}

	t.Logf("Statistics: %d total tests, %d compatible, %d total assertions",
		stats.TotalTests, stats.CompatibleTests, stats.TotalAssertions)

	// Verify statistics make sense
	if stats.TotalTests < 0 {
		t.Error("Total tests should not be negative")
	}
	if stats.CompatibleTests > stats.TotalTests {
		t.Error("Compatible tests should not exceed total tests")
	}
	if len(stats.ByFunction) == 0 && stats.TotalTests > 0 {
		t.Error("Should have function breakdown if tests exist")
	}
}

func TestIntegration_RealDataGeneration(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping generation tests")
	}

	// Check if source tests directory exists
	sourceDir := filepath.Join(testDataPath, "tests")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Skip("Source tests directory not found - skipping generation tests")
	}

	// Create temporary output directory
	outputDir := t.TempDir()

	// Test flat format generation
	err := GenerateFlat(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("Failed to generate flat format from real source data: %v", err)
	}

	// Verify output files were created
	files, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to find generated files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected output files to be generated from real source data")
	}

	t.Logf("Generated %d files from real CCL test data", len(files))

	// Verify at least one file has valid content
	if len(files) > 0 {
		data, err := os.ReadFile(files[0])
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		// Should be valid JSON
		var content interface{}
		if err := json.Unmarshal(data, &content); err != nil {
			t.Errorf("Generated file contains invalid JSON: %v", err)
		}

		t.Logf("Successfully validated generated file: %s", filepath.Base(files[0]))
	}
}

func TestIntegration_ProgressiveImplementation(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping progressive implementation tests")
	}

	// Test different implementation levels
	levels := []struct {
		name      string
		functions []config.CCLFunction
		features  []config.CCLFeature
	}{
		{
			name: "minimal",
			functions: []config.CCLFunction{
				config.FunctionParse,
			},
			features: []config.CCLFeature{},
		},
		{
			name: "basic",
			functions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
			},
			features: []config.CCLFeature{
				config.FeatureComments,
			},
		},
		{
			name: "advanced",
			functions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
				config.FunctionGetString,
				config.FunctionGetInt,
				config.FunctionGetBool,
			},
			features: []config.CCLFeature{
				config.FeatureComments,
				config.FeatureMultiline,
			},
		},
	}

	var previousCount int
	for _, level := range levels {
		cfg := config.ImplementationConfig{
			Name:               level.name,
			Version:            "v1.0.0",
			SupportedFunctions: level.functions,
			SupportedFeatures:  level.features,
			BehaviorChoices: []config.CCLBehavior{
				config.BehaviorCRLFNormalize,
				config.BehaviorBooleanLenient,
			},
			VariantChoice: config.VariantProposed,
		}

		tests, err := LoadCompatibleTests(testDataPath, cfg)
		if err != nil {
			t.Fatalf("Failed to load tests for %s implementation: %v", level.name, err)
		}

		t.Logf("%s implementation: %d compatible tests", level.name, len(tests))

		// More advanced implementations should generally have more or equal compatible tests
		if len(tests) < previousCount {
			t.Logf("Note: %s implementation has fewer tests than previous level (%d vs %d)",
				level.name, len(tests), previousCount)
		}

		previousCount = len(tests)

		// Verify test compatibility
		loader := NewLoader(testDataPath, cfg)
		for i, test := range tests {
			if i >= 10 { // Only check first 10 tests for performance
				break
			}
			if !loader.IsTestCompatible(test) {
				t.Errorf("Test %s should be compatible with %s implementation", test.Name, level.name)
			}
		}
	}
}

func TestIntegration_SpecificFunctionFiltering(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping function filtering tests")
	}

	cfg := config.ImplementationConfig{
		Name:    "function-filter-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
		},
		VariantChoice: config.VariantProposed,
	}

	loader := NewLoader(testDataPath, cfg)

	// Test loading specific functions
	functions := []config.CCLFunction{
		config.FunctionParse,
		config.FunctionBuildHierarchy,
		config.FunctionGetString,
	}

	for _, fn := range functions {
		tests, err := loader.LoadTestsByFunction(fn, LoadOptions{
			Format:     FormatFlat,
			FilterMode: FilterCompatible,
		})
		if err != nil {
			t.Fatalf("Failed to load tests for function %s: %v", fn, err)
		}

		t.Logf("Function %s: %d compatible tests", fn, len(tests))

		// Verify all loaded tests are for the requested function
		for _, test := range tests {
			if test.Validation != string(fn) {
				// Check if function is in the metadata
				found := false
				for _, testFn := range test.Functions {
					if testFn == string(fn) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Test %s should be for function %s, got validation %s", test.Name, fn, test.Validation)
				}
			}
		}
	}
}

func TestIntegration_CapabilityCoverage(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping capability coverage tests")
	}

	cfg := config.ImplementationConfig{
		Name:    "coverage-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
			config.FunctionGetInt,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
			config.FeatureMultiline,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorBooleanLenient,
		},
		VariantChoice: config.VariantProposed,
	}

	loader := NewLoader(testDataPath, cfg)
	coverage := loader.GetCapabilityCoverage()

	t.Logf("Capability coverage analysis:")

	// Check function coverage
	for fn, info := range coverage.Functions {
		t.Logf("  Function %s: %d available, %d compatible", fn, info.Available, info.Compatible)
		if info.Available < 0 {
			t.Errorf("Available count should not be negative for function %s", fn)
		}
		if info.Compatible < 0 {
			t.Errorf("Compatible count should not be negative for function %s", fn)
		}
		if info.Compatible > info.Available {
			t.Errorf("Compatible count should not exceed available for function %s", fn)
		}
	}

	// Check feature coverage
	for feature, info := range coverage.Features {
		t.Logf("  Feature %s: %d available, %d compatible", feature, info.Available, info.Compatible)
		if info.Available < 0 {
			t.Errorf("Available count should not be negative for feature %s", feature)
		}
		if info.Compatible < 0 {
			t.Errorf("Compatible count should not be negative for feature %s", feature)
		}
	}

	// Verify we have coverage information for our supported capabilities
	for _, fn := range cfg.SupportedFunctions {
		if _, exists := coverage.Functions[fn]; !exists {
			t.Errorf("Should have coverage information for supported function %s", fn)
		}
	}

	for _, feature := range cfg.SupportedFeatures {
		if _, exists := coverage.Features[feature]; !exists {
			t.Errorf("Should have coverage information for supported feature %s", feature)
		}
	}
}

func TestIntegration_LevelBasedFiltering(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping level-based filtering tests")
	}


	// Test progressive implementation - different capability levels
	capabilities := []struct {
		name      string
		functions []config.CCLFunction
	}{
		{"minimal", []config.CCLFunction{config.FunctionParse}},
		{"basic", []config.CCLFunction{config.FunctionParse, config.FunctionBuildHierarchy}},
		{"intermediate", []config.CCLFunction{config.FunctionParse, config.FunctionBuildHierarchy, config.FunctionGetString}},
		{"advanced", []config.CCLFunction{config.FunctionParse, config.FunctionBuildHierarchy, config.FunctionGetString, config.FunctionGetInt, config.FunctionGetBool}},
	}

	var previousCount int
	for _, cap := range capabilities {
		// Create a new config with this capability level
		testCfg := config.ImplementationConfig{
			Name:               "progressive-test",
			Version:            "v1.0.0",
			SupportedFunctions: cap.functions,
			SupportedFeatures:  []config.CCLFeature{config.FeatureComments},
			BehaviorChoices:    []config.CCLBehavior{config.BehaviorCRLFNormalize},
			VariantChoice:      config.VariantProposed,
		}

		testLoader := NewLoader(testDataPath, testCfg)
		tests, err := testLoader.LoadAllTests(LoadOptions{
			Format:     FormatFlat,
			FilterMode: FilterCompatible,
		})
		if err != nil {
			t.Fatalf("Failed to load tests for %s capability: %v", cap.name, err)
		}

		t.Logf("%s capability: %d tests", cap.name, len(tests))

		// Higher capability levels should include more or equal tests (cumulative)
		if len(tests) < previousCount {
			t.Errorf("%s capability should have at least as many tests as previous capability (%d vs %d)",
				cap.name, len(tests), previousCount)
			}


		previousCount = len(tests)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping error handling tests")
	}

	cfg := config.ImplementationConfig{
		Name:    "error-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
		},
		SupportedFeatures: []config.CCLFeature{},
		BehaviorChoices:   []config.CCLBehavior{},
		VariantChoice:     config.VariantProposed,
	}

	// Test with non-existent subdirectory
	invalidPath := filepath.Join(testDataPath, "nonexistent")
	tests, err := LoadCompatibleTests(invalidPath, cfg)
	if err != nil {
		t.Fatalf("LoadCompatibleTests should handle missing subdirectory gracefully: %v", err)
	}
	if len(tests) != 0 {
		t.Errorf("Expected 0 tests from non-existent directory, got %d", len(tests))
	}

	// Test statistics with invalid path
	stats, err := GetTestStats(invalidPath, cfg)
	if err != nil {
		t.Fatalf("GetTestStats should handle missing directory gracefully: %v", err)
	}
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests from non-existent directory, got %d", stats.TotalTests)
	}
}

func TestIntegration_LargeConfigurationSpaces(t *testing.T) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		t.Skip("ccl-test-data directory not found - skipping large configuration tests")
	}

	// Test with comprehensive configuration
	comprehensiveConfig := config.ImplementationConfig{
		Name:               "comprehensive-test",
		Version:            "v1.0.0",
		SupportedFunctions: config.AllFunctions(),
		SupportedFeatures:  config.AllFeatures(),
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorTabsPreserve,
			config.BehaviorStrictSpacing,
			config.BehaviorBooleanStrict,
			config.BehaviorListCoercionOn,
		},
		VariantChoice: config.VariantProposed,
	}

	// Verify configuration is valid
	if err := comprehensiveConfig.IsValid(); err != nil {
		t.Fatalf("Comprehensive configuration should be valid: %v", err)
	}

	// Test loading with comprehensive config
	tests, err := LoadCompatibleTests(testDataPath, comprehensiveConfig)
	if err != nil {
		t.Fatalf("Failed to load tests with comprehensive config: %v", err)
	}

	t.Logf("Comprehensive config loaded %d compatible tests", len(tests))

	// Test statistics with comprehensive config
	stats, err := GetTestStats(testDataPath, comprehensiveConfig)
	if err != nil {
		t.Fatalf("Failed to get stats with comprehensive config: %v", err)
	}

	t.Logf("Comprehensive config stats: %d total, %d compatible",
		stats.TotalTests, stats.CompatibleTests)

	// Should have high compatibility with comprehensive config
	if stats.TotalTests > 0 {
		compatibilityRatio := float64(stats.CompatibleTests) / float64(stats.TotalTests)
		t.Logf("Compatibility ratio: %.2f", compatibilityRatio)

		if compatibilityRatio < 0.5 { // Expect at least 50% compatibility
			t.Logf("Warning: Low compatibility ratio (%.2f) with comprehensive config", compatibilityRatio)
		}
	}
}

// Benchmark real data loading for performance validation
func BenchmarkIntegration_LoadCompatibleTests(b *testing.B) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		b.Skip("ccl-test-data directory not found - skipping benchmark")
	}

	cfg := config.ImplementationConfig{
		Name:    "benchmark-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
		},
		VariantChoice: config.VariantProposed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadCompatibleTests(testDataPath, cfg)
		if err != nil {
			b.Fatalf("LoadCompatibleTests failed: %v", err)
		}
	}
}

func BenchmarkIntegration_GetTestStats(b *testing.B) {
	testDataPath := getTestDataPath()
	if testDataPath == "" {
		b.Skip("ccl-test-data directory not found - skipping benchmark")
	}

	cfg := config.ImplementationConfig{
		Name:    "benchmark-stats-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
		},
		VariantChoice: config.VariantProposed,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetTestStats(testDataPath, cfg)
		if err != nil {
			b.Fatalf("GetTestStats failed: %v", err)
		}
	}
}
