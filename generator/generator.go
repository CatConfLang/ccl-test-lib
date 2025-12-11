// Package generator provides utilities for transforming source format
// CCL tests to implementation-friendly flat format.
package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/CatConfLang/ccl-test-lib/config"
	"github.com/CatConfLang/ccl-test-lib/loader"
	"github.com/CatConfLang/ccl-test-lib/types"
	"github.com/CatConfLang/ccl-test-lib/types/generated"
)

// Export format constants for convenience
const (
	FormatCompact = loader.FormatCompact
	FormatFlat    = loader.FormatFlat
)

// FlatGenerator transforms source format to implementation-friendly flat format
type FlatGenerator struct {
	SourceDir string
	OutputDir string
	Options   GenerateOptions
}

// GenerateOptions controls flat format generation behavior
type GenerateOptions struct {
	SkipPropertyTests bool                 // Skip property-*.json files
	SkipFunctions     []config.CCLFunction // Skip specific functions
	OnlyFunctions     []config.CCLFunction // Generate only these functions
	SourceFormat      loader.TestFormat    // Input format (compact or flat)
	Verbose           bool                 // Enable verbose output
}

// NewFlatGenerator creates a new flat format generator
func NewFlatGenerator(sourceDir, outputDir string, opts GenerateOptions) *FlatGenerator {
	return &FlatGenerator{
		SourceDir: sourceDir,
		OutputDir: outputDir,
		Options:   opts,
	}
}

// GenerateAll processes all source test files and generates flat format
func (fg *FlatGenerator) GenerateAll() error {
	if err := os.MkdirAll(fg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	pattern := filepath.Join(fg.SourceDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find source files: %w", err)
	}

	for _, file := range files {
		basename := filepath.Base(file)

		// Skip property tests if requested
		if fg.Options.SkipPropertyTests && strings.HasPrefix(basename, "property-") {
			if fg.Options.Verbose {
				fmt.Printf("Skipping property test file: %s\n", basename)
			}
			continue
		}

		if err := fg.GenerateFile(file); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file, err)
		}

		if fg.Options.Verbose {
			fmt.Printf("Generated flat format for: %s\n", basename)
		}
	}

	return nil
}

// GenerateFile processes a single source file
func (fg *FlatGenerator) GenerateFile(sourceFile string) error {
	// Use loader to handle format detection and parsing
	testLoader := loader.NewTestLoader("", config.ImplementationConfig{})

	sourceSuite, err := testLoader.LoadTestFile(sourceFile, loader.LoadOptions{
		Format:     fg.Options.SourceFormat,
		FilterMode: loader.FilterAll,
	})
	if err != nil {
		return fmt.Errorf("failed to load source file: %w", err)
	}

	// Transform to flat format
	flatSuite := types.TestSuite{
		Suite:       sourceSuite.Suite,
		Version:     sourceSuite.Version,
		Description: sourceSuite.Description + " (flat format)",
		Tests:       []types.TestCase{},
	}

	for _, sourceTest := range sourceSuite.Tests {
		flatTests, err := fg.TransformSourceToFlat(sourceTest)
		if err != nil {
			return fmt.Errorf("failed to transform test %s: %w", sourceTest.Name, err)
		}
		flatSuite.Tests = append(flatSuite.Tests, flatTests...)
	}

	// Apply filtering options
	flatSuite.Tests = fg.applyFiltering(flatSuite.Tests)

	// Convert to generated flat format types (array of flat test cases)
	var flatTests []generated.GeneratedFormatSimpleJsonTestsElem
	for _, test := range flatSuite.Tests {
		flatTest := fg.convertToFlatFormat(test)
		flatTests = append(flatTests, flatTest)
	}

	// Create object format with $schema at top level
	wrapper := generated.GeneratedFormatSimpleJson{
		Schema: "http://json-schema.org/draft-07/schema#",
		Tests:  flatTests,
	}

	// Write flat format file
	outputFile := filepath.Join(fg.OutputDir, filepath.Base(sourceFile))
	flatData, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal flat JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, flatData, 0644); err != nil {
		return fmt.Errorf("failed to write flat file: %w", err)
	}

	return nil
}

// TransformSourceToFlat transforms a source test to multiple flat tests (1:N transformation)
func (fg *FlatGenerator) TransformSourceToFlat(sourceTest types.TestCase) ([]types.TestCase, error) {
	if sourceTest.Validations == nil {
		// Already flat format or no validations
		return []types.TestCase{sourceTest}, nil
	}

	var flatTests []types.TestCase
	validations := sourceTest.Validations

	// Use reflection to iterate over validation fields
	v := reflect.ValueOf(validations).Elem()
	t := reflect.TypeOf(validations).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.IsNil() {
			continue // Skip nil validations
		}

		// Use JSON tag if available, otherwise convert field name
		validationName := getValidationName(fieldType)

		// Parse the validation value to extract components (args, expect, error)
		validationComponents := parseValidationValue(field.Interface())

		// Create flat test for this validation
		flatTest := types.TestCase{
			Name:        fmt.Sprintf("%s_%s", sourceTest.Name, validationName),
			Inputs:      sourceTest.Inputs,
			Validation:  validationName,
			Expected:    validationComponents.Expected,
			Args:        validationComponents.Args,
			ExpectError: validationComponents.Error,
			Meta:        sourceTest.Meta,
			SourceTest:  sourceTest.Name,
		}

		// Extract and populate type-safe metadata
		generatedFunctions, generatedFeatures := fg.GenerateMetadataFromValidation(validationName)
		flatTest.Functions = generatedFunctions

		// Merge generated features with source features, ensuring never nil and no duplicates
		flatTest.Features = make([]string, 0)
		if sourceTest.Features != nil {
			flatTest.Features = append(flatTest.Features, sourceTest.Features...)
		}
		if generatedFeatures != nil {
			flatTest.Features = append(flatTest.Features, generatedFeatures...)
		}
		// Remove duplicates by using a map
		seen := make(map[string]bool)
		uniqueFeatures := make([]string, 0, len(flatTest.Features))
		for _, feature := range flatTest.Features {
			if !seen[feature] {
				seen[feature] = true
				uniqueFeatures = append(uniqueFeatures, feature)
			}
		}
		flatTest.Features = uniqueFeatures

		// Filter behaviors to only include those relevant to this validation function.
		// This ensures function-specific behaviors (like boolean_strict/lenient) are
		// only tagged on functions where they actually affect behavior.
		flatTest.Behaviors = filterBehaviorsForFunction(sourceTest.Behaviors, validationName)

		// Copy variants from source, ensuring never nil
		if sourceTest.Variants != nil {
			flatTest.Variants = sourceTest.Variants
		} else {
			flatTest.Variants = make([]string, 0)
		}

		// Filter conflicts to only include behavior conflicts relevant to this function
		flatTest.Conflicts = filterConflictsForFunction(sourceTest.Conflicts, validationName)

		// Validation components are already parsed and applied above
		// No special case handling needed - all validation types are handled uniformly

		flatTests = append(flatTests, flatTest)
	}

	return flatTests, nil
}

// GenerateMetadataFromValidation creates type-safe metadata from validation type
func (fg *FlatGenerator) GenerateMetadataFromValidation(validationName string) (functions []string, features []string) {
	// Map validation names to functions
	functions = []string{validationName}

	// Initialize features as empty slice, never nil
	features = make([]string, 0)

	// Map validation names to required features
	switch validationName {
	case "filter":
		features = append(features, string(config.FeatureComments))
	case "expand_dotted":
		features = append(features, string(config.FeatureExperimentalDottedKeys))
	}

	return functions, features
}

// ExtractMetadataFromTags extracts typed metadata from legacy tags
func ExtractMetadataFromTags(tags []string) (functions, features, behaviors, variants []string) {
	for _, tag := range tags {
		switch {
		case strings.HasPrefix(tag, "function:"):
			functions = append(functions, strings.TrimPrefix(tag, "function:"))
		case strings.HasPrefix(tag, "feature:"):
			features = append(features, strings.TrimPrefix(tag, "feature:"))
		case strings.HasPrefix(tag, "behavior:"):
			behaviors = append(behaviors, strings.TrimPrefix(tag, "behavior:"))
		case strings.HasPrefix(tag, "variant:"):
			variants = append(variants, strings.TrimPrefix(tag, "variant:"))
		}
	}
	return
}

// ValidateGenerated validates the generated flat format files
func (fg *FlatGenerator) ValidateGenerated() error {
	pattern := filepath.Join(fg.OutputDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find generated files: %w", err)
	}

	for _, file := range files {
		if err := fg.validateFile(file); err != nil {
			return fmt.Errorf("validation failed for %s: %w", file, err)
		}
	}

	return nil
}

// applyFiltering applies generation options to filter tests
func (fg *FlatGenerator) applyFiltering(tests []types.TestCase) []types.TestCase {
	var filtered []types.TestCase

	for _, test := range tests {
		var skip bool

		// Skip functions if specified
		if len(fg.Options.SkipFunctions) > 0 {
			skip = false
			for _, skipFn := range fg.Options.SkipFunctions {
				if test.Validation == string(skipFn) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// Include only specified functions if set
		if len(fg.Options.OnlyFunctions) > 0 {
			include := false
			for _, onlyFn := range fg.Options.OnlyFunctions {
				if test.Validation == string(onlyFn) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		filtered = append(filtered, test)
	}

	return filtered
}

// validateFile validates a single generated file
func (fg *FlatGenerator) validateFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var suite types.TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate each test has required flat format fields
	for _, test := range suite.Tests {
		if test.Validation == "" {
			return fmt.Errorf("test %s missing validation field", test.Name)
		}
		if test.Expected == nil {
			return fmt.Errorf("test %s missing expected field", test.Name)
		}
	}

	return nil
}

// convertToFlatFormat converts old TestCase to generated flat format with proper Expected structure
func (fg *FlatGenerator) convertToFlatFormat(test types.TestCase) generated.GeneratedFormatSimpleJsonTestsElem {
	// Create the proper Expected structure based on validation type
	expected := fg.createExpectedStructure(test.Validation, test.Expected)

	// Convert behaviors, features, variants to the generated enum types
	// Ensure these are never nil - initialize as empty if needed
	testBehaviors := test.Behaviors
	if testBehaviors == nil {
		testBehaviors = make([]string, 0)
	}
	testFeatures := test.Features
	if testFeatures == nil {
		testFeatures = make([]string, 0)
	}
	testVariants := test.Variants
	if testVariants == nil {
		testVariants = make([]string, 0)
	}
	testFunctions := test.Functions
	if testFunctions == nil {
		testFunctions = make([]string, 0)
	}

	behaviors := fg.convertBehaviors(testBehaviors)
	features := fg.convertFeatures(testFeatures)
	variants := fg.convertVariants(testVariants)
	functions := fg.convertFunctions(testFunctions)

	// Create the flat test directly using the generated type
	flatTest := generated.GeneratedFormatSimpleJsonTestsElem{
		Name:       test.Name,
		Inputs:     test.Inputs,
		Validation: generated.GeneratedFormatSimpleJsonTestsElemValidation(test.Validation),
		Expected:   expected,
		Functions:  functions,
		Features:   features,
		Behaviors:  behaviors,
		Variants:   variants,
		Args:       fg.getArgsForValidation(test.Validation, test.Args),
		SourceTest: &test.SourceTest,
	}

	return flatTest
}

// getArgsForValidation returns args only for typed access functions, nil for others
func (fg *FlatGenerator) getArgsForValidation(validation string, args []string) []string {
	// Only typed access functions need args field
	typedAccessFunctions := map[string]bool{
		"get_string": true,
		"get_int":    true,
		"get_bool":   true,
		"get_float":  true,
		"get_list":   true,
	}

	if typedAccessFunctions[validation] {
		// For typed access functions, return args (even if empty)
		return args
	}

	// For other functions, return nil so omitempty will omit the field
	return nil
}

// createExpectedStructure creates the proper Expected object with Count and data fields
func (fg *FlatGenerator) createExpectedStructure(validation string, data interface{}) generated.GeneratedFormatSimpleJsonTestsElemExpected {
	expected := generated.GeneratedFormatSimpleJsonTestsElemExpected{}

	switch validation {
	case "parse", "parse_indented", "filter", "compose", "expand_dotted":
		// These validations expect entries (key-value pairs)
		if entries, ok := data.([]interface{}); ok {
			expected.Count = len(entries)
			var entryList []generated.GeneratedFormatSimpleJsonTestsElemExpectedEntriesElem
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					if key, hasKey := entryMap["key"].(string); hasKey {
						if value, hasValue := entryMap["value"].(string); hasValue {
							entryList = append(entryList, generated.GeneratedFormatSimpleJsonTestsElemExpectedEntriesElem{
								Key:   key,
								Value: value,
							})
						}
					}
				}
			}
			expected.Entries = entryList
		}
	case "build_hierarchy":
		// Hierarchy expects an object
		expected.Count = 1
		expected.Object = data
	case "get_string", "get_int", "get_bool", "get_float":
		// Typed access expects a single value
		expected.Count = 1
		expected.Value = data
	case "get_list":
		// List access expects a list
		if list, ok := data.([]interface{}); ok {
			expected.Count = len(list)
			expected.List = list
		}
	default:
		// Default case - try to infer from data type
		expected.Count = 1
		expected.Value = data
	}

	return expected
}

// Helper functions for converting enum types
func (fg *FlatGenerator) convertBehaviors(behaviors []string) []generated.GeneratedFormatSimpleJsonTestsElemBehaviorsElem {
	result := make([]generated.GeneratedFormatSimpleJsonTestsElemBehaviorsElem, 0, len(behaviors))
	for _, b := range behaviors {
		result = append(result, generated.GeneratedFormatSimpleJsonTestsElemBehaviorsElem(b))
	}
	return result
}

func (fg *FlatGenerator) convertFeatures(features []string) []string {
	// Features is just []string in the simplified schema (no enum constraints)
	if features == nil {
		return make([]string, 0)
	}
	return features
}

func (fg *FlatGenerator) convertVariants(variants []string) []generated.GeneratedFormatSimpleJsonTestsElemVariantsElem {
	result := make([]generated.GeneratedFormatSimpleJsonTestsElemVariantsElem, 0, len(variants))
	for _, v := range variants {
		result = append(result, generated.GeneratedFormatSimpleJsonTestsElemVariantsElem(v))
	}
	return result
}

func (fg *FlatGenerator) convertFunctions(functions []string) []generated.GeneratedFormatSimpleJsonTestsElemFunctionsElem {
	result := make([]generated.GeneratedFormatSimpleJsonTestsElemFunctionsElem, 0, len(functions))
	for _, f := range functions {
		result = append(result, generated.GeneratedFormatSimpleJsonTestsElemFunctionsElem(f))
	}
	return result
}

// Helper functions

// getValidationName extracts the validation name from JSON tag or field name
func getValidationName(fieldType reflect.StructField) string {
	// Check for JSON tag first
	if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
		// Remove ",omitempty" suffix if present
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			return jsonTag[:idx]
		}
		return jsonTag
	}
	// Fallback to camelToSnake conversion of field name
	return camelToSnake(fieldType.Name)
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// ValidationComponents represents the parsed components of a validation value
type ValidationComponents struct {
	Expected interface{}
	Args     []string
	Error    bool
}

// parseValidationValue parses a validation value that may be either:
// - A simple expected value (legacy format)
// - A structured validation object with args, expect, error fields (source format)
func parseValidationValue(value interface{}) ValidationComponents {
	// Try to parse as structured validation object first
	if validationMap, ok := value.(map[string]interface{}); ok {
		result := ValidationComponents{
			Expected: value, // Default to the whole object
			Args:     []string{},
			Error:    false,
		}

		// Extract expect field if present
		if expect, hasExpect := validationMap["expect"]; hasExpect {
			result.Expected = expect
		}

		// Extract args field if present
		if argsInterface, hasArgs := validationMap["args"]; hasArgs {
			if argsSlice, ok := argsInterface.([]interface{}); ok {
				for _, arg := range argsSlice {
					if argStr, ok := arg.(string); ok {
						result.Args = append(result.Args, argStr)
					}
				}
			} else if argsStringSlice, ok := argsInterface.([]string); ok {
				result.Args = argsStringSlice
			}
		}

		// Extract error field if present
		if errorVal, hasError := validationMap["error"]; hasError {
			if errorBool, ok := errorVal.(bool); ok {
				result.Error = errorBool
			}
		}

		return result
	}

	// Fallback to treating value as expected result (legacy format)
	return ValidationComponents{
		Expected: value,
		Args:     []string{},
		Error:    expectErrorFromValue(value),
	}
}

// expectErrorFromValue checks if a validation value indicates an error expectation
func expectErrorFromValue(value interface{}) bool {
	if str, ok := value.(string); ok {
		return strings.Contains(strings.ToLower(str), "error") ||
			strings.Contains(strings.ToLower(str), "invalid")
	}
	return false
}

// behaviorFunctionMap defines which behaviors apply to which functions.
// Behaviors not listed here apply to all functions (global behaviors).
// This mapping ensures that function-specific behaviors like boolean_strict/lenient
// are only tagged on the functions where they actually affect behavior.
var behaviorFunctionMap = map[string][]string{
	// Boolean parsing behavior only affects get_bool
	"boolean_strict":  {"get_bool"},
	"boolean_lenient": {"get_bool"},

	// List coercion only affects get_list
	"list_coercion_enabled":  {"get_list"},
	"list_coercion_disabled": {"get_list"},

	// CRLF handling affects parsing and formatting functions
	"crlf_preserve_literal": {"parse", "parse_indented", "canonical_format", "load"},
	"crlf_normalize_to_lf":  {"parse", "parse_indented", "canonical_format", "load"},

	// Tab handling affects parsing, formatting, and hierarchy building functions
	"tabs_preserve":  {"parse", "parse_indented", "canonical_format", "load", "build_hierarchy"},
	"tabs_to_spaces": {"parse", "parse_indented", "canonical_format", "load", "build_hierarchy"},

	// Spacing behavior affects parsing
	"strict_spacing": {"parse", "parse_indented"},
	"loose_spacing":  {"parse", "parse_indented"},

	// Array ordering affects hierarchy building and list access
	"array_order_insertion":     {"build_hierarchy", "get_list"},
	"array_order_lexicographic": {"build_hierarchy", "get_list"},
}

// filterBehaviorsForFunction filters behaviors to only include those relevant
// to the given validation function. Behaviors not in behaviorFunctionMap are
// considered global and always included.
func filterBehaviorsForFunction(behaviors []string, validationName string) []string {
	if behaviors == nil {
		return make([]string, 0)
	}

	filtered := make([]string, 0, len(behaviors))
	for _, behavior := range behaviors {
		applicableFunctions, hasMapping := behaviorFunctionMap[behavior]
		if !hasMapping {
			// Behavior not in map = global behavior, always include
			filtered = append(filtered, behavior)
			continue
		}

		// Check if this validation function is in the applicable list
		for _, fn := range applicableFunctions {
			if fn == validationName {
				filtered = append(filtered, behavior)
				break
			}
		}
	}

	return filtered
}

// filterConflictsForFunction filters conflict behaviors to only include those
// relevant to the given validation function.
func filterConflictsForFunction(conflicts *types.ConflictSet, validationName string) *types.ConflictSet {
	if conflicts == nil {
		return nil
	}

	// Filter behavior conflicts
	filteredBehaviors := filterBehaviorsForFunction(conflicts.Behaviors, validationName)

	// Check if we still have any conflicts after filtering
	hasConflicts := len(conflicts.Functions) > 0 ||
		len(filteredBehaviors) > 0 ||
		len(conflicts.Variants) > 0 ||
		len(conflicts.Features) > 0

	if !hasConflicts {
		return nil
	}

	return &types.ConflictSet{
		Functions: conflicts.Functions,
		Behaviors: filteredBehaviors,
		Variants:  conflicts.Variants,
		Features:  conflicts.Features,
	}
}
