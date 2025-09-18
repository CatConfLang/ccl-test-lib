package main

import (
	"fmt"
	"log"

	ccl "github.com/tylerbu/ccl-test-lib"
	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/generator"
	"github.com/tylerbu/ccl-test-lib/loader"
)

func main() {
	fmt.Println("=== CCL Test Data Project Usage ===")

	// Example: Generate flat format from source tests
	fmt.Println("\n--- Generating Flat Format ---")

	// Basic generation
	err := ccl.GenerateFlat("../ccl-test-data/tests", "../ccl-test-data/generated-tests-new")
	if err != nil {
		log.Printf("Warning: basic generation failed: %v", err)
		fmt.Println("(This is expected if ccl-test-data source directory structure is different)")
	} else {
		fmt.Println("Successfully generated flat format with default options")
	}

	// Advanced generation with filtering
	fmt.Println("\n--- Advanced Generation ---")
	gen := generator.NewFlatGenerator("../ccl-test-data/tests", "../ccl-test-data/generated-tests-filtered", generator.GenerateOptions{
		OnlyFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
		},
		SkipPropertyTests: true,
		Verbose:           true,
	})

	err = gen.GenerateAll()
	if err != nil {
		log.Printf("Warning: advanced generation failed: %v", err)
	} else {
		fmt.Println("Successfully generated filtered flat format")
	}

	// Example: Analyze test coverage across all implementations
	fmt.Println("\n--- Test Coverage Analysis ---")

	// Mock implementation config for analysis
	mockImpl := config.ImplementationConfig{
		Name:               "analysis-mock",
		Version:            "v1.0.0",
		SupportedFunctions: config.AllFunctions(), // All functions for full coverage
		SupportedFeatures:  config.AllFeatures(),  // All features for full coverage
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorTabsPreserve,
			config.BehaviorBooleanLenient,
		},
		VariantChoice: config.VariantProposed,
	}

	// Load all tests to get comprehensive statistics
	testLoader := ccl.NewLoader("../ccl-test-data", mockImpl)
	allTests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterAll, // Load all tests, not just compatible
	})
	if err != nil {
		log.Printf("Warning: could not load all tests: %v", err)
	} else {
		stats := testLoader.GetTestStatistics(allTests)

		fmt.Printf("Total test cases: %d\n", stats.TotalTests)
		fmt.Printf("Total assertions: %d\n", stats.TotalAssertions)

		fmt.Println("\nCoverage by Level:")
		for level, count := range stats.ByLevel {
			fmt.Printf("  Level %d: %d tests\n", level, count)
		}

		fmt.Println("\nCoverage by Function:")
		for fn, count := range stats.ByFunction {
			fmt.Printf("  %s: %d tests\n", fn, count)
		}

		fmt.Println("\nCoverage by Feature:")
		for feature, count := range stats.ByFeature {
			fmt.Printf("  %s: %d tests\n", feature, count)
		}
	}

	// Example: Implementation-specific compatibility analysis
	fmt.Println("\n--- Implementation Compatibility ---")

	// Simulate a minimal implementation
	minimalImpl := config.ImplementationConfig{
		Name:    "minimal-impl",
		Version: "v0.1.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorBooleanStrict,
		},
		VariantChoice: config.VariantReference,
	}

	minimalLoader := ccl.NewLoader("../ccl-test-data", minimalImpl)
	compatibleTests, err := minimalLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterCompatible,
	})
	if err != nil {
		log.Printf("Warning: could not load compatible tests: %v", err)
	} else {
		fmt.Printf("Minimal implementation can run %d tests\n", len(compatibleTests))

		// Show coverage analysis
		coverage := minimalLoader.GetCapabilityCoverage()
		fmt.Println("\nFunction Coverage for Minimal Implementation:")
		for fn, info := range coverage.Functions {
			fmt.Printf("  %s: %d available, %d compatible\n", fn, info.Available, info.Compatible)
		}
	}

	// Example: Level-based progressive testing
	fmt.Println("\n--- Progressive Implementation Support ---")

	for level := 1; level <= 4; level++ {
		levelTests, err := testLoader.LoadTestsByLevel(level, loader.LoadOptions{
			Format:     loader.FormatFlat,
			FilterMode: loader.FilterAll,
		})
		if err != nil {
			log.Printf("Warning: could not load Level %d tests: %v", level, err)
			continue
		}
		fmt.Printf("Level %d: %d tests (cumulative)\n", level, len(levelTests))
	}

	fmt.Println("\nCCL Test Data analysis completed successfully!")
}
