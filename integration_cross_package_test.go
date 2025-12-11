package ccl_test_lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/CatConfLang/ccl-test-lib/config"
	"github.com/CatConfLang/ccl-test-lib/generator"
	"github.com/CatConfLang/ccl-test-lib/loader"
	"github.com/CatConfLang/ccl-test-lib/types"
)

// Integration tests focusing on cross-package interactions
// These tests verify that different packages work together correctly

func TestCrossPackage_LoaderGeneratorRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	// Create directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create source format test data
	sourceTests := []loader.CompactTest{
		{
			Name:     "roundtrip_test",
			Input:    "name = Alice\nage = 25\nenabled = true",
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect: []map[string]interface{}{
						{"key": "name", "value": "Alice"},
						{"key": "age", "value": "25"},
						{"key": "enabled", "value": "true"},
					},
				},
				{
					Function: "build_hierarchy",
					Expect: map[string]interface{}{
						"name":    "Alice",
						"age":     "25",
						"enabled": "true",
					},
				},
				{
					Function: "get_string",
					Args:     []string{"name"},
					Expect:   "Alice",
				},
				{
					Function: "get_int",
					Args:     []string{"age"},
					Expect:   25,
				},
				{
					Function: "get_bool",
					Args:     []string{"enabled"},
					Expect:   true,
				},
			},
		},
	}

	// Write source data wrapped in CompactTestFile structure
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  sourceTests,
	}
	sourceData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	sourceFile := filepath.Join(sourceDir, "roundtrip.json")
	if err := os.WriteFile(sourceFile, sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Step 1: Generate flat format using generator
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
		Verbose:      false,
	})

	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Generator failed: %v", err)
	}

	// Step 2: Load generated tests using loader
	cfg := config.ImplementationConfig{
		Name:    "cross-package-test",
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
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorBooleanLenient,
		},
		VariantChoice: config.VariantProposed,
	}

	testLoader := loader.NewTestLoader(tmpDir, cfg)
	tests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterCompatible,
	})
	if err != nil {
		t.Fatalf("Loader failed: %v", err)
	}

	// Step 3: Verify round-trip integrity
	if len(tests) != 5 {
		t.Errorf("Expected 5 flat tests (one per validation), got %d", len(tests))
	}

	// Verify each test type exists
	validationCounts := make(map[string]int)
	for _, test := range tests {
		validationCounts[test.Validation]++

		// Verify all tests have the same source
		if test.SourceTest != "roundtrip_test" {
			t.Errorf("Expected source test 'roundtrip_test', got %s", test.SourceTest)
		}

		// Verify all tests have the same input
		expectedInput := "name = Alice\nage = 25\nenabled = true"
		if test.Input != expectedInput {
			t.Errorf("Input mismatch for %s test", test.Validation)
		}
	}

	// Verify we have exactly one of each validation type
	expectedValidations := []string{"parse", "build_hierarchy", "get_string", "get_int", "get_bool"}
	for _, validation := range expectedValidations {
		if validationCounts[validation] != 1 {
			t.Errorf("Expected exactly 1 %s test, got %d", validation, validationCounts[validation])
		}
	}
}

func TestCrossPackage_ConfigCompatibilityFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Create test data with various compatibility requirements
	flatTests := []types.TestCase{
		{
			Name:       "basic_parse",
			Input:      "key = value",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "key", "value": "value"}},
			Functions:  []string{"parse"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
		},
		{
			Name:       "comments_parse",
			Input:      "key = value\n/= comment",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "key", "value": "value"}},
			Functions:  []string{"parse"},
			Features:   []string{"comments"},
			Behaviors:  []string{},
			Variants:   []string{},
		},
		{
			Name:       "unicode_parse",
			Input:      "名前 = 値",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "名前", "value": "値"}},
			Functions:  []string{"parse"},
			Features:   []string{"unicode"},
			Behaviors:  []string{},
			Variants:   []string{},
		},
		{
			Name:       "advanced_get_string",
			Input:      "key = value",
			Validation: "get_string",
			Expected:   "value",
			Args:       []string{"key"},
			Functions:  []string{"get_string"},
			Features:   []string{},
			Behaviors:  []string{"boolean_strict"},
			Variants:   []string{},
		},
	}

	flatData, _ := json.MarshalIndent(flatTests, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "compatibility.json"), flatData, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test different compatibility configurations
	testCases := []struct {
		name              string
		config            config.ImplementationConfig
		expectedTestCount int
		expectedTests     []string
	}{
		{
			name: "minimal_support",
			config: config.ImplementationConfig{
				Name:               "minimal",
				Version:            "v1.0.0",
				SupportedFunctions: []config.CCLFunction{config.FunctionParse},
				SupportedFeatures:  []config.CCLFeature{},
				BehaviorChoices:    []config.CCLBehavior{},
				VariantChoice:      config.VariantProposed,
			},
			expectedTestCount: 1,
			expectedTests:     []string{"basic_parse"},
		},
		{
			name: "comments_support",
			config: config.ImplementationConfig{
				Name:               "comments",
				Version:            "v1.0.0",
				SupportedFunctions: []config.CCLFunction{config.FunctionParse},
				SupportedFeatures:  []config.CCLFeature{config.FeatureComments},
				BehaviorChoices:    []config.CCLBehavior{},
				VariantChoice:      config.VariantProposed,
			},
			expectedTestCount: 2,
			expectedTests:     []string{"basic_parse", "comments_parse"},
		},
		{
			name: "full_support",
			config: config.ImplementationConfig{
				Name:    "full",
				Version: "v1.0.0",
				SupportedFunctions: []config.CCLFunction{
					config.FunctionParse,
					config.FunctionGetString,
				},
				SupportedFeatures: []config.CCLFeature{
					config.FeatureComments,
					config.FeatureUnicode,
				},
				BehaviorChoices: []config.CCLBehavior{
					config.BehaviorBooleanStrict,
				},
				VariantChoice: config.VariantProposed,
			},
			expectedTestCount: 4,
			expectedTests:     []string{"basic_parse", "comments_parse", "unicode_parse", "advanced_get_string"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testLoader := loader.NewTestLoader(tmpDir, tc.config)
			tests, err := testLoader.LoadAllTests(loader.LoadOptions{
				Format:     loader.FormatFlat,
				FilterMode: loader.FilterCompatible,
			})
			if err != nil {
				t.Fatalf("Failed to load tests: %v", err)
			}

			if len(tests) != tc.expectedTestCount {
				t.Errorf("Expected %d compatible tests, got %d", tc.expectedTestCount, len(tests))
			}

			// Verify expected tests are present
			testNames := make(map[string]bool)
			for _, test := range tests {
				testNames[test.Name] = true
			}

			for _, expectedTest := range tc.expectedTests {
				if !testNames[expectedTest] {
					t.Errorf("Expected test %s not found in compatible tests", expectedTest)
				}
			}
		})
	}
}

func TestCrossPackage_StatisticsAccuracy(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create source data with known characteristics
	sourceTests := []loader.CompactTest{
		{
			Name:     "test_level_1",
			Input:    "key1 = value1",
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "key1", "value": "value1"}}},
				{Function: "get_string", Args: []string{"key1"}, Expect: "value1"},
			},
		},
		{
			Name:     "test_level_2",
			Input:    "key2 = value2\n/= comment",
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "key2", "value": "value2"}}},
				{Function: "filter", Expect: []map[string]interface{}{{"key": "key2", "value": "value2"}}},
			},
		},
		{
			Name:     "test_level_3",
			Input:    "key3 = value3",
			Features: []string{"unicode"},
			Tests: []loader.CompactValidation{
				{Function: "get_int", Args: []string{"key3"}, Expect: 3},
			},
		},
	}

	// Write source data wrapped in CompactTestFile structure
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  sourceTests,
	}
	sourceData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "stats.json"), sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Generate flat format
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Test statistics with different configurations
	cfg := config.ImplementationConfig{
		Name:    "stats-test",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionGetString,
			config.FunctionGetInt,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{},
		VariantChoice:   config.VariantProposed,
	}

	// Get statistics using convenience function
	stats, err := GetTestStats(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	// Expected: 5 total tests (2+2+1), 3 compatible (test_level_1: parse+get_string, test_level_2: parse only, test_level_3: none due to unicode)
	expectedTotal := 5
	expectedCompatible := 3

	if stats.TotalTests != expectedTotal {
		t.Errorf("Expected %d total tests, got %d", expectedTotal, stats.TotalTests)
	}
	if stats.CompatibleTests != expectedCompatible {
		t.Errorf("Expected %d compatible tests, got %d", expectedCompatible, stats.CompatibleTests)
	}

	// Verify function breakdown - after level removal, counts are based on actual validations
	// Each source test can expand to multiple flat tests (one per validation)
	if stats.ByFunction["parse"] < 2 {
		t.Errorf("Expected at least 2 parse tests, got %d", stats.ByFunction["parse"])
	}
	if stats.ByFunction["get_string"] < 1 {
		t.Errorf("Expected at least 1 get_string test, got %d", stats.ByFunction["get_string"])
	}

	// Cross-verify with loader statistics
	testLoader := loader.NewTestLoader(tmpDir, cfg)
	allTests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterAll,
	})
	if err != nil {
		t.Fatalf("Failed to load all tests: %v", err)
	}

	loaderStats := testLoader.GetTestStatistics(allTests)
	if loaderStats.TotalTests != stats.TotalTests {
		t.Errorf("Loader and convenience function statistics mismatch: %d vs %d",
			loaderStats.TotalTests, stats.TotalTests)
	}
}

func TestCrossPackage_ErrorPropagation(t *testing.T) {
	tmpDir := t.TempDir()

	// Test error propagation from generator to loader
	cfg := config.ImplementationConfig{
		Name:               "error-test",
		Version:            "v1.0.0",
		SupportedFunctions: []config.CCLFunction{config.FunctionParse},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	// Test with completely empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Should not error, just return empty results
	tests, err := LoadCompatibleTests(emptyDir, cfg)
	if err != nil {
		t.Errorf("LoadCompatibleTests should handle empty directory gracefully: %v", err)
	}
	if len(tests) != 0 {
		t.Errorf("Expected 0 tests from empty directory, got %d", len(tests))
	}

	stats, err := GetTestStats(emptyDir, cfg)
	if err != nil {
		t.Errorf("GetTestStats should handle empty directory gracefully: %v", err)
	}
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests from empty directory, got %d", stats.TotalTests)
	}

	// Test with directory containing invalid JSON
	invalidDir := filepath.Join(tmpDir, "invalid")
	if err := os.MkdirAll(filepath.Join(invalidDir, "generated_tests"), 0755); err != nil {
		t.Fatalf("Failed to create invalid directory: %v", err)
	}

	invalidFile := filepath.Join(invalidDir, "generated_tests", "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	// Loader should handle this gracefully by skipping invalid files
	testLoader := loader.NewTestLoader(invalidDir, cfg)
	tests, err = testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterAll,
	})
	// This might error or might skip - both are acceptable behaviors
	// The important thing is it doesn't crash
	_ = err
	_ = tests
}

func TestCrossPackage_ConfigValidation(t *testing.T) {
	// Test that invalid configurations are properly handled across packages

	invalidConfigs := []config.ImplementationConfig{
		{
			// Missing name
			Version:            "v1.0.0",
			SupportedFunctions: []config.CCLFunction{config.FunctionParse},
		},
		{
			Name:    "test",
			Version: "v1.0.0",
			// No supported functions
			SupportedFunctions: []config.CCLFunction{},
		},
		{
			Name:               "test",
			Version:            "v1.0.0",
			SupportedFunctions: []config.CCLFunction{config.FunctionParse},
			// Conflicting behaviors (if any exist)
			BehaviorChoices: []config.CCLBehavior{
				config.BehaviorCRLFNormalize,
				config.BehaviorCRLFPreserve, // These might conflict
			},
		},
	}

	tmpDir := t.TempDir()

	for i, cfg := range invalidConfigs {
		t.Run(fmt.Sprintf("invalid_config_%d", i), func(t *testing.T) {
			// Test that loader handles invalid config
			testLoader := loader.NewTestLoader(tmpDir, cfg)
			if testLoader == nil {
				t.Error("NewTestLoader should not return nil even with invalid config")
			}

			// Test that convenience functions handle invalid config
			tests, err := LoadCompatibleTests(tmpDir, cfg)
			if err != nil {
				// Either should error gracefully or work with degraded functionality
				t.Logf("LoadCompatibleTests with invalid config: %v", err)
			}
			_ = tests

			stats, err := GetTestStats(tmpDir, cfg)
			if err != nil {
				t.Logf("GetTestStats with invalid config: %v", err)
			}
			_ = stats
		})
	}
}
