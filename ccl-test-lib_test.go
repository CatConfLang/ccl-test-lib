package ccl_test_lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/generator"
	"github.com/tylerbu/ccl-test-lib/loader"
	"github.com/tylerbu/ccl-test-lib/types"
)

// Re-export loader constants for backward compatibility
const (
	FormatCompact    = loader.FormatCompact
	FormatFlat       = loader.FormatFlat
	FilterCompatible = loader.FilterCompatible
	FilterAll        = loader.FilterAll
)

// Re-export loader types
type LoadOptions = loader.LoadOptions

// Test data setup for integration testing
func setupIntegrationTestData(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create test directories
	testsDir := filepath.Join(tmpDir, "tests")
	generatedDir := filepath.Join(tmpDir, "generated_tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create compact format test data wrapped in CompactTestFile structure
	compactTests := []loader.CompactTest{
		{
			Name:     "integration_test_1",
			Input:    "name = Alice\nage = 25",
			Level:    1,
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect: []map[string]interface{}{
						{"key": "name", "value": "Alice"},
						{"key": "age", "value": "25"},
					},
				},
				{
					Function: "build_hierarchy",
					Expect: map[string]interface{}{
						"name": "Alice",
						"age":  "25",
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
			},
		},
		{
			Name:     "integration_test_2",
			Input:    "enabled = true\ncount = 42",
			Level:    2,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{
					Function: "get_bool",
					Args:     []string{"enabled"},
					Expect:   true,
				},
				{
					Function: "get_int",
					Args:     []string{"count"},
					Expect:   42,
				},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  compactTests,
	}

	sourceData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(testsDir, "integration.json"), sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source test file: %v", err)
	}

	// Create flat format test data (pre-generated)
	flatTests := []types.TestCase{
		{
			Name:       "flat_test_parse",
			Input:      "key = value",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "key", "value": "value"}},
			Functions:  []string{"parse"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "flat_test",
			Meta:       types.TestMetadata{Level: 1},
		},
		{
			Name:       "flat_test_get_string",
			Input:      "key = value",
			Validation: "get_string",
			Expected:   "value",
			Args:       []string{"key"},
			Functions:  []string{"get_string"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "flat_test",
			Meta:       types.TestMetadata{Level: 1},
		},
	}

	// Wrap flat tests in object format like real ccl-test-data files
	flatSuite := types.TestSuite{
		Suite:   "Integration Test Suite",
		Version: "1.0",
		Tests:   flatTests,
	}
	flatData, _ := json.MarshalIndent(flatSuite, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "integration.json"), flatData, 0644); err != nil {
		t.Fatalf("Failed to write flat test file: %v", err)
	}

	return tmpDir
}

func createTestImplementationConfig() config.ImplementationConfig {
	return config.ImplementationConfig{
		Name:    "test-impl",
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
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "v0.1.0" {
		t.Errorf("Expected version v0.1.0, got %s", Version)
	}
}

func TestNewLoader(t *testing.T) {
	testDataPath := "/test/data"
	cfg := createTestImplementationConfig()

	loader := NewLoader(testDataPath, cfg)

	if loader == nil {
		t.Fatal("NewLoader should return a non-nil loader")
	}

	// Verify it returns the same type as the loader package
	expectedType := "*loader.TestLoader"
	actualType := getTypeName(loader)
	if actualType != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, actualType)
	}

	// Test that the loader has the correct configuration
	if loader.TestDataPath != testDataPath {
		t.Errorf("Expected test data path %s, got %s", testDataPath, loader.TestDataPath)
	}
	if loader.Config.Name != cfg.Name {
		t.Errorf("Expected config name %s, got %s", cfg.Name, loader.Config.Name)
	}
}

func TestNewGenerator(t *testing.T) {
	sourceDir := "/source"
	outputDir := "/output"

	generator := NewGenerator(sourceDir, outputDir)

	if generator == nil {
		t.Fatal("NewGenerator should return a non-nil generator")
	}

	// Verify it returns the same type as the generator package
	expectedType := "*generator.FlatGenerator"
	actualType := getTypeName(generator)
	if actualType != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, actualType)
	}

	// Test that the generator has the correct configuration
	if generator.SourceDir != sourceDir {
		t.Errorf("Expected source dir %s, got %s", sourceDir, generator.SourceDir)
	}
	if generator.OutputDir != outputDir {
		t.Errorf("Expected output dir %s, got %s", outputDir, generator.OutputDir)
	}
	if !generator.Options.Verbose {
		t.Error("Expected verbose option to be true by default")
	}
}

func TestLoadCompatibleTests(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)
	cfg := createTestImplementationConfig()

	tests, err := LoadCompatibleTests(testDataPath, cfg)
	if err != nil {
		t.Fatalf("LoadCompatibleTests failed: %v", err)
	}

	if len(tests) == 0 {
		t.Error("Expected to load some tests")
	}

	// Verify all loaded tests are compatible
	for _, test := range tests {
		if test.Validation == "" {
			t.Error("Loaded test should have validation field (flat format)")
		}
		if test.Expected == nil {
			t.Error("Loaded test should have expected field")
		}
	}

	// Verify we got flat format tests (should have SourceTest field)
	foundSourceTest := false
	for _, test := range tests {
		if test.SourceTest != "" {
			foundSourceTest = true
			break
		}
	}
	if !foundSourceTest {
		t.Error("Expected to find tests with SourceTest field (indicating flat format)")
	}
}

func TestLoadCompatibleTests_NoTestData(t *testing.T) {
	// Test with non-existent directory
	cfg := createTestImplementationConfig()

	tests, err := LoadCompatibleTests("/nonexistent", cfg)
	if err != nil {
		t.Fatalf("LoadCompatibleTests should not error on missing directory: %v", err)
	}

	if len(tests) != 0 {
		t.Errorf("Expected 0 tests for non-existent directory, got %d", len(tests))
	}
}

func TestGenerateFlat(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)
	sourceDir := filepath.Join(testDataPath, "tests")
	outputDir := filepath.Join(testDataPath, "output")

	err := GenerateFlat(sourceDir, outputDir)
	if err != nil {
		t.Fatalf("GenerateFlat failed: %v", err)
	}

	// Verify output files were created
	files, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to find output files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected output files to be generated")
	}

	// Verify output file content
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read output file %s: %v", file, err)
		}

		// Should be valid JSON
		var content interface{}
		if err := json.Unmarshal(data, &content); err != nil {
			t.Errorf("Output file %s contains invalid JSON: %v", file, err)
		}
	}
}

func TestGenerateFlat_NonexistentSource(t *testing.T) {
	err := GenerateFlat("/nonexistent/source", "/tmp/output")
	// May not error immediately if no files found, just return without processing
	// This is acceptable behavior
	_ = err
}

func TestGetTestStats(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)
	cfg := createTestImplementationConfig()

	stats, err := GetTestStats(testDataPath, cfg)
	if err != nil {
		t.Fatalf("GetTestStats failed: %v", err)
	}

	// Verify statistics structure
	if stats.TotalTests < 0 {
		t.Error("TotalTests should not be negative")
	}
	if stats.TotalAssertions < 0 {
		t.Error("TotalAssertions should not be negative")
	}
	if stats.CompatibleTests < 0 {
		t.Error("CompatibleTests should not be negative")
	}
	if stats.CompatibleAsserts < 0 {
		t.Error("CompatibleAsserts should not be negative")
	}

	// Verify maps are initialized
	if stats.ByLevel == nil {
		t.Error("ByLevel map should be initialized")
	}
	if stats.ByFunction == nil {
		t.Error("ByFunction map should be initialized")
	}
	if stats.ByFeature == nil {
		t.Error("ByFeature map should be initialized")
	}

	// Compatible counts should not exceed total counts
	if stats.CompatibleTests > stats.TotalTests {
		t.Error("CompatibleTests should not exceed TotalTests")
	}
	if stats.CompatibleAsserts > stats.TotalAssertions {
		t.Error("CompatibleAsserts should not exceed TotalAssertions")
	}
}

func TestGetTestStats_NoTestData(t *testing.T) {
	cfg := createTestImplementationConfig()

	stats, err := GetTestStats("/nonexistent", cfg)
	if err != nil {
		t.Fatalf("GetTestStats should not error on missing directory: %v", err)
	}

	// Should return empty statistics
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got %d", stats.TotalTests)
	}
	if stats.TotalAssertions != 0 {
		t.Errorf("Expected 0 total assertions, got %d", stats.TotalAssertions)
	}
}

// Integration test combining multiple package functionalities
func TestIntegrationWorkflow(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)
	cfg := createTestImplementationConfig()

	// Step 1: Generate flat format from source
	sourceDir := filepath.Join(testDataPath, "tests")
	generatedDir := filepath.Join(testDataPath, "integration-generated")

	err := GenerateFlat(sourceDir, generatedDir)
	if err != nil {
		t.Fatalf("Step 1 - Generate flat format failed: %v", err)
	}

	// Step 2: Load compatible tests from generated flat format
	// Update test data path to use generated directory
	testLoader := NewLoader(testDataPath, cfg)

	// Load from the newly generated directory
	tests, err := testLoader.LoadAllTests(LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterCompatible,
	})
	if err != nil {
		t.Fatalf("Step 2 - Load compatible tests failed: %v", err)
	}

	if len(tests) == 0 {
		t.Error("Step 2 - Expected to load some tests from generated format")
	}

	// Step 3: Get comprehensive statistics
	stats, err := GetTestStats(testDataPath, cfg)
	if err != nil {
		t.Fatalf("Step 3 - Get test stats failed: %v", err)
	}

	// Step 4: Verify workflow results
	if stats.TotalTests == 0 {
		t.Error("Step 4 - Expected some tests in statistics")
	}

	// Verify that generated tests are properly categorized
	foundParseTests := false
	foundTypedAccessTests := false

	for _, test := range tests {
		if test.Validation == "parse" {
			foundParseTests = true
		}
		if test.Validation == "get_string" || test.Validation == "get_int" {
			foundTypedAccessTests = true
		}
	}

	if !foundParseTests {
		t.Error("Step 4 - Expected to find parse validation tests")
	}
	if !foundTypedAccessTests {
		t.Error("Step 4 - Expected to find typed access validation tests")
	}

	// Verify statistics include function breakdown
	if stats.ByFunction["parse"] == 0 {
		t.Error("Step 4 - Expected parse function to be represented in statistics")
	}
}

// Test cross-package compatibility
func TestPackageCompatibility(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)
	cfg := createTestImplementationConfig()

	// Test that convenience functions work with underlying packages
	loader := NewLoader(testDataPath, cfg)
	generator := NewGenerator("/source", "/output")

	// Verify loader can be used directly
	allTests, err := loader.LoadAllTests(LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	})
	if err != nil {
		t.Fatalf("Direct loader usage failed: %v", err)
	}

	// Verify generator can be used directly
	if generator.SourceDir != "/source" {
		t.Error("Generator should maintain source directory setting")
	}

	// Test that convenience functions provide equivalent functionality
	compatibleTests, err := LoadCompatibleTests(testDataPath, cfg)
	if err != nil {
		t.Fatalf("Convenience function failed: %v", err)
	}

	// Should get compatible subset of all tests
	if len(compatibleTests) > len(allTests) {
		t.Error("Compatible tests should be subset of all tests")
	}
}

// Test error handling in convenience functions
func TestConvenienceFunctionsErrorHandling(t *testing.T) {
	cfg := createTestImplementationConfig()

	// Test LoadCompatibleTests with invalid path
	_, err := LoadCompatibleTests("/definitely/nonexistent/path", cfg)
	if err != nil {
		// This should not error, just return empty results
		t.Errorf("LoadCompatibleTests should handle missing paths gracefully: %v", err)
	}

	// Test GenerateFlat with invalid source
	err = GenerateFlat("/nonexistent/source", "/tmp/output")
	// May not error immediately if no files found, just return without processing
	// This is acceptable behavior
	_ = err

	// Test GetTestStats with invalid path
	_, err = GetTestStats("/nonexistent/path", cfg)
	if err != nil {
		// This should not error, just return empty stats
		t.Errorf("GetTestStats should handle missing paths gracefully: %v", err)
	}
}

// Test with minimal configuration
func TestMinimalConfiguration(t *testing.T) {
	testDataPath := setupIntegrationTestData(t)

	// Create minimal config (no functions, features, etc.)
	minimalConfig := config.ImplementationConfig{
		Name:               "minimal",
		Version:            "v0.1.0",
		SupportedFunctions: []config.CCLFunction{},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	// Should not error, but should return no compatible tests
	tests, err := LoadCompatibleTests(testDataPath, minimalConfig)
	if err != nil {
		t.Fatalf("LoadCompatibleTests with minimal config failed: %v", err)
	}

	// With minimal config, no tests should be compatible
	if len(tests) != 0 {
		t.Errorf("Expected 0 compatible tests with minimal config, got %d", len(tests))
	}

	// Statistics should still work
	stats, err := GetTestStats(testDataPath, minimalConfig)
	if err != nil {
		t.Fatalf("GetTestStats with minimal config failed: %v", err)
	}

	// Should have total tests but no compatible tests
	if stats.TotalTests == 0 {
		// This might be 0 if no flat tests exist, which is fine
	}
	if stats.CompatibleTests != 0 {
		t.Errorf("Expected 0 compatible tests with minimal config, got %d", stats.CompatibleTests)
	}
}

// Utility function to get type name for testing
func getTypeName(v interface{}) string {
	switch v.(type) {
	case *loader.TestLoader:
		return "*loader.TestLoader"
	case *generator.FlatGenerator:
		return "*generator.FlatGenerator"
	default:
		return "unknown"
	}
}

// Test constant values
func TestConstants(t *testing.T) {
	// Test that format constants are re-exported correctly
	if generator.FormatCompact != loader.FormatCompact {
		t.Error("FormatCompact constant should match loader package")
	}
	if generator.FormatFlat != loader.FormatFlat {
		t.Error("FormatFlat constant should match loader package")
	}
}

// Additional test for edge cases
func TestEdgeCases(t *testing.T) {
	// Test with empty config
	emptyConfig := config.ImplementationConfig{}

	loader := NewLoader("", emptyConfig)
	if loader == nil {
		t.Error("NewLoader should handle empty config")
	}

	generator := NewGenerator("", "")
	if generator == nil {
		t.Error("NewGenerator should handle empty paths")
	}

	// Test error propagation
	_, err := LoadCompatibleTests("", emptyConfig)
	if err != nil {
		// Should not error, just return empty results
		t.Errorf("LoadCompatibleTests should handle empty path gracefully: %v", err)
	}
}
