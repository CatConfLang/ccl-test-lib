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
	"github.com/tylerbu/ccl-test-lib/types"
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
	// Load source test suite
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	var sourceSuite types.TestSuite
	if err := json.Unmarshal(data, &sourceSuite); err != nil {
		return fmt.Errorf("failed to parse source JSON: %w", err)
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

	// Write flat format file
	outputFile := filepath.Join(fg.OutputDir, filepath.Base(sourceFile))
	flatData, err := json.MarshalIndent(flatSuite, "", "  ")
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
		
		// Copy behaviors and variants from source
		flatTest.Behaviors = sourceTest.Behaviors
		flatTest.Variants = sourceTest.Variants
		flatTest.Conflicts = sourceTest.Conflicts

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