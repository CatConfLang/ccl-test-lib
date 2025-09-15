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
	Format       TestFormat                  // Source or Flat
	FilterMode   FilterMode                  // Compatible, All, or Custom
	CustomFilter func(types.TestCase) bool   // Custom filtering function
	LevelLimit   int                         // Maximum level to include (0 = no limit)
}

// TestFormat specifies which test format to load
type TestFormat int

const (
	FormatSource TestFormat = iota // tests/*.json (maintainable)
	FormatFlat                     // generated-tests/ (implementation-friendly)
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
	case FormatSource:
		testDir = filepath.Join(tl.TestDataPath, "tests")
		pattern = "*.json"
	case FormatFlat:
		testDir = filepath.Join(tl.TestDataPath, "generated-tests")
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
	
	// Handle flat format (array of TestCase) vs source format (TestSuite)
	if opts.Format == FormatFlat {
		// Try to parse as array of TestCase first (current ccl-test-data flat format)
		var tests []types.TestCase
		if err := json.Unmarshal(data, &tests); err == nil {
			// Successfully parsed as array - create wrapper TestSuite
			suite = types.TestSuite{
				Suite:   "Generated Flat Format",
				Version: "1.0",
				Tests:   tests,
			}
		} else {
			// Fall back to TestSuite format
			if err := json.Unmarshal(data, &suite); err != nil {
				return nil, fmt.Errorf("failed to parse JSON as both array and suite: %w", err)
			}
		}
	} else {
		// Source format - always TestSuite
		if err := json.Unmarshal(data, &suite); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	// Apply level filtering if specified
	if opts.LevelLimit > 0 {
		var filteredTests []types.TestCase
		for _, test := range suite.Tests {
			if test.Meta.Level <= opts.LevelLimit {
				filteredTests = append(filteredTests, test)
			}
		}
		suite.Tests = filteredTests
	}

	return &suite, nil
}

// LoadTestsByLevel loads tests filtered by implementation level
func (tl *TestLoader) LoadTestsByLevel(level int, opts LoadOptions) ([]types.TestCase, error) {
	opts.LevelLimit = level
	return tl.LoadAllTests(opts)
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
		ByLevel:         make(map[int]int),
		ByFunction:      make(map[string]int),
		ByFeature:       make(map[string]int),
	}

	compatibleTests := tl.FilterCompatibleTests(tests)
	stats.CompatibleTests = len(compatibleTests)
	stats.CompatibleAsserts = len(compatibleTests)

	for _, test := range tests {
		// Level statistics
		stats.ByLevel[test.Meta.Level]++

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
		fnStr := string(fn)
		for _, test := range allTests {
			if test.Validation == fnStr {
				fnTests++
				continue
			}
			for _, testFn := range test.Functions {
				if testFn == fnStr {
					fnTests++
					break
				}
			}
		}
		coverage.Functions[fn] = CoverageInfo{
			Available: fnTests,
			Compatible: len(tl.FilterCompatibleTests(allTests)),
		}
	}

	// Analyze feature coverage
	for _, feature := range tl.Config.SupportedFeatures {
		featureTests := 0
		featureStr := string(feature)
		for _, test := range allTests {
			for _, testFeature := range test.Features {
				if testFeature == featureStr {
					featureTests++
					break
				}
			}
		}
		coverage.Features[feature] = CoverageInfo{
			Available:  featureTests,
			Compatible: len(tl.FilterCompatibleTests(allTests)),
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