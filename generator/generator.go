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

	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/loader"
	"github.com/tylerbu/ccl-test-lib/types"
	"github.com/tylerbu/ccl-test-lib/types/generated"
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
	SkipPropertyTests bool                     // Skip property-*.json files
	SkipLevels       []int                    // Skip specific levels
	SkipFunctions    []config.CCLFunction     // Skip specific functions
	OnlyFunctions    []config.CCLFunction     // Generate only these functions
	SourceFormat     loader.TestFormat        // Input format (compact or flat)
	Verbose          bool                     // Enable verbose output
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
	var flatTests generated.GeneratedFormatJson
	for _, test := range flatSuite.Tests {
		flatTest := fg.convertToFlatFormat(test)
		flatTests = append(flatTests, flatTest)
	}

	// Write flat format file
	outputFile := filepath.Join(fg.OutputDir, filepath.Base(sourceFile))
	flatData, err := json.MarshalIndent(flatTests, "", "  ")
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

		validationName := strings.ToLower(fieldType.Name)
		// Convert CamelCase to snake_case
		validationName = camelToSnake(validationName)

		// Create flat test for this validation
		flatTest := types.TestCase{
			Name:       fmt.Sprintf("%s_%s", sourceTest.Name, validationName),
			Input:      sourceTest.Input,
			Input1:     sourceTest.Input1,
			Input2:     sourceTest.Input2,
			Input3:     sourceTest.Input3,
			Validation: validationName,
			Expected:   field.Interface(),
			Meta:       sourceTest.Meta,
			SourceTest: sourceTest.Name,
		}

		// Extract and populate type-safe metadata
		flatTest.Functions, flatTest.Features = fg.GenerateMetadataFromValidation(validationName)
		
		// Copy behaviors and variants from source, ensuring never nil
		if sourceTest.Behaviors != nil {
			flatTest.Behaviors = sourceTest.Behaviors
		} else {
			flatTest.Behaviors = make([]string, 0)
		}
		if sourceTest.Variants != nil {
			flatTest.Variants = sourceTest.Variants
		} else {
			flatTest.Variants = make([]string, 0)
		}
		
		// Only set conflicts if they exist and are non-empty
		if sourceTest.Conflicts != nil {
			// Check if the ConflictSet has any actual conflicts
			hasConflicts := len(sourceTest.Conflicts.Functions) > 0 ||
				len(sourceTest.Conflicts.Behaviors) > 0 ||
				len(sourceTest.Conflicts.Variants) > 0 ||
				len(sourceTest.Conflicts.Features) > 0
			if hasConflicts {
				flatTest.Conflicts = sourceTest.Conflicts
			}
			// If no conflicts, leave flatTest.Conflicts as nil (omitted)
		}

		// Handle special validation types
		switch validationName {
		case "parse_value":
			// parse_value tests may have args
			if sourceTest.Args != nil {
				flatTest.Args = sourceTest.Args
			}
		}

		// Check for error expectations in the validation value
		if expectErrorFromValue(field.Interface()) {
			flatTest.ExpectError = true
		}

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
		// Skip levels if specified
		skip := false
		for _, skipLevel := range fg.Options.SkipLevels {
			if test.Meta.Level == skipLevel {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

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

// FlatTestCase is the element type of GeneratedFormatJson slice
type FlatTestCase = struct {
	Schema string `json:"$schema" yaml:"$schema" mapstructure:"$schema"`
	Args []string `json:"args,omitempty" yaml:"args,omitempty" mapstructure:"args,omitempty"`
	Behaviors []generated.GeneratedFormatJsonElemBehaviorsElem `json:"behaviors" yaml:"behaviors" mapstructure:"behaviors"`
	Conflicts *generated.GeneratedFormatJsonElemConflicts `json:"conflicts,omitempty" yaml:"conflicts,omitempty" mapstructure:"conflicts,omitempty"`
	ErrorType *string `json:"error_type,omitempty" yaml:"error_type,omitempty" mapstructure:"error_type,omitempty"`
	ExpectError bool `json:"expect_error,omitempty" yaml:"expect_error,omitempty" mapstructure:"expect_error,omitempty"`
	Expected generated.GeneratedFormatJsonElemExpected `json:"expected" yaml:"expected" mapstructure:"expected"`
	Features []generated.GeneratedFormatJsonElemFeaturesElem `json:"features" yaml:"features" mapstructure:"features"`
	Functions []generated.GeneratedFormatJsonElemFunctionsElem `json:"functions,omitempty" yaml:"functions,omitempty" mapstructure:"functions,omitempty"`
	Input string `json:"input" yaml:"input" mapstructure:"input"`
	Level *int `json:"level,omitempty" yaml:"level,omitempty" mapstructure:"level,omitempty"`
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	Requires []string `json:"requires,omitempty" yaml:"requires,omitempty" mapstructure:"requires,omitempty"`
	SourceTest *string `json:"source_test,omitempty" yaml:"source_test,omitempty" mapstructure:"source_test,omitempty"`
	Validation generated.GeneratedFormatJsonElemValidation `json:"validation" yaml:"validation" mapstructure:"validation"`
	Variants []generated.GeneratedFormatJsonElemVariantsElem `json:"variants" yaml:"variants" mapstructure:"variants"`
}

// convertToFlatFormat converts old TestCase to generated flat format with proper Expected structure
func (fg *FlatGenerator) convertToFlatFormat(test types.TestCase) FlatTestCase {
	// Create the proper Expected structure based on validation type
	expected := fg.createExpectedStructure(test.Validation, test.Expected)
	
	// Convert behaviors, features, variants to the generated enum types
	behaviors := fg.convertBehaviors(test.Behaviors)
	features := fg.convertFeatures(test.Features)
	variants := fg.convertVariants(test.Variants)
	functions := fg.convertFunctions(test.Functions)
	
	return FlatTestCase{
		Schema:     "ccl-test-current-flat-format",
		Name:       test.Name,
		Input:      test.Input,
		Validation: generated.GeneratedFormatJsonElemValidation(test.Validation),
		Expected:   expected,
		Functions:  functions,
		Features:   features,
		Behaviors:  behaviors,
		Variants:   variants,
		Args:       test.Args,
		SourceTest: &test.SourceTest,
		Level:      &test.Meta.Level,
	}
}

// createExpectedStructure creates the proper Expected object with Count and data fields
func (fg *FlatGenerator) createExpectedStructure(validation string, data interface{}) generated.GeneratedFormatJsonElemExpected {
	expected := generated.GeneratedFormatJsonElemExpected{}
	
	switch validation {
	case "parse", "parse_value", "filter", "compose", "expand_dotted":
		// These validations expect entries (key-value pairs)
		if entries, ok := data.([]interface{}); ok {
			expected.Count = len(entries)
			var entryList []generated.GeneratedFormatJsonElemExpectedEntriesElem
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					if key, hasKey := entryMap["key"].(string); hasKey {
						if value, hasValue := entryMap["value"].(string); hasValue {
							entryList = append(entryList, generated.GeneratedFormatJsonElemExpectedEntriesElem{
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
func (fg *FlatGenerator) convertBehaviors(behaviors []string) []generated.GeneratedFormatJsonElemBehaviorsElem {
	var result []generated.GeneratedFormatJsonElemBehaviorsElem
	for _, b := range behaviors {
		result = append(result, generated.GeneratedFormatJsonElemBehaviorsElem(b))
	}
	return result
}

func (fg *FlatGenerator) convertFeatures(features []string) []generated.GeneratedFormatJsonElemFeaturesElem {
	var result []generated.GeneratedFormatJsonElemFeaturesElem
	for _, f := range features {
		result = append(result, generated.GeneratedFormatJsonElemFeaturesElem(f))
	}
	return result
}

func (fg *FlatGenerator) convertVariants(variants []string) []generated.GeneratedFormatJsonElemVariantsElem {
	var result []generated.GeneratedFormatJsonElemVariantsElem
	for _, v := range variants {
		result = append(result, generated.GeneratedFormatJsonElemVariantsElem(v))
	}
	return result
}

func (fg *FlatGenerator) convertFunctions(functions []string) []generated.GeneratedFormatJsonElemFunctionsElem {
	var result []generated.GeneratedFormatJsonElemFunctionsElem
	for _, f := range functions {
		result = append(result, generated.GeneratedFormatJsonElemFunctionsElem(f))
	}
	return result
}

// Helper functions

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

// expectErrorFromValue checks if a validation value indicates an error expectation
func expectErrorFromValue(value interface{}) bool {
	if str, ok := value.(string); ok {
		return strings.Contains(strings.ToLower(str), "error") ||
			   strings.Contains(strings.ToLower(str), "invalid")
	}
	return false
}