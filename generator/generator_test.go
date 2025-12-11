package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CatConfLang/ccl-test-lib/config"
	"github.com/CatConfLang/ccl-test-lib/loader"
	"github.com/CatConfLang/ccl-test-lib/types"
	"github.com/CatConfLang/ccl-test-lib/types/generated"
)

// Test data setup
func setupGeneratorTestData(t *testing.T) (string, string) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create compact format test file first
	compactTests := []loader.CompactTest{
		{
			Name:     "multi_validation_test",
			Input:    "key = value\ncount = 42",
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect: []map[string]interface{}{
						{"key": "key", "value": "value"},
						{"key": "count", "value": "42"},
					},
				},
				{
					Function: "build_hierarchy",
					Expect: map[string]interface{}{
						"key":   "value",
						"count": "42",
					},
				},
				{
					Function: "get_string",
					Args:     []string{"key"},
					Expect:   "value",
				},
				{
					Function: "get_int",
					Args:     []string{"count"},
					Expect:   42,
				},
			},
		},
		{
			Name:     "single_validation_test",
			Input:    "flag = true",
			Features: []string{},
			Tests: []loader.CompactValidation{
				{
					Function: "get_bool",
					Args:     []string{"flag"},
					Expect:   true,
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
	if err := os.WriteFile(filepath.Join(sourceDir, "test-source.json"), sourceData, 0644); err != nil {
		t.Fatalf("Failed to write source test file: %v", err)
	}

	// Create a different compact format test file
	compactTests2 := []loader.CompactTest{
		{
			Name:     "compact_test",
			Input:    "name = test",
			Features: []string{"multiline"},
			Tests: []loader.CompactValidation{
				{
					Function: "parse",
					Expect:   []map[string]interface{}{{"key": "name", "value": "test"}},
				},
				{
					Function: "get_string",
					Args:     []string{"name"},
					Expect:   "test",
				},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	compactTestFile2 := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  compactTests2,
	}
	compactData, _ := json.MarshalIndent(compactTestFile2, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "test-compact.json"), compactData, 0644); err != nil {
		t.Fatalf("Failed to write compact test file: %v", err)
	}

	// Create property test file (should be skipped when SkipPropertyTests is true)
	propertyTests := []loader.CompactTest{
		{
			Name:     "property_test",
			Input:    "a = 1",
			Features: []string{},
			Tests: []loader.CompactValidation{
				{
					Function: "round_trip",
					Expect:   "a = 1",
				},
			},
		},
	}

	// Wrap in CompactTestFile structure for correct parsing
	propertyTestFile := loader.CompactTestFile{
		Schema: "https://schemas.ccl.example.com/compact-format/v1.0.json",
		Tests:  propertyTests,
	}
	propertyData, _ := json.MarshalIndent(propertyTestFile, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "property-test.json"), propertyData, 0644); err != nil {
		t.Fatalf("Failed to write property test file: %v", err)
	}

	return sourceDir, outputDir
}

func TestNewFlatGenerator(t *testing.T) {
	sourceDir := "/source"
	outputDir := "/output"
	opts := GenerateOptions{
		Verbose: true,
	}

	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	if generator.SourceDir != sourceDir {
		t.Errorf("Expected source dir %s, got %s", sourceDir, generator.SourceDir)
	}
	if generator.OutputDir != outputDir {
		t.Errorf("Expected output dir %s, got %s", outputDir, generator.OutputDir)
	}
	if !generator.Options.Verbose {
		t.Error("Expected verbose option to be true")
	}
}

func TestFlatGenerator_GenerateFile_SourceFormat(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SourceFormat: FormatCompact, // Actually loading from our ValidationSet format
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	sourceFile := filepath.Join(sourceDir, "test-source.json")
	if err := generator.GenerateFile(sourceFile); err != nil {
		t.Fatalf("Failed to generate file: %v", err)
	}

	// Verify output file exists
	outputFile := filepath.Join(outputDir, "test-source.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Load and verify output content
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var wrapper generated.GeneratedFormatSimpleJson
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	if wrapper.Schema != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("Expected schema field, got %s", wrapper.Schema)
	}

	if len(wrapper.Tests) < 1 {
		t.Error("Expected at least 1 generated test")
	}

	// Verify first test structure (TestItem is an interface, so we can't check fields directly)
	// Just verify we have tests
	if len(wrapper.Tests) == 0 {
		t.Error("Expected at least one test in generated format")
	}
}

func TestFlatGenerator_GenerateFile_CompactFormat(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SourceFormat: FormatCompact,
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	compactFile := filepath.Join(sourceDir, "test-compact.json")
	if err := generator.GenerateFile(compactFile); err != nil {
		t.Fatalf("Failed to generate from compact file: %v", err)
	}

	// Verify output file exists
	outputFile := filepath.Join(outputDir, "test-compact.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Load and verify content has multiple tests (one per validation)
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var wrapper generated.GeneratedFormatSimpleJson
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("Failed to unmarshal generated JSON: %v", err)
	}

	if len(wrapper.Tests) != 2 {
		t.Errorf("Expected 2 generated tests (one per validation), got %d", len(wrapper.Tests))
	}

}

func TestFlatGenerator_GenerateAll(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SourceFormat: FormatCompact,
		Verbose:      true,
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	if err := generator.GenerateAll(); err != nil {
		t.Fatalf("Failed to generate all files: %v", err)
	}

	// Verify all expected output files exist
	expectedFiles := []string{"test-source.json", "test-compact.json", "property-test.json"}
	for _, expectedFile := range expectedFiles {
		outputFile := filepath.Join(outputDir, expectedFile)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Expected output file %s was not created", expectedFile)
		}
	}
}

func TestFlatGenerator_GenerateAll_SkipPropertyTests(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SourceFormat:      FormatCompact,
		SkipPropertyTests: true,
		Verbose:           true,
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	if err := generator.GenerateAll(); err != nil {
		t.Fatalf("Failed to generate all files: %v", err)
	}

	// Verify property test file was NOT created
	propertyFile := filepath.Join(outputDir, "property-test.json")
	if _, err := os.Stat(propertyFile); !os.IsNotExist(err) {
		t.Error("Property test file should have been skipped")
	}

	// Verify other files were created
	expectedFiles := []string{"test-source.json", "test-compact.json"}
	for _, expectedFile := range expectedFiles {
		outputFile := filepath.Join(outputDir, expectedFile)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Expected output file %s was not created", expectedFile)
		}
	}
}

func TestFlatGenerator_TransformSourceToFlat(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	sourceTest := types.TestCase{
		Name:  "transform_test",
		Input: "key = value",
		Validations: &types.ValidationSet{
			Parse: []map[string]interface{}{
				{"key": "key", "value": "value"},
			},
			GetString: map[string]interface{}{
				"args":   []string{"key"},
				"expect": "value",
			},
		},
		Features: []string{"comments"},
	}

	flatTests, err := generator.TransformSourceToFlat(sourceTest)
	if err != nil {
		t.Fatalf("Failed to transform source to flat: %v", err)
	}

	if len(flatTests) != 2 {
		t.Errorf("Expected 2 flat tests, got %d", len(flatTests))
	}

	// Verify first test (parse)
	parseTest := flatTests[0]
	if parseTest.Validation != "parse" {
		t.Errorf("Expected parse validation, got %s", parseTest.Validation)
	}
	if parseTest.SourceTest != "transform_test" {
		t.Errorf("Expected source test 'transform_test', got %s", parseTest.SourceTest)
	}
	if len(parseTest.Functions) != 1 || parseTest.Functions[0] != "parse" {
		t.Error("Expected functions metadata to contain 'parse'")
	}

	// Verify second test (get_string)
	getStringTest := flatTests[1]
	if getStringTest.Validation != "get_string" {
		t.Errorf("Expected get_string validation, got %s", getStringTest.Validation)
	}
	if len(getStringTest.Args) != 1 || getStringTest.Args[0] != "key" {
		t.Error("Expected args to contain 'key'")
	}
}

func TestFlatGenerator_TransformSourceToFlat_AlreadyFlat(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	// Test with already flat test (no Validations)
	flatTest := types.TestCase{
		Name:       "already_flat",
		Input:      "key = value",
		Validation: "parse",
		Expected:   []map[string]interface{}{{"key": "key", "value": "value"}},
	}

	result, err := generator.TransformSourceToFlat(flatTest)
	if err != nil {
		t.Fatalf("Failed to transform already flat test: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 test (unchanged), got %d", len(result))
	}

	if result[0].Name != "already_flat" {
		t.Error("Test should be returned unchanged")
	}
}

func TestFlatGenerator_TransformSourceToFlat_WithVariants(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	sourceTest := types.TestCase{
		Name:  "test_with_variants",
		Input: "key = value",
		Validations: &types.ValidationSet{
			Parse: []map[string]interface{}{
				{"key": "key", "value": "value"},
			},
		},
		Features:  []string{"comments"},
		Behaviors: []string{"crlf_normalize_to_lf"}, // Use a behavior that applies to parse
		Variants:  []string{"proposed_behavior", "reference_compliant"},
	}

	flatTests, err := generator.TransformSourceToFlat(sourceTest)
	if err != nil {
		t.Fatalf("Failed to transform source to flat: %v", err)
	}

	if len(flatTests) != 1 {
		t.Errorf("Expected 1 flat test, got %d", len(flatTests))
	}

	flatTest := flatTests[0]

	// Verify that variants are copied from source to flat
	if len(flatTest.Variants) != 2 {
		t.Errorf("Expected 2 variants, got %d: %v", len(flatTest.Variants), flatTest.Variants)
	}

	expectedVariants := []string{"proposed_behavior", "reference_compliant"}
	for i, expected := range expectedVariants {
		if i >= len(flatTest.Variants) || flatTest.Variants[i] != expected {
			t.Errorf("Expected variant %s at index %d, got %v", expected, i, flatTest.Variants)
		}
	}

	// Verify behaviors are filtered correctly - crlf_normalize_to_lf applies to parse
	if len(flatTest.Behaviors) != 1 || flatTest.Behaviors[0] != "crlf_normalize_to_lf" {
		t.Errorf("Expected behaviors [crlf_normalize_to_lf], got %v", flatTest.Behaviors)
	}

	if len(flatTest.Features) != 1 || flatTest.Features[0] != "comments" {
		t.Errorf("Expected features [comments], got %v", flatTest.Features)
	}
}

func TestFlatGenerator_BehaviorFiltering(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	// Test that boolean_strict only applies to get_bool, not parse
	sourceTest := types.TestCase{
		Name:  "test_boolean_behavior_filtering",
		Input: "enabled = true",
		Validations: &types.ValidationSet{
			Parse: []map[string]interface{}{
				{"key": "enabled", "value": "true"},
			},
			GetBool: map[string]interface{}{
				"args":   []string{"enabled"},
				"expect": true,
			},
		},
		Behaviors: []string{"boolean_strict"},
	}

	flatTests, err := generator.TransformSourceToFlat(sourceTest)
	if err != nil {
		t.Fatalf("Failed to transform source to flat: %v", err)
	}

	if len(flatTests) != 2 {
		t.Fatalf("Expected 2 flat tests, got %d", len(flatTests))
	}

	// Find the parse and get_bool tests
	var parseTest, getBoolTest *types.TestCase
	for i := range flatTests {
		if flatTests[i].Validation == "parse" {
			parseTest = &flatTests[i]
		} else if flatTests[i].Validation == "get_bool" {
			getBoolTest = &flatTests[i]
		}
	}

	if parseTest == nil {
		t.Fatal("Expected to find parse test")
	}
	if getBoolTest == nil {
		t.Fatal("Expected to find get_bool test")
	}

	// parse should NOT have boolean_strict (it's not relevant to parsing)
	if len(parseTest.Behaviors) != 0 {
		t.Errorf("parse test should have no behaviors, got %v", parseTest.Behaviors)
	}

	// get_bool SHOULD have boolean_strict
	if len(getBoolTest.Behaviors) != 1 || getBoolTest.Behaviors[0] != "boolean_strict" {
		t.Errorf("get_bool test should have [boolean_strict], got %v", getBoolTest.Behaviors)
	}
}

func TestFlatGenerator_GenerateMetadataFromValidation(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	testCases := []struct {
		validation       string
		expectedFunction string
		expectedFeatures []string
	}{
		{"parse", "parse", []string{}},
		{"filter", "filter", []string{"comments"}},
		{"expand_dotted", "expand_dotted", []string{"experimental_dotted_keys"}},
		{"get_string", "get_string", []string{}},
		{"build_hierarchy", "build_hierarchy", []string{}},
	}

	for _, tc := range testCases {
		functions, features := generator.GenerateMetadataFromValidation(tc.validation)

		if len(functions) != 1 || functions[0] != tc.expectedFunction {
			t.Errorf("For validation %s, expected function [%s], got %v", tc.validation, tc.expectedFunction, functions)
		}

		if len(features) != len(tc.expectedFeatures) {
			t.Errorf("For validation %s, expected %d features, got %d", tc.validation, len(tc.expectedFeatures), len(features))
		}

		for i, expectedFeature := range tc.expectedFeatures {
			if i >= len(features) || features[i] != expectedFeature {
				t.Errorf("For validation %s, expected feature %s at index %d, got %v", tc.validation, expectedFeature, i, features)
			}
		}
	}
}

func TestExtractMetadataFromTags(t *testing.T) {
	tags := []string{
		"function:parse",
		"function:build_hierarchy",
		"feature:comments",
		"feature:multiline",
		"behavior:crlf_normalize_to_lf",
		"variant:proposed_behavior",
		"level:1", // Should be ignored (not a known prefix)
	}

	functions, features, behaviors, variants := ExtractMetadataFromTags(tags)

	expectedFunctions := []string{"parse", "build_hierarchy"}
	expectedFeatures := []string{"comments", "multiline"}
	expectedBehaviors := []string{"crlf_normalize_to_lf"}
	expectedVariants := []string{"proposed_behavior"}

	if len(functions) != len(expectedFunctions) {
		t.Errorf("Expected %d functions, got %d", len(expectedFunctions), len(functions))
	}
	for i, expected := range expectedFunctions {
		if i >= len(functions) || functions[i] != expected {
			t.Errorf("Expected function %s at index %d, got %v", expected, i, functions)
		}
	}

	if len(features) != len(expectedFeatures) {
		t.Errorf("Expected %d features, got %d", len(expectedFeatures), len(features))
	}
	for i, expected := range expectedFeatures {
		if i >= len(features) || features[i] != expected {
			t.Errorf("Expected feature %s at index %d, got %v", expected, i, features)
		}
	}

	if len(behaviors) != len(expectedBehaviors) {
		t.Errorf("Expected %d behaviors, got %d", len(expectedBehaviors), len(behaviors))
	}
	if len(behaviors) > 0 && behaviors[0] != expectedBehaviors[0] {
		t.Errorf("Expected behavior %s, got %s", expectedBehaviors[0], behaviors[0])
	}

	if len(variants) != len(expectedVariants) {
		t.Errorf("Expected %d variants, got %d", len(expectedVariants), len(variants))
	}
	if len(variants) > 0 && variants[0] != expectedVariants[0] {
		t.Errorf("Expected variant %s, got %s", expectedVariants[0], variants[0])
	}
}

func TestFlatGenerator_ValidateGenerated(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SourceFormat: FormatCompact,
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	// Generate files first
	if err := generator.GenerateAll(); err != nil {
		t.Fatalf("Failed to generate files: %v", err)
	}

	// Validate generated files
	if err := generator.ValidateGenerated(); err != nil {
		t.Errorf("Generated files should be valid: %v", err)
	}
}

func TestFlatGenerator_ValidateGenerated_InvalidFile(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	// Create invalid generated file
	invalidData := types.TestSuite{
		Tests: []types.TestCase{
			{
				Name: "invalid_test",
				// Missing Validation and Expected fields
			},
		},
	}

	data, _ := json.MarshalIndent(invalidData, "", "  ")
	invalidFile := filepath.Join(outputDir, "invalid.json")
	if err := os.WriteFile(invalidFile, data, 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// Validation should fail
	if err := generator.ValidateGenerated(); err == nil {
		t.Error("Expected validation to fail for invalid file")
	}
}

func TestFlatGenerator_ApplyFiltering_SkipFunctions(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		SkipFunctions: []config.CCLFunction{config.FunctionParse},
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	tests := []types.TestCase{
		{Name: "parse_test", Validation: "parse"},
		{Name: "build_test", Validation: "build_hierarchy"},
		{Name: "get_test", Validation: "get_string"},
	}

	filtered := generator.applyFiltering(tests)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests after filtering parse, got %d", len(filtered))
	}

	for _, test := range filtered {
		if test.Validation == "parse" {
			t.Error("Parse test should have been filtered out")
		}
	}
}

func TestFlatGenerator_ApplyFiltering_OnlyFunctions(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)

	opts := GenerateOptions{
		OnlyFunctions: []config.CCLFunction{config.FunctionParse, config.FunctionGetString},
	}
	generator := NewFlatGenerator(sourceDir, outputDir, opts)

	tests := []types.TestCase{
		{Name: "parse_test", Validation: "parse"},
		{Name: "build_test", Validation: "build_hierarchy"},
		{Name: "get_test", Validation: "get_string"},
	}

	filtered := generator.applyFiltering(tests)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests (parse and get_string only), got %d", len(filtered))
	}

	validations := make(map[string]bool)
	for _, test := range filtered {
		validations[test.Validation] = true
	}

	if !validations["parse"] {
		t.Error("Parse test should be included")
	}
	if !validations["get_string"] {
		t.Error("Get string test should be included")
	}
	if validations["build_hierarchy"] {
		t.Error("Build hierarchy test should be excluded")
	}
}

func TestParseValidationValue(t *testing.T) {
	// Test structured validation object
	structuredValue := map[string]interface{}{
		"expect": "expected_result",
		"args":   []interface{}{"arg1", "arg2"},
		"error":  true,
	}

	result := parseValidationValue(structuredValue)

	if result.Expected != "expected_result" {
		t.Errorf("Expected 'expected_result', got %v", result.Expected)
	}
	if len(result.Args) != 2 || result.Args[0] != "arg1" || result.Args[1] != "arg2" {
		t.Errorf("Expected args ['arg1', 'arg2'], got %v", result.Args)
	}
	if !result.Error {
		t.Error("Expected error to be true")
	}

	// Test simple value (legacy format)
	simpleValue := "simple_result"
	result = parseValidationValue(simpleValue)

	if result.Expected != "simple_result" {
		t.Errorf("Expected 'simple_result', got %v", result.Expected)
	}
	if len(result.Args) != 0 {
		t.Error("Expected empty args for simple value")
	}
	if result.Error {
		t.Error("Expected error to be false for simple value")
	}

	// Test value with error indication
	errorValue := "invalid error result"
	result = parseValidationValue(errorValue)

	if !result.Error {
		t.Error("Expected error to be true for error-indicating value")
	}
}

func TestCreateExpectedStructure(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	// Test parse validation (expects entries)
	entriesData := []interface{}{
		map[string]interface{}{"key": "k1", "value": "v1"},
		map[string]interface{}{"key": "k2", "value": "v2"},
	}

	expected := generator.createExpectedStructure("parse", entriesData)
	if expected.Count != 2 {
		t.Errorf("Expected count 2 for parse validation, got %d", expected.Count)
	}
	if len(expected.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(expected.Entries))
	}

	// Test build_hierarchy validation (expects object)
	objectData := map[string]interface{}{"key": "value"}
	expected = generator.createExpectedStructure("build_hierarchy", objectData)
	if expected.Count != 1 {
		t.Errorf("Expected count 1 for build_hierarchy validation, got %d", expected.Count)
	}
	if expected.Object == nil {
		t.Error("Expected object to be set for build_hierarchy validation")
	}

	// Test get_string validation (expects single value)
	stringData := "test_value"
	expected = generator.createExpectedStructure("get_string", stringData)
	if expected.Count != 1 {
		t.Errorf("Expected count 1 for get_string validation, got %d", expected.Count)
	}
	if expected.Value != stringData {
		t.Errorf("Expected value %s, got %v", stringData, expected.Value)
	}

	// Test get_list validation (expects list)
	listData := []interface{}{"a", "b", "c"}
	expected = generator.createExpectedStructure("get_list", listData)
	if expected.Count != 3 {
		t.Errorf("Expected count 3 for get_list validation, got %d", expected.Count)
	}
	if expected.List == nil {
		t.Error("Expected list to be preserved for get_list validation")
	} else if len(expected.List) != 3 {
		t.Errorf("Expected list with 3 elements, got %d", len(expected.List))
	}
}

func TestGetArgsForValidation(t *testing.T) {
	sourceDir, outputDir := setupGeneratorTestData(t)
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	args := []string{"key", "default"}

	// Test typed access function (should return args)
	result := generator.getArgsForValidation("get_string", args)
	if len(result) != 2 || result[0] != "key" || result[1] != "default" {
		t.Errorf("Expected args to be returned for get_string, got %v", result)
	}

	// Test non-typed function (should return nil)
	result = generator.getArgsForValidation("parse", args)
	if result != nil {
		t.Errorf("Expected nil args for parse function, got %v", result)
	}

	// Test other typed access functions
	typedFunctions := []string{"get_int", "get_bool", "get_float", "get_list"}
	for _, fn := range typedFunctions {
		result = generator.getArgsForValidation(fn, args)
		if result == nil {
			t.Errorf("Expected args to be returned for %s", fn)
		}
	}
}

// Test error conditions

func TestFlatGenerator_GenerateFile_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewFlatGenerator(tmpDir, tmpDir, GenerateOptions{})

	err := generator.GenerateFile("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestFlatGenerator_GenerateFile_InvalidJSON(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	generator := NewFlatGenerator(sourceDir, outputDir, GenerateOptions{})

	// Create invalid JSON file
	invalidFile := filepath.Join(sourceDir, "invalid.json")
	if err := os.WriteFile(invalidFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	err := generator.GenerateFile(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFlatGenerator_ValidateFile_MissingFields(t *testing.T) {
	outputDir := t.TempDir()
	generator := NewFlatGenerator("", outputDir, GenerateOptions{})

	// Create file with missing required fields
	invalidSuite := types.TestSuite{
		Tests: []types.TestCase{
			{
				Name: "invalid_test",
				// Missing Validation and Expected fields
			},
		},
	}

	data, _ := json.MarshalIndent(invalidSuite, "", "  ")
	invalidFile := filepath.Join(outputDir, "invalid.json")
	if err := os.WriteFile(invalidFile, data, 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	err := generator.validateFile(invalidFile)
	if err == nil {
		t.Error("Expected validation error for missing fields")
	}
}

// Test utility functions

func TestGetValidationName(t *testing.T) {
	// This would require importing reflect and creating StructField instances
	// For now, we'll test the camelToSnake function which is used as fallback
}

func TestCamelToSnake(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Parse", "parse"},
		{"BuildHierarchy", "build_hierarchy"},
		{"GetString", "get_string"},
		{"HTMLParser", "h_t_m_l_parser"},
		{"A", "a"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := camelToSnake(tc.input)
		if result != tc.expected {
			t.Errorf("camelToSnake(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestExpectErrorFromValue(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected bool
	}{
		{"error message", true},
		{"invalid input", true},
		{"Error occurred", true},
		{"successful result", false},
		{"normal value", false},
		{42, false},
		{true, false},
	}

	for _, tc := range testCases {
		result := expectErrorFromValue(tc.value)
		if result != tc.expected {
			t.Errorf("expectErrorFromValue(%v) = %t, expected %t", tc.value, result, tc.expected)
		}
	}
}

func TestValidationComponents_Structure(t *testing.T) {
	components := ValidationComponents{
		Expected: "test_result",
		Args:     []string{"arg1", "arg2"},
		Error:    true,
	}

	if components.Expected != "test_result" {
		t.Errorf("Expected 'test_result', got %v", components.Expected)
	}
	if len(components.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(components.Args))
	}
	if !components.Error {
		t.Error("Expected error to be true")
	}
}
