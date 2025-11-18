// Package loader provides modern test loading infrastructure with type-safe filtering
// supporting both source and flat format CCL test suites.
package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/types"
)

// TestLoader handles both source and flat format loading with type-safe filtering
type TestLoader struct {
	TestDataPath string
	Config       config.ImplementationConfig
	UseFlat      bool // true = generated flat format, false = source format
}

// LoadOptions controls test loading behavior
type LoadOptions struct {
	Format       TestFormat                // Source or Flat
	FilterMode   FilterMode                // Compatible, All, or Custom
	CustomFilter func(types.TestCase) bool // Custom filtering function
}

// TestFormat specifies which test format to load
type TestFormat int

const (
	FormatCompact TestFormat = iota // source_tests/*.json (compact arrays)
	FormatFlat                      // generated-tests/ (implementation-friendly)
)

// FilterMode specifies how tests should be filtered
type FilterMode int

const (
	FilterCompatible FilterMode = iota // Only tests compatible with config
	FilterAll                          // All tests (no filtering)
	FilterCustom                       // Use custom filter function
)

// NewTestLoader creates a new test loader instance
func NewTestLoader(testDataPath string, cfg config.ImplementationConfig) *TestLoader {
	return &TestLoader{
		TestDataPath: testDataPath,
		Config:       cfg,
		UseFlat:      true, // Default to flat format
	}
}

// LoadAllTests loads all tests from the configured test data path
func (tl *TestLoader) LoadAllTests(opts LoadOptions) ([]types.TestCase, error) {
	var testDir string
	var pattern string

	switch opts.Format {
	case FormatCompact:
		testDir = filepath.Join(tl.TestDataPath, "tests")
		pattern = "*.json"
	case FormatFlat:
		// Use TestDataPath directly - caller should provide the full path to generated_tests
		testDir = tl.TestDataPath
		pattern = "*.json"
	default:
		return nil, fmt.Errorf("unsupported test format: %v", opts.Format)
	}

	files, err := filepath.Glob(filepath.Join(testDir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to find test files: %w", err)
	}

	var allTests []types.TestCase
	for _, file := range files {
		suite, err := tl.LoadTestFile(file, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", file, err)
		}
		allTests = append(allTests, suite.Tests...)
	}

	return tl.applyFiltering(allTests, opts), nil
}

// LoadTestFile loads a single test file
func (tl *TestLoader) LoadTestFile(filename string, opts LoadOptions) (*types.TestSuite, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var suite types.TestSuite

	// Handle format detection
	if opts.Format == FormatFlat {
		// Flat format - can be either array of TestCase or object with tests array
		var tests []types.TestCase

		// Try to unmarshal as TestSuite first (object with "tests" field)
		var testSuite types.TestSuite
		if err := json.Unmarshal(data, &testSuite); err == nil && len(testSuite.Tests) > 0 {
			tests = testSuite.Tests
		} else {
			// Fallback: try as array of TestCase
			if err := json.Unmarshal(data, &tests); err != nil {
				return nil, fmt.Errorf("failed to parse flat format JSON: %w", err)
			}
		}

		// Convert structured Expected objects to simple values for flat format tests
		for i := range tests {
			tests[i].Expected = tl.extractExpectedValue(tests[i].Validation, tests[i].Expected)
		}

		suite = types.TestSuite{
			Suite:   "Flat Format",
			Version: "1.0",
			Tests:   tests,
		}
	} else {
		// Compact format - array of compact test objects
		tests, err := tl.loadCompactFormat(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse compact format: %w", err)
		}
		suite = types.TestSuite{
			Suite:   "Compact Format",
			Version: "1.0",
			Tests:   tests,
		}
	}

	return &suite, nil
}

// LoadTestsByFunction loads tests filtered by CCL function
func (tl *TestLoader) LoadTestsByFunction(fn config.CCLFunction, opts LoadOptions) ([]types.TestCase, error) {
	allTests, err := tl.LoadAllTests(opts)
	if err != nil {
		return nil, err
	}

	var filtered []types.TestCase
	fnStr := string(fn)
	for _, test := range allTests {
		// Check flat format validation field
		if test.Validation == fnStr {
			filtered = append(filtered, test)
			continue
		}
		// Check type-safe function metadata
		for _, testFn := range test.Functions {
			if testFn == fnStr {
				filtered = append(filtered, test)
				break
			}
		}
	}

	return filtered, nil
}

// applyFiltering applies the appropriate filtering based on options
func (tl *TestLoader) applyFiltering(tests []types.TestCase, opts LoadOptions) []types.TestCase {
	switch opts.FilterMode {
	case FilterCompatible:
		return tl.FilterCompatibleTests(tests)
	case FilterCustom:
		if opts.CustomFilter == nil {
			return tests
		}
		var filtered []types.TestCase
		for _, test := range tests {
			if opts.CustomFilter(test) {
				filtered = append(filtered, test)
			}
		}
		return filtered
	case FilterAll:
		return tests
	default:
		return tests
	}
}

// FilterCompatibleTests filters tests based on implementation capabilities
func (tl *TestLoader) FilterCompatibleTests(tests []types.TestCase) []types.TestCase {
	var compatible []types.TestCase
	for _, test := range tests {
		if tl.IsTestCompatible(test) {
			compatible = append(compatible, test)
		}
	}
	return compatible
}

// IsTestCompatible checks if a test is compatible with the implementation
func (tl *TestLoader) IsTestCompatible(test types.TestCase) bool {
	// Check function requirements
	if test.Validation != "" {
		fn := config.CCLFunction(test.Validation)
		if !tl.Config.HasFunction(fn) {
			return false
		}
	}

	// Check functions in type-safe metadata
	for _, fnStr := range test.Functions {
		fn := config.CCLFunction(fnStr)
		if !tl.Config.HasFunction(fn) {
			return false
		}
	}

	// Check feature requirements
	for _, featureStr := range test.Features {
		feature := config.CCLFeature(featureStr)
		if !tl.Config.HasFeature(feature) {
			return false
		}
	}

	// Check behavioral conflicts
	if test.Conflicts != nil {
		for _, behaviorStr := range test.Conflicts.Behaviors {
			behavior := config.CCLBehavior(behaviorStr)
			if tl.Config.HasBehavior(behavior) {
				return false // This test conflicts with our behavior choice
			}
		}

		for _, variantStr := range test.Conflicts.Variants {
			variant := config.CCLVariant(variantStr)
			if tl.Config.HasVariant(variant) {
				return false // This test conflicts with our variant choice
			}
		}
	}

	// Check required behaviors (if specified)
	for _, behaviorStr := range test.Behaviors {
		behavior := config.CCLBehavior(behaviorStr)
		if !tl.Config.HasBehavior(behavior) {
			return false
		}
	}

	// Check required variants (if specified)
	for _, variantStr := range test.Variants {
		variant := config.CCLVariant(variantStr)
		if !tl.Config.HasVariant(variant) {
			return false
		}
	}

	return true
}

// FilterByTags provides legacy tag-based filtering for backward compatibility
func (tl *TestLoader) FilterByTags(tests []types.TestCase, includeTags, excludeTags []string) []types.TestCase {
	var filtered []types.TestCase

testLoop:
	for _, test := range tests {
		// Check exclude tags first
		for _, excludeTag := range excludeTags {
			for _, testTag := range test.Meta.Tags {
				if testTag == excludeTag {
					continue testLoop
				}
			}
		}

		// Check include tags (if any specified, at least one must match)
		if len(includeTags) > 0 {
			hasIncludeTag := false
			for _, includeTag := range includeTags {
				for _, testTag := range test.Meta.Tags {
					if testTag == includeTag {
						hasIncludeTag = true
						break
					}
				}
				if hasIncludeTag {
					break
				}
			}
			if !hasIncludeTag {
				continue testLoop
			}
		}

		filtered = append(filtered, test)
	}

	return filtered
}

// GetTestStatistics provides comprehensive test suite analysis
func (tl *TestLoader) GetTestStatistics(tests []types.TestCase) types.TestStatistics {
	stats := types.TestStatistics{
		TotalTests:      len(tests),
		TotalAssertions: len(tests), // Each test case is one assertion in flat format
		ByFunction:      make(map[string]int),
		ByFeature:       make(map[string]int),
	}

	compatibleTests := tl.FilterCompatibleTests(tests)
	stats.CompatibleTests = len(compatibleTests)
	stats.CompatibleAsserts = len(compatibleTests)

	for _, test := range tests {

		// Function statistics
		if test.Validation != "" {
			stats.ByFunction[test.Validation]++
		}
		for _, fn := range test.Functions {
			stats.ByFunction[fn]++
		}

		// Feature statistics
		for _, feature := range test.Features {
			stats.ByFeature[feature]++
		}
	}

	return stats
}

// GetCapabilityCoverage analyzes test coverage against implementation capabilities
func (tl *TestLoader) GetCapabilityCoverage() CapabilityCoverage {
	allTests, _ := tl.LoadAllTests(LoadOptions{
		Format:     FormatFlat,
		FilterMode: FilterAll,
	})

	coverage := CapabilityCoverage{
		Functions: make(map[config.CCLFunction]CoverageInfo),
		Features:  make(map[config.CCLFeature]CoverageInfo),
	}

	// Analyze function coverage
	for _, fn := range tl.Config.SupportedFunctions {
		fnTests := 0
		var functionSpecificTests []types.TestCase
		fnStr := string(fn)
		for _, test := range allTests {
			if test.Validation == fnStr {
				fnTests++
				functionSpecificTests = append(functionSpecificTests, test)
				continue
			}
			for _, testFn := range test.Functions {
				if testFn == fnStr {
					fnTests++
					functionSpecificTests = append(functionSpecificTests, test)
					break
				}
			}
		}
		coverage.Functions[fn] = CoverageInfo{
			Available:  fnTests,
			Compatible: len(tl.FilterCompatibleTests(functionSpecificTests)),
		}
	}

	// Analyze feature coverage
	for _, feature := range tl.Config.SupportedFeatures {
		featureTests := 0
		var featureSpecificTests []types.TestCase
		featureStr := string(feature)
		for _, test := range allTests {
			for _, testFeature := range test.Features {
				if testFeature == featureStr {
					featureTests++
					featureSpecificTests = append(featureSpecificTests, test)
					break
				}
			}
		}
		coverage.Features[feature] = CoverageInfo{
			Available:  featureTests,
			Compatible: len(tl.FilterCompatibleTests(featureSpecificTests)),
		}
	}

	return coverage
}

// CapabilityCoverage provides analysis of test coverage for implementation capabilities
type CapabilityCoverage struct {
	Functions map[config.CCLFunction]CoverageInfo
	Features  map[config.CCLFeature]CoverageInfo
}

// CoverageInfo provides coverage statistics for a capability
type CoverageInfo struct {
	Available  int // Total tests available for this capability
	Compatible int // Tests compatible with this implementation
}

// CompactTestFile represents the top-level structure of source test files with $schema support
type CompactTestFile struct {
	Schema string        `json:"$schema,omitempty"`
	Tests  []CompactTest `json:"tests"`
}

// CompactTest represents a test in compact format (source_tests/ files)
type CompactTest struct {
	Name      string              `json:"name"`
	Input     string              `json:"input"`
	Tests     []CompactValidation `json:"tests"`
	Features  []string            `json:"features,omitempty"`
	Behaviors []string            `json:"behaviors,omitempty"`
	Variants  []string            `json:"variants,omitempty"`
	Conflicts *types.ConflictSet  `json:"conflicts,omitempty"`
}

// CompactValidation represents a single validation in compact format
type CompactValidation struct {
	Function string      `json:"function"`
	Expect   interface{} `json:"expect"`
	Args     []string    `json:"args,omitempty"`
	Error    bool        `json:"error,omitempty"`
}

// loadCompactFormat parses compact format and converts to TestCase array
func (tl *TestLoader) loadCompactFormat(data []byte) ([]types.TestCase, error) {
	// Parse as object format with $schema and tests array
	var compactTestFile CompactTestFile
	if err := json.Unmarshal(data, &compactTestFile); err != nil {
		return nil, fmt.Errorf("failed to parse compact format JSON: %w", err)
	}

	compactTests := compactTestFile.Tests

	var testCases []types.TestCase
	for _, compact := range compactTests {
		// Convert compact test to TestCase with validations
		// Only set conflicts if they exist in the source data
		var conflicts *types.ConflictSet
		if compact.Conflicts != nil {
			// Check if the ConflictSet has any actual conflicts
			hasConflicts := len(compact.Conflicts.Functions) > 0 ||
				len(compact.Conflicts.Behaviors) > 0 ||
				len(compact.Conflicts.Variants) > 0 ||
				len(compact.Conflicts.Features) > 0
			if hasConflicts {
				conflicts = compact.Conflicts
			}
		}

		// Ensure all slice fields are never nil
		features := compact.Features
		if features == nil {
			features = make([]string, 0)
		}
		behaviors := compact.Behaviors
		if behaviors == nil {
			behaviors = make([]string, 0)
		}
		variants := compact.Variants
		if variants == nil {
			variants = make([]string, 0)
		}

		testCase := types.TestCase{
			Name:      compact.Name,
			Input:     compact.Input,
			Features:  features,
			Behaviors: behaviors,
			Variants:  variants,
			Conflicts: conflicts,
			Meta:      types.TestMetadata{},
		}

		// Create ValidationSet from compact tests array
		validations := &types.ValidationSet{}

		for _, test := range compact.Tests {
			// Create validation object with expect and args fields if present
			validationValue := createValidationObject(test)

			switch test.Function {
			case "parse":
				validations.Parse = validationValue
			case "parse_dedented":
				validations.ParseDedented = validationValue
			case "filter":
				validations.Filter = validationValue
			case "combine":
				validations.Combine = validationValue
			case "expand_dotted":
				validations.ExpandDotted = validationValue
			case "build_hierarchy":
				validations.BuildHierarchy = validationValue
			case "get_string":
				validations.GetString = validationValue
			case "get_int":
				validations.GetInt = validationValue
			case "get_bool":
				validations.GetBool = validationValue
			case "get_float":
				validations.GetFloat = validationValue
			case "get_list":
				validations.GetList = validationValue
			case "pretty_print":
				validations.PrettyPrint = validationValue
			case "round_trip":
				validations.RoundTrip = validationValue
			case "associativity":
				validations.Associativity = validationValue
			case "canonical_format":
				validations.Canonical = validationValue
			}
		}

		testCase.Validations = validations
		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

// createValidationObject creates a validation object that preserves both expect and args fields
func createValidationObject(test CompactValidation) interface{} {
	// Only typed access functions need args field
	typedAccessFunctions := map[string]bool{
		"get_string": true,
		"get_int":    true,
		"get_bool":   true,
		"get_float":  true,
		"get_list":   true,
	}

	validationObj := map[string]interface{}{
		"expect": test.Expect,
	}

	// Only include args for typed access functions
	if typedAccessFunctions[test.Function] {
		validationObj["args"] = test.Args
	}

	return validationObj
}

// extractExpectedValue extracts the appropriate value from a structured Expected object
// based on the validation type, handling both simple values and structured objects
func (tl *TestLoader) extractExpectedValue(validation string, expected interface{}) interface{} {
	// If it's already a simple value, return as-is
	if expected == nil {
		return nil
	}

	// Check if it's a structured Expected object with Count/Value/Object/etc fields
	expectedMap, isMap := expected.(map[string]interface{})
	if !isMap {
		// Simple value, return as-is
		return expected
	}

	// Check if it has the structured format fields
	_, hasCount := expectedMap["count"]
	if !hasCount {
		// Not a structured format, return as-is
		return expected
	}

	// Extract the appropriate field based on validation type
	switch validation {
	case "parse", "parse_dedented", "filter", "compose", "expand_dotted":
		// These expect entries
		if entries, ok := expectedMap["entries"]; ok {
			return entries
		}
	case "build_hierarchy":
		// Expects an object
		if object, ok := expectedMap["object"]; ok {
			return object
		}
	case "get_string", "get_int", "get_bool", "get_float":
		// Typed access expects a single value
		if value, ok := expectedMap["value"]; ok {
			return value
		}
	case "get_list":
		// List access expects a list
		if list, ok := expectedMap["list"]; ok {
			return list
		}
	}

	// Fallback: return the original expected value
	return expected
}
