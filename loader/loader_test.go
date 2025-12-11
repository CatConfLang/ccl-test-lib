package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CatConfLang/ccl-test-lib/config"
	"github.com/CatConfLang/ccl-test-lib/types"
)

// Test data setup
func setupTestData(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create test directories
	testsDir := filepath.Join(tmpDir, "source_tests")
	generatedDir := filepath.Join(tmpDir, "generated_tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create compact format test file
	compactTests := []CompactTest{
		{
			Name:     "test_parse",
			Input:    "key = value",
			Features: []string{"comments"},
			Tests: []CompactValidation{
				{
					Function: "parse",
					Expect:   []map[string]interface{}{{"key": "key", "value": "value"}},
				},
				{
					Function: "build_hierarchy",
					Expect:   map[string]interface{}{"key": "value"},
				},
			},
		},
		{
			Name:     "test_typed_access",
			Input:    "count = 42\nflag = true",
			Features: []string{},
			Tests: []CompactValidation{
				{
					Function: "get_int",
					Args:     []string{"count"},
					Expect:   42,
				},
				{
					Function: "get_bool",
					Args:     []string{"flag"},
					Expect:   true,
				},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile := CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  compactTests,
	}
	compactData, _ := json.MarshalIndent(compactTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(testsDir, "test-basic.json"), compactData, 0644); err != nil {
		t.Fatalf("Failed to write compact test file: %v", err)
	}

	// Create flat format test file
	flatTests := []types.TestCase{
		{
			Name:       "test_parse_parse",
			Input:      "key = value",
			Validation: "parse",
			Expected:   []map[string]interface{}{{"key": "key", "value": "value"}},
			Functions:  []string{"parse"},
			Features:   []string{"comments"},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "test_parse",
		},
		{
			Name:       "test_parse_build_hierarchy",
			Input:      "key = value",
			Validation: "build_hierarchy",
			Expected:   map[string]interface{}{"key": "value"},
			Functions:  []string{"build_hierarchy"},
			Features:   []string{"comments"},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "test_parse",
		},
		{
			Name:       "test_typed_access_get_int",
			Input:      "count = 42\nflag = true",
			Validation: "get_int",
			Expected:   42,
			Args:       []string{"count"},
			Functions:  []string{"get_int"},
			Features:   []string{},
			Behaviors:  []string{},
			Variants:   []string{},
			SourceTest: "test_typed_access",
		},
	}

	flatData, _ := json.MarshalIndent(flatTests, "", "  ")
	if err := os.WriteFile(filepath.Join(generatedDir, "test-basic.json"), flatData, 0644); err != nil {
		t.Fatalf("Failed to write flat test file: %v", err)
	}

	return tmpDir
}

func createTestConfig() config.ImplementationConfig {
	return config.ImplementationConfig{
		Name:    "test-implementation",
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

func TestNewTestLoader(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("/test/path", cfg)

	if loader.TestDataPath != "/test/path" {
		t.Errorf("Expected test data path '/test/path', got %s", loader.TestDataPath)
	}
	if loader.Config.Name != cfg.Name {
		t.Errorf("Expected config name %s, got %s", cfg.Name, loader.Config.Name)
	}
	if !loader.UseFlat {
		t.Error("Expected UseFlat to be true by default")
	}
}

func TestTestLoader_LoadTestFile_FlatFormat(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	flatFile := filepath.Join(tmpDir, "generated_tests", "test-basic.json")
	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	suite, err := loader.LoadTestFile(flatFile, opts)
	if err != nil {
		t.Fatalf("Failed to load flat format file: %v", err)
	}

	if suite.Suite != "Flat Format" {
		t.Errorf("Expected suite name 'Flat Format', got %s", suite.Suite)
	}
	if len(suite.Tests) != 3 {
		t.Errorf("Expected 3 tests, got %d", len(suite.Tests))
	}

	// Verify first test
	test := suite.Tests[0]
	if test.Name != "test_parse_parse" {
		t.Errorf("Expected test name 'test_parse_parse', got %s", test.Name)
	}
	if test.Validation != "parse" {
		t.Errorf("Expected validation 'parse', got %s", test.Validation)
	}
	if test.SourceTest != "test_parse" {
		t.Errorf("Expected source test 'test_parse', got %s", test.SourceTest)
	}
}

func TestTestLoader_LoadTestFile_CompactFormat(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	compactFile := filepath.Join(tmpDir, "source_tests", "test-basic.json")
	opts := LoadOptions{
		Format:     FormatCompact,
		FilterMode: FilterAll,
	}

	suite, err := loader.LoadTestFile(compactFile, opts)
	if err != nil {
		t.Fatalf("Failed to load compact format file: %v", err)
	}

	if suite.Suite != "Compact Format" {
		t.Errorf("Expected suite name 'Compact Format', got %s", suite.Suite)
	}
	if len(suite.Tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(suite.Tests))
	}

	// Verify first test has validations
	test := suite.Tests[0]
	if test.Validations == nil {
		t.Error("Expected validations to be populated")
	}
	if test.Validations.Parse == nil {
		t.Error("Expected parse validation to be populated")
	}
	if test.Validations.BuildHierarchy == nil {
		t.Error("Expected build_hierarchy validation to be populated")
	}
}

func TestTestLoader_LoadAllTests_FlatFormat(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	tests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load all tests: %v", err)
	}

	if len(tests) != 3 {
		t.Errorf("Expected 3 tests, got %d", len(tests))
	}

	// Verify test names
	expectedNames := []string{"test_parse_parse", "test_parse_build_hierarchy", "test_typed_access_get_int"}
	for i, test := range tests {
		if test.Name != expectedNames[i] {
			t.Errorf("Expected test %d name %s, got %s", i, expectedNames[i], test.Name)
		}
	}
}

func TestTestLoader_LoadAllTests_CompactFormat(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatCompact,
		FilterMode: FilterAll,
	}

	tests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load all compact tests: %v", err)
	}

	if len(tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(tests))
	}

	// Verify test names
	expectedNames := []string{"test_parse", "test_typed_access"}
	for i, test := range tests {
		if test.Name != expectedNames[i] {
			t.Errorf("Expected test %d name %s, got %s", i, expectedNames[i], test.Name)
		}
	}
}

func TestTestLoader_LoadTestsByFunction(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	// Load parse function tests only
	tests, err := loader.LoadTestsByFunction(config.FunctionParse, opts)
	if err != nil {
		t.Fatalf("Failed to load parse function tests: %v", err)
	}

	if len(tests) != 1 {
		t.Errorf("Expected 1 parse test, got %d", len(tests))
	}

	// Verify test is for parse function
	test := tests[0]
	if test.Validation != "parse" {
		t.Errorf("Expected parse validation, got %s", test.Validation)
	}
}

func TestTestLoader_FilterCompatibleTests(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	allTests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load all tests: %v", err)
	}

	compatible := loader.FilterCompatibleTests(allTests)

	// All tests should be compatible with our test config
	if len(compatible) != len(allTests) {
		t.Errorf("Expected %d compatible tests, got %d", len(allTests), len(compatible))
	}
}

func TestTestLoader_IsTestCompatible_Function(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	// Test compatible function
	compatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
	}
	if !loader.IsTestCompatible(compatibleTest) {
		t.Error("Test with supported function should be compatible")
	}

	// Test incompatible function
	incompatibleTest := types.TestCase{
		Validation: "get_float",
		Functions:  []string{"get_float"},
	}
	if loader.IsTestCompatible(incompatibleTest) {
		t.Error("Test with unsupported function should be incompatible")
	}
}

func TestTestLoader_IsTestCompatible_Feature(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	// Test compatible feature
	compatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Features:   []string{"comments"},
	}
	if !loader.IsTestCompatible(compatibleTest) {
		t.Error("Test with supported feature should be compatible")
	}

	// Test incompatible feature
	incompatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Features:   []string{"unicode"},
	}
	if loader.IsTestCompatible(incompatibleTest) {
		t.Error("Test with unsupported feature should be incompatible")
	}
}

func TestTestLoader_IsTestCompatible_Behavior(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	// Test compatible behavior
	compatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Behaviors:  []string{"crlf_normalize_to_lf"},
	}
	if !loader.IsTestCompatible(compatibleTest) {
		t.Error("Test with supported behavior should be compatible")
	}

	// Test incompatible behavior
	incompatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Behaviors:  []string{"crlf_preserve_literal"},
	}
	if loader.IsTestCompatible(incompatibleTest) {
		t.Error("Test with unsupported behavior should be incompatible")
	}
}

func TestTestLoader_IsTestCompatible_Variant(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	// Test compatible variant
	compatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Variants:   []string{"proposed_behavior"},
	}
	if !loader.IsTestCompatible(compatibleTest) {
		t.Error("Test with supported variant should be compatible")
	}

	// Test incompatible variant
	incompatibleTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Variants:   []string{"reference_compliant"},
	}
	if loader.IsTestCompatible(incompatibleTest) {
		t.Error("Test with unsupported variant should be incompatible")
	}
}

func TestTestLoader_IsTestCompatible_Conflicts(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	// Test with conflicting behavior
	conflictingTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Conflicts: &types.ConflictSet{
			Behaviors: []string{"crlf_normalize_to_lf"}, // This conflicts with our choice
		},
	}
	if loader.IsTestCompatible(conflictingTest) {
		t.Error("Test with conflicting behavior should be incompatible")
	}

	// Test with conflicting variant
	conflictingVariantTest := types.TestCase{
		Validation: "parse",
		Functions:  []string{"parse"},
		Conflicts: &types.ConflictSet{
			Variants: []string{"proposed_behavior"}, // This conflicts with our choice
		},
	}
	if loader.IsTestCompatible(conflictingVariantTest) {
		t.Error("Test with conflicting variant should be incompatible")
	}
}

func TestTestLoader_GetTestStatistics(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	tests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load tests for statistics: %v", err)
	}

	stats := loader.GetTestStatistics(tests)

	if stats.TotalTests != 3 {
		t.Errorf("Expected 3 total tests, got %d", stats.TotalTests)
	}
	if stats.TotalAssertions != 3 {
		t.Errorf("Expected 3 total assertions, got %d", stats.TotalAssertions)
	}
	if stats.CompatibleTests != 3 {
		t.Errorf("Expected 3 compatible tests, got %d", stats.CompatibleTests)
	}

	// Check function distribution (may have multiple due to Functions metadata)
	if stats.ByFunction["parse"] < 1 {
		t.Errorf("Expected at least 1 parse test, got %d", stats.ByFunction["parse"])
	}
	if stats.ByFunction["build_hierarchy"] < 1 {
		t.Errorf("Expected at least 1 build_hierarchy test, got %d", stats.ByFunction["build_hierarchy"])
	}
	if stats.ByFunction["get_int"] < 1 {
		t.Errorf("Expected at least 1 get_int test, got %d", stats.ByFunction["get_int"])
	}
}

func TestTestLoader_GetCapabilityCoverage(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	coverage := loader.GetCapabilityCoverage()

	// Check function coverage
	if parseCoverage, exists := coverage.Functions[config.FunctionParse]; exists {
		if parseCoverage.Available != 1 {
			t.Errorf("Expected 1 available parse test, got %d", parseCoverage.Available)
		}
		if parseCoverage.Compatible != 1 {
			t.Errorf("Expected 1 compatible parse test, got %d", parseCoverage.Compatible)
		}
	} else {
		t.Error("Expected parse function coverage to be present")
	}

	// Check feature coverage
	if commentsCoverage, exists := coverage.Features[config.FeatureComments]; exists {
		if commentsCoverage.Available != 2 {
			t.Errorf("Expected 2 available comments tests, got %d", commentsCoverage.Available)
		}
	} else {
		t.Error("Expected comments feature coverage to be present")
	}
}

func TestTestLoader_FilterByTags(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	tests := []types.TestCase{
		{
			Name: "test1",
			Meta: types.TestMetadata{Tags: []string{"function:parse", "level:1"}},
		},
		{
			Name: "test2",
			Meta: types.TestMetadata{Tags: []string{"function:build_hierarchy", "level:2"}},
		},
		{
			Name: "test3",
			Meta: types.TestMetadata{Tags: []string{"function:get_string", "level:1"}},
		},
	}

	// Test include tags
	includeTags := []string{"level:1"}
	filtered := loader.FilterByTags(tests, includeTags, nil)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests with level:1 tag, got %d", len(filtered))
	}

	// Test exclude tags
	excludeTags := []string{"function:parse"}
	filtered = loader.FilterByTags(tests, nil, excludeTags)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests without parse tag, got %d", len(filtered))
	}

	// Test both include and exclude
	includeTags = []string{"level:1"}
	excludeTags = []string{"function:parse"}
	filtered = loader.FilterByTags(tests, includeTags, excludeTags)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 test with level:1 but not parse, got %d", len(filtered))
	}
	if filtered[0].Name != "test3" {
		t.Errorf("Expected test3, got %s", filtered[0].Name)
	}
}

func TestLoadOptions_FilterModes(t *testing.T) {
	tmpDir := setupTestData(t)
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	// Test FilterAll mode
	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}
	allTests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load tests with FilterAll: %v", err)
	}

	// Test FilterCompatible mode
	opts.FilterMode = FilterCompatible
	compatibleTests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load tests with FilterCompatible: %v", err)
	}

	// In our test setup, all tests should be compatible
	if len(compatibleTests) != len(allTests) {
		t.Errorf("Expected same number of tests in compatible and all modes, got %d vs %d", len(compatibleTests), len(allTests))
	}

	// Test FilterCustom mode
	opts.FilterMode = FilterCustom
	opts.CustomFilter = func(test types.TestCase) bool {
		return test.Validation == "parse"
	}
	customTests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("Failed to load tests with FilterCustom: %v", err)
	}

	if len(customTests) < 1 {
		t.Errorf("Expected at least 1 parse test with custom filter, got %d", len(customTests))
	}

	// Verify all returned tests are parse tests
	for _, test := range customTests {
		if test.Validation != "parse" {
			t.Errorf("Custom filter should only return parse tests, got %s", test.Validation)
		}
	}
}

// Test error conditions

func TestTestLoader_LoadTestFile_NonexistentFile(t *testing.T) {
	cfg := createTestConfig()
	loader := NewTestLoader("", cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	_, err := loader.LoadTestFile("/nonexistent/file.json", opts)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestTestLoader_LoadTestFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	// Create invalid JSON file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	_, err := loader.LoadTestFile(invalidFile, opts)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestTestLoader_LoadAllTests_NoTestFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := createTestConfig()
	loader := NewTestLoader(tmpDir, cfg)

	opts := LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	}

	tests, err := loader.LoadAllTests(opts)
	if err != nil {
		t.Fatalf("LoadAllTests should not error when no files found: %v", err)
	}

	if len(tests) != 0 {
		t.Errorf("Expected 0 tests when no files exist, got %d", len(tests))
	}
}

// Test compact format parsing edge cases

func TestCompactTest_JSONMarshaling(t *testing.T) {
	compact := CompactTest{
		Name:     "test_compact",
		Input:    "key = value",
		Features: []string{"comments"},
		Tests: []CompactValidation{
			{
				Function: "parse",
				Expect:   []map[string]interface{}{{"key": "key", "value": "value"}},
			},
		},
	}

	data, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("Failed to marshal CompactTest: %v", err)
	}

	var unmarshaled CompactTest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CompactTest: %v", err)
	}

	if unmarshaled.Name != compact.Name {
		t.Errorf("Expected name %s, got %s", compact.Name, unmarshaled.Name)
	}
	if len(unmarshaled.Tests) != 1 {
		t.Errorf("Expected 1 validation, got %d", len(unmarshaled.Tests))
	}
}

func TestCompactValidation_AllFields(t *testing.T) {
	validation := CompactValidation{
		Function: "get_string",
		Expect:   "expected_value",
		Args:     []string{"key"},
		Error:    false,
	}

	data, err := json.Marshal(validation)
	if err != nil {
		t.Fatalf("Failed to marshal CompactValidation: %v", err)
	}

	var unmarshaled CompactValidation
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CompactValidation: %v", err)
	}

	if unmarshaled.Function != validation.Function {
		t.Errorf("Expected function %s, got %s", validation.Function, unmarshaled.Function)
	}
	if len(unmarshaled.Args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(unmarshaled.Args))
	}
	if unmarshaled.Error != false {
		t.Error("Expected error to be false")
	}
}

func TestCoverageInfo_Structure(t *testing.T) {
	info := CoverageInfo{
		Available:  10,
		Compatible: 8,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal CoverageInfo: %v", err)
	}

	var unmarshaled CoverageInfo
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CoverageInfo: %v", err)
	}

	if unmarshaled.Available != 10 {
		t.Errorf("Expected 10 available, got %d", unmarshaled.Available)
	}
	if unmarshaled.Compatible != 8 {
		t.Errorf("Expected 8 compatible, got %d", unmarshaled.Compatible)
	}
}

func TestCapabilityCoverage_Structure(t *testing.T) {
	coverage := CapabilityCoverage{
		Functions: map[config.CCLFunction]CoverageInfo{
			config.FunctionParse: {Available: 5, Compatible: 4},
		},
		Features: map[config.CCLFeature]CoverageInfo{
			config.FeatureComments: {Available: 3, Compatible: 2},
		},
	}

	data, err := json.Marshal(coverage)
	if err != nil {
		t.Fatalf("Failed to marshal CapabilityCoverage: %v", err)
	}

	var unmarshaled CapabilityCoverage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal CapabilityCoverage: %v", err)
	}

	if len(unmarshaled.Functions) != 1 {
		t.Errorf("Expected 1 function coverage, got %d", len(unmarshaled.Functions))
	}
	if len(unmarshaled.Features) != 1 {
		t.Errorf("Expected 1 feature coverage, got %d", len(unmarshaled.Features))
	}
}
