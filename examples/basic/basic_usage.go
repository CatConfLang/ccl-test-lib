package main

import (
	"fmt"
	"log"

	ccl "github.com/CatConfLang/ccl-test-lib"
	"github.com/CatConfLang/ccl-test-lib/config"
	"github.com/CatConfLang/ccl-test-lib/loader"
)

func main() {
	// Example: ccl-go implementation capabilities
	impl := config.ImplementationConfig{
		Name:    "ccl-go",
		Version: "v1.0.0",
		SupportedFunctions: []config.CCLFunction{
			config.FunctionParse,
			config.FunctionBuildHierarchy,
			config.FunctionGetString,
			config.FunctionGetInt,
			config.FunctionGetBool,
			config.FunctionGetFloat,
			config.FunctionGetList,
		},
		SupportedFeatures: []config.CCLFeature{
			config.FeatureComments,
			config.FeatureExperimentalDottedKeys,
			config.FeatureMultiline,
			config.FeatureWhitespace,
		},
		BehaviorChoices: []config.CCLBehavior{
			config.BehaviorCRLFNormalize,
			config.BehaviorTabsPreserve,
			config.BehaviorBooleanLenient,
		},
		VariantChoice: config.VariantProposed,
	}

	// Validate configuration
	if err := impl.IsValid(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Simple usage: load compatible tests
	fmt.Println("=== Simple Usage ===")
	tests, err := ccl.LoadCompatibleTests("../ccl-test-data", impl)
	if err != nil {
		log.Printf("Warning: could not load from ccl-test-data: %v", err)
		fmt.Println("(This is expected if ccl-test-data is not available)")
	} else {
		fmt.Printf("Loaded %d compatible tests\n", len(tests))

		// Show some examples
		for i, test := range tests[:min(3, len(tests))] {
			fmt.Printf("Test %d: %s -> %s\n", i+1, test.Name, test.Validation)
		}
	}

	// Advanced usage: custom filtering
	fmt.Println("\n=== Advanced Usage ===")
	testLoader := ccl.NewLoader("../ccl-test-data", impl)

	// Load compatible tests
	basicTests, err := testLoader.LoadAllTests(loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterCompatible,
	})
	if err != nil {
		log.Printf("Warning: could not load basic tests: %v", err)
	} else {
		fmt.Printf("Loaded %d compatible tests\n", len(basicTests))
	}

	// Load only parsing tests
	parseTests, err := testLoader.LoadTestsByFunction(config.FunctionParse, loader.LoadOptions{
		Format:     loader.FormatFlat,
		FilterMode: loader.FilterCompatible,
	})
	if err != nil {
		log.Printf("Warning: could not load parse tests: %v", err)
	} else {
		fmt.Printf("Loaded %d parse tests\n", len(parseTests))
	}

	// Get statistics
	fmt.Println("\n=== Statistics ===")
	stats, err := ccl.GetTestStats("../ccl-test-data", impl)
	if err != nil {
		log.Printf("Warning: could not get stats: %v", err)
	} else {
		fmt.Printf("Total tests: %d\n", stats.TotalTests)
		fmt.Printf("Compatible: %d\n", stats.CompatibleTests)
		fmt.Printf("Functions covered: %d\n", len(stats.ByFunction))
		fmt.Printf("Features covered: %d\n", len(stats.ByFeature))
	}

	// Show capability coverage
	fmt.Println("\n=== Implementation Capabilities ===")
	fmt.Printf("Functions: %v\n", impl.SupportedFunctions)
	fmt.Printf("Features: %v\n", impl.SupportedFeatures)
	fmt.Printf("Behaviors: %v\n", impl.BehaviorChoices)
	fmt.Printf("Variant: %v\n", impl.VariantChoice)

	fmt.Println("\nExample completed successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
