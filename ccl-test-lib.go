// Package ccl_test_lib provides shared CCL test infrastructure
// for loading, filtering, and generating test suites across implementations.
package ccl_test_lib

import (
	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/generator"
	"github.com/tylerbu/ccl-test-lib/loader"
	"github.com/tylerbu/ccl-test-lib/types"
)

// Version of the ccl-test-lib package
const Version = "v0.1.0"

// Quick constructor functions for common use cases

// NewLoader creates a test loader with sensible defaults
func NewLoader(testDataPath string, cfg config.ImplementationConfig) *loader.TestLoader {
	return loader.NewTestLoader(testDataPath, cfg)
}

// NewGenerator creates a flat format generator with sensible defaults
func NewGenerator(sourceDir, outputDir string) *generator.FlatGenerator {
	return generator.NewFlatGenerator(sourceDir, outputDir, generator.GenerateOptions{
		Verbose: true,
	})
}

// LoadCompatibleTests is a convenience function for the most common use case
func LoadCompatibleTests(testDataPath string, cfg config.ImplementationConfig) ([]types.TestCase, error) {
	testLoader := NewLoader(testDataPath, cfg)
	return testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterCompatible,
	})
}

// GenerateFlat is a convenience function for generating flat format from source
func GenerateFlat(sourceDir, outputDir string) error {
	gen := NewGenerator(sourceDir, outputDir)
	return gen.GenerateAll()
}

// GetTestStats provides quick statistics for a test set
func GetTestStats(testDataPath string, cfg config.ImplementationConfig) (types.TestStatistics, error) {
	testLoader := NewLoader(testDataPath, cfg)
	tests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterAll,
	})
	if err != nil {
		return types.TestStatistics{}, err
	}
	return testLoader.GetTestStatistics(tests), nil
}