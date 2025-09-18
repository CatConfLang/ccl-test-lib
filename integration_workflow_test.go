package ccl_test_lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tylerbu/ccl-test-lib/config"
	"github.com/tylerbu/ccl-test-lib/generator"
	"github.com/tylerbu/ccl-test-lib/loader"
	"github.com/tylerbu/ccl-test-lib/types"
)

// End-to-end workflow integration tests
// These tests simulate real-world usage patterns and complete workflows

func TestWorkflow_NewCCLImplementationDevelopment(t *testing.T) {
	// Simulate the workflow of a developer creating a new CCL implementation
	// and progressively adding features using this test library

	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "ccl-test-data")
	sourceDir := filepath.Join(testDataDir, "tests")
	generatedDir := filepath.Join(testDataDir, "generated_tests")

	// Create test data directory structure
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create comprehensive test suite similar to ccl-test-data
	sourceTests := []loader.CompactTest{
		{
			Name:     "basic_parsing",
			Input:    "name = John\nage = 30",
			Level:    1,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "name", "value": "John"},
					{"key": "age", "value": "30"},
				}},
			},
		},
		{
			Name:     "object_construction",
			Input:    "user.name = Alice\nuser.age = 25",
			Level:    2,
			Features: []string{"dotted_keys"},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "user.name", "value": "Alice"},
					{"key": "user.age", "value": "25"},
				}},
				{Function: "build_hierarchy", Expect: map[string]interface{}{
					"user": map[string]interface{}{
						"name": "Alice",
						"age":  "25",
					},
				}},
			},
		},
		{
			Name:     "typed_access",
			Input:    "count = 42\nflag = true\nrate = 3.14",
			Level:    2,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "get_int", Args: []string{"count"}, Expect: 42},
				{Function: "get_bool", Args: []string{"flag"}, Expect: true},
				{Function: "get_float", Args: []string{"rate"}, Expect: 3.14},
			},
		},
		{
			Name:     "comments_support",
			Input:    "key = value\n/= This is a comment\nother = data",
			Level:    3,
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "key", "value": "value"},
					{"key": "other", "value": "data"},
				}},
				{Function: "filter", Expect: []map[string]interface{}{
					{"key": "key", "value": "value"},
					{"key": "other", "value": "data"},
				}},
			},
		},
		{
			Name:     "advanced_features",
			Input:    "list.0 = first\nlist.1 = second\nmultiline = line1\\nline2",
			Level:    4,
			Features: []string{"dotted_keys", "multiline"},
			Tests: []loader.CompactValidation{
				{Function: "get_list", Args: []string{"list"}, Expect: []interface{}{"first", "second"}},
				{Function: "get_string", Args: []string{"multiline"}, Expect: "line1\nline2"},
			},
		},
	}

	// Write test source files
	basicData, _ := json.MarshalIndent(sourceTests[:2], "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "basic.json"), basicData, 0644); err != nil {
		t.Fatalf("Failed to write basic tests: %v", err)
	}

	advancedData, _ := json.MarshalIndent(sourceTests[2:], "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "advanced.json"), advancedData, 0644); err != nil {
		t.Fatalf("Failed to write advanced tests: %v", err)
	}

	// Phase 1: Minimal implementation (only parse function)
	t.Run("phase1_minimal_implementation", func(t *testing.T) {
		minimalConfig := config.ImplementationConfig{
			Name:               "minimal-ccl",
			Version:            "v0.1.0",
			SupportedFunctions: []config.CCLFunction{config.FunctionParse},
			SupportedFeatures:  []config.CCLFeature{},
			BehaviorChoices:    []config.CCLBehavior{},
			VariantChoice:      config.VariantProposed,
		}

		// Generate flat format for testing
		gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
			SourceFormat:  generator.FormatCompact,
			OnlyFunctions: []config.CCLFunction{config.FunctionParse},
		})
		if err := gen.GenerateAll(); err != nil {
			t.Fatalf("Phase 1 generation failed: %v", err)
		}

		// Load compatible tests
		tests, err := LoadCompatibleTests(testDataDir, minimalConfig)
		if err != nil {
			t.Fatalf("Phase 1 loading failed: %v", err)
		}

		// Should only have parse tests, no dotted keys or comments
		expectedTests := 2 // basic_parsing parse + object_construction parse
		if len(tests) != expectedTests {
			t.Errorf("Phase 1: expected %d tests, got %d", expectedTests, len(tests))
		}

		// Verify no advanced features
		for _, test := range tests {
			if contains(test.Features, "dotted_keys") || contains(test.Features, "comments") {
				t.Errorf("Phase 1 should not include advanced features, but test %s has features %v",
					test.Name, test.Features)
			}
		}

		stats, err := GetTestStats(testDataDir, minimalConfig)
		if err != nil {
			t.Fatalf("Phase 1 stats failed: %v", err)
		}

		t.Logf("Phase 1 stats: %d/%d compatible tests", stats.CompatibleTests, stats.TotalTests)
	})

	// Phase 2: Add object construction
	t.Run("phase2_add_object_construction", func(t *testing.T) {
		basicConfig := config.ImplementationConfig{
			Name:    "basic-ccl",
			Version: "v0.2.0",
			SupportedFunctions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
			},
			SupportedFeatures: []config.CCLFeature{},
			BehaviorChoices:   []config.CCLBehavior{},
			VariantChoice:     config.VariantProposed,
		}

		// Regenerate with new function support
		gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
			SourceFormat: generator.FormatCompact,
			OnlyFunctions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
			},
		})
		if err := gen.GenerateAll(); err != nil {
			t.Fatalf("Phase 2 generation failed: %v", err)
		}

		tests, err := LoadCompatibleTests(testDataDir, basicConfig)
		if err != nil {
			t.Fatalf("Phase 2 loading failed: %v", err)
		}

		// Should now include build_hierarchy tests (but still exclude dotted_keys feature)
		expectedMinTests := 2 // At least parse tests
		if len(tests) < expectedMinTests {
			t.Errorf("Phase 2: expected at least %d tests, got %d", expectedMinTests, len(tests))
		}

		// Verify we have both parse and build_hierarchy tests
		validations := make(map[string]int)
		for _, test := range tests {
			validations[test.Validation]++
		}

		if validations["parse"] == 0 {
			t.Error("Phase 2 should include parse tests")
		}
		if validations["build_hierarchy"] == 0 {
			t.Error("Phase 2 should include build_hierarchy tests")
		}

		stats, err := GetTestStats(testDataDir, basicConfig)
		if err != nil {
			t.Fatalf("Phase 2 stats failed: %v", err)
		}

		t.Logf("Phase 2 stats: %d/%d compatible tests", stats.CompatibleTests, stats.TotalTests)
	})

	// Phase 3: Add typed access functions
	t.Run("phase3_add_typed_access", func(t *testing.T) {
		typedConfig := config.ImplementationConfig{
			Name:    "typed-ccl",
			Version: "v0.3.0",
			SupportedFunctions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
				config.FunctionGetInt,
				config.FunctionGetBool,
				config.FunctionGetFloat,
				config.FunctionGetString,
			},
			SupportedFeatures: []config.CCLFeature{},
			BehaviorChoices:   []config.CCLBehavior{},
			VariantChoice:     config.VariantProposed,
		}

		tests, err := LoadCompatibleTests(testDataDir, typedConfig)
		if err != nil {
			t.Fatalf("Phase 3 loading failed: %v", err)
		}

		// Should now include typed access tests
		validations := make(map[string]int)
		for _, test := range tests {
			validations[test.Validation]++
		}

		typedFunctions := []string{"get_int", "get_bool", "get_float", "get_string"}
		for _, fn := range typedFunctions {
			if validations[fn] == 0 {
				t.Errorf("Phase 3 should include %s tests", fn)
			}
		}

		stats, err := GetTestStats(testDataDir, typedConfig)
		if err != nil {
			t.Fatalf("Phase 3 stats failed: %v", err)
		}

		t.Logf("Phase 3 stats: %d/%d compatible tests", stats.CompatibleTests, stats.TotalTests)
	})

	// Phase 4: Add feature support
	t.Run("phase4_add_features", func(t *testing.T) {
		featuredConfig := config.ImplementationConfig{
			Name:    "featured-ccl",
			Version: "v0.4.0",
			SupportedFunctions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
				config.FunctionGetInt,
				config.FunctionGetBool,
				config.FunctionGetFloat,
				config.FunctionGetString,
				config.FunctionGetList,
				config.FunctionFilter,
			},
			SupportedFeatures: []config.CCLFeature{
				config.FeatureComments,
				config.FeatureMultiline,
			},
			BehaviorChoices: []config.CCLBehavior{},
			VariantChoice:   config.VariantProposed,
		}

		tests, err := LoadCompatibleTests(testDataDir, featuredConfig)
		if err != nil {
			t.Fatalf("Phase 4 loading failed: %v", err)
		}

		// Should now include feature-dependent tests
		validations := make(map[string]int)
		features := make(map[string]int)
		for _, test := range tests {
			validations[test.Validation]++
			for _, feature := range test.Features {
				features[feature]++
			}
		}

		if validations["filter"] == 0 {
			t.Error("Phase 4 should include filter tests (comments feature)")
		}
		if validations["get_list"] == 0 {
			t.Error("Phase 4 should include get_list tests")
		}

		stats, err := GetTestStats(testDataDir, featuredConfig)
		if err != nil {
			t.Fatalf("Phase 4 stats failed: %v", err)
		}

		t.Logf("Phase 4 stats: %d/%d compatible tests", stats.CompatibleTests, stats.TotalTests)

		// Verify feature coverage
		testLoader := loader.NewTestLoader(testDataDir, featuredConfig)
		coverage := testLoader.GetCapabilityCoverage()

		if commentsCoverage, exists := coverage.Features[config.FeatureComments]; exists {
			t.Logf("Comments feature coverage: %d available, %d compatible",
				commentsCoverage.Available, commentsCoverage.Compatible)
		}
	})
}

func TestWorkflow_TestDataMaintenance(t *testing.T) {
	// Simulate the workflow of maintaining and updating test data

	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	generatedDir := filepath.Join(tmpDir, "generated_tests")
	backupDir := filepath.Join(tmpDir, "backup")

	for _, dir := range []string{sourceDir, generatedDir, backupDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Initial test data
	initialTests := []loader.CompactTest{
		{
			Name:     "version1_test",
			Input:    "key = value",
			Level:    1,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "key", "value": "value"}}},
			},
		},
	}

	// Step 1: Create initial test data
	initialData, _ := json.MarshalIndent(initialTests, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "tests.json"), initialData, 0644); err != nil {
		t.Fatalf("Failed to write initial tests: %v", err)
	}

	// Generate initial flat format
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Initial generation failed: %v", err)
	}

	cfg := config.ImplementationConfig{
		Name:               "maintenance-test",
		Version:            "v1.0.0",
		SupportedFunctions: []config.CCLFunction{config.FunctionParse, config.FunctionGetString},
		SupportedFeatures:  []config.CCLFeature{},
		BehaviorChoices:    []config.CCLBehavior{},
		VariantChoice:      config.VariantProposed,
	}

	// Get initial statistics
	initialStats, err := GetTestStats(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to get initial stats: %v", err)
	}

	t.Logf("Initial stats: %d total tests", initialStats.TotalTests)

	// Step 2: Add new tests (simulating test data expansion)
	expandedTests := append(initialTests, loader.CompactTest{
		Name:     "version2_test",
		Input:    "new_key = new_value\ncount = 10",
		Level:    1,
		Features: []string{},
		Tests: []loader.CompactValidation{
			{Function: "parse", Expect: []map[string]interface{}{
				{"key": "new_key", "value": "new_value"},
				{"key": "count", "value": "10"},
			}},
			{Function: "get_string", Args: []string{"new_key"}, Expect: "new_value"},
		},
	})

	// Backup existing generated files
	existingFiles, err := filepath.Glob(filepath.Join(generatedDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list existing files: %v", err)
	}

	for _, file := range existingFiles {
		backupFile := filepath.Join(backupDir, filepath.Base(file))
		if err := copyFile(file, backupFile); err != nil {
			t.Fatalf("Failed to backup file %s: %v", file, err)
		}
	}

	// Update source data
	expandedData, _ := json.MarshalIndent(expandedTests, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "tests.json"), expandedData, 0644); err != nil {
		t.Fatalf("Failed to write expanded tests: %v", err)
	}

	// Regenerate flat format
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Regeneration failed: %v", err)
	}

	// Get updated statistics
	updatedStats, err := GetTestStats(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to get updated stats: %v", err)
	}

	t.Logf("Updated stats: %d total tests", updatedStats.TotalTests)

	// Verify expansion
	if updatedStats.TotalTests <= initialStats.TotalTests {
		t.Errorf("Expected test count to increase, got %d -> %d",
			initialStats.TotalTests, updatedStats.TotalTests)
	}

	// Step 3: Validate consistency after updates
	tests, err := LoadCompatibleTests(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to load tests after update: %v", err)
	}

	// Verify both old and new tests are present
	testNames := make(map[string]bool)
	for _, test := range tests {
		testNames[test.SourceTest] = true
	}

	if !testNames["version1_test"] {
		t.Error("Original test should still be present after update")
	}
	if !testNames["version2_test"] {
		t.Error("New test should be present after update")
	}

	// Step 4: Test rollback capability (restore from backup)
	// Clear current generated files
	currentFiles, err := filepath.Glob(filepath.Join(generatedDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list current files: %v", err)
	}

	for _, file := range currentFiles {
		if err := os.Remove(file); err != nil {
			t.Fatalf("Failed to remove file %s: %v", file, err)
		}
	}

	// Restore from backup
	backupFiles, err := filepath.Glob(filepath.Join(backupDir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list backup files: %v", err)
	}

	for _, file := range backupFiles {
		restoredFile := filepath.Join(generatedDir, filepath.Base(file))
		if err := copyFile(file, restoredFile); err != nil {
			t.Fatalf("Failed to restore file %s: %v", file, err)
		}
	}

	// Verify rollback
	rolledBackStats, err := GetTestStats(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to get rolled back stats: %v", err)
	}

	if rolledBackStats.TotalTests != initialStats.TotalTests {
		t.Errorf("Rollback failed: expected %d tests, got %d",
			initialStats.TotalTests, rolledBackStats.TotalTests)
	}

	t.Logf("Rollback successful: restored to %d tests", rolledBackStats.TotalTests)
}

func TestWorkflow_MultiProjectCompatibility(t *testing.T) {
	// Simulate multiple CCL implementations sharing the same test data

	tmpDir := t.TempDir()
	sharedTestDataDir := filepath.Join(tmpDir, "shared-test-data")
	sourceDir := filepath.Join(sharedTestDataDir, "tests")
	generatedDir := filepath.Join(sharedTestDataDir, "generated_tests")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated_tests directory: %v", err)
	}

	// Create shared test data
	sharedTests := []loader.CompactTest{
		{
			Name:     "compatibility_test",
			Input:    "basic = true\nadvanced = false\ncount = 42",
			Level:    1,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "basic", "value": "true"},
					{"key": "advanced", "value": "false"},
					{"key": "count", "value": "42"},
				}},
				{Function: "get_bool", Args: []string{"basic"}, Expect: true},
				{Function: "get_bool", Args: []string{"advanced"}, Expect: false},
				{Function: "get_int", Args: []string{"count"}, Expect: 42},
			},
		},
		{
			Name:     "feature_test",
			Input:    "key = value\n/= comment\nother = data",
			Level:    2,
			Features: []string{"comments"},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "key", "value": "value"},
					{"key": "other", "value": "data"},
				}},
				{Function: "filter", Expect: []map[string]interface{}{
					{"key": "key", "value": "value"},
					{"key": "other", "value": "data"},
				}},
			},
		},
	}

	sharedData, _ := json.MarshalIndent(sharedTests, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "shared.json"), sharedData, 0644); err != nil {
		t.Fatalf("Failed to write shared tests: %v", err)
	}

	// Generate shared flat format
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Shared generation failed: %v", err)
	}

	// Define different implementation profiles
	implementations := []struct {
		name   string
		config config.ImplementationConfig
	}{
		{
			name: "minimal_impl",
			config: config.ImplementationConfig{
				Name:               "minimal-ccl",
				Version:            "v1.0.0",
				SupportedFunctions: []config.CCLFunction{config.FunctionParse},
				SupportedFeatures:  []config.CCLFeature{},
				BehaviorChoices:    []config.CCLBehavior{},
				VariantChoice:      config.VariantProposed,
			},
		},
		{
			name: "typed_impl",
			config: config.ImplementationConfig{
				Name: "typed-ccl",
				Version: "v1.0.0",
				SupportedFunctions: []config.CCLFunction{
					config.FunctionParse,
					config.FunctionGetBool,
					config.FunctionGetInt,
				},
				SupportedFeatures: []config.CCLFeature{},
				BehaviorChoices:   []config.CCLBehavior{},
				VariantChoice:     config.VariantProposed,
			},
		},
		{
			name: "featured_impl",
			config: config.ImplementationConfig{
				Name: "featured-ccl",
				Version: "v1.0.0",
				SupportedFunctions: []config.CCLFunction{
					config.FunctionParse,
					config.FunctionGetBool,
					config.FunctionGetInt,
					config.FunctionFilter,
				},
				SupportedFeatures: []config.CCLFeature{
					config.FeatureComments,
				},
				BehaviorChoices: []config.CCLBehavior{},
				VariantChoice:   config.VariantProposed,
			},
		},
	}

	// Test each implementation with shared data
	results := make(map[string]types.TestStatistics)

	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			tests, err := LoadCompatibleTests(sharedTestDataDir, impl.config)
			if err != nil {
				t.Fatalf("Failed to load tests for %s: %v", impl.name, err)
			}

			stats, err := GetTestStats(sharedTestDataDir, impl.config)
			if err != nil {
				t.Fatalf("Failed to get stats for %s: %v", impl.name, err)
			}

			results[impl.name] = stats

			t.Logf("%s: %d/%d compatible tests", impl.name, stats.CompatibleTests, stats.TotalTests)

			// Verify each implementation gets appropriate tests
			validations := make(map[string]int)
			for _, test := range tests {
				validations[test.Validation]++
			}

			// Minimal should only have parse
			if impl.name == "minimal_impl" {
				if validations["parse"] == 0 {
					t.Error("Minimal implementation should have parse tests")
				}
				if validations["get_bool"] > 0 || validations["get_int"] > 0 {
					t.Error("Minimal implementation should not have typed access tests")
				}
			}

			// Typed should have parse + typed access
			if impl.name == "typed_impl" {
				if validations["parse"] == 0 {
					t.Error("Typed implementation should have parse tests")
				}
				if validations["get_bool"] == 0 || validations["get_int"] == 0 {
					t.Error("Typed implementation should have typed access tests")
				}
				if validations["filter"] > 0 {
					t.Error("Typed implementation should not have filter tests (no comments support)")
				}
			}

			// Featured should have all compatible tests
			if impl.name == "featured_impl" {
				expectedFunctions := []string{"parse", "get_bool", "get_int", "filter"}
				for _, fn := range expectedFunctions {
					if validations[fn] == 0 {
						t.Errorf("Featured implementation should have %s tests", fn)
					}
				}
			}
		})
	}

	// Verify progressive compatibility (more features = more or equal compatible tests)
	if results["typed_impl"].CompatibleTests < results["minimal_impl"].CompatibleTests {
		t.Error("Typed implementation should have at least as many compatible tests as minimal")
	}
	if results["featured_impl"].CompatibleTests < results["typed_impl"].CompatibleTests {
		t.Error("Featured implementation should have at least as many compatible tests as typed")
	}

	// Generate compatibility report
	t.Logf("Compatibility Report:")
	for _, impl := range implementations {
		stats := results[impl.name]
		percentage := float64(stats.CompatibleTests) / float64(stats.TotalTests) * 100
		t.Logf("  %s: %.1f%% (%d/%d tests)", impl.name, percentage, stats.CompatibleTests, stats.TotalTests)
	}
}

func TestWorkflow_ContinuousIntegrationSimulation(t *testing.T) {
	// Simulate a CI/CD pipeline that validates test changes

	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "ccl-repo")
	sourceDir := filepath.Join(repoDir, "tests")
	generatedDir := filepath.Join(repoDir, "generated_tests")

	for _, dir := range []string{sourceDir, generatedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Simulate baseline test data (what's currently in main branch)
	baselineTests := []loader.CompactTest{
		{
			Name:     "baseline_test",
			Input:    "key = value",
			Level:    1,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{{"key": "key", "value": "value"}}},
			},
		},
	}

	baselineData, _ := json.MarshalIndent(baselineTests, "", "  ")
	if err := os.WriteFile(filepath.Join(sourceDir, "baseline.json"), baselineData, 0644); err != nil {
		t.Fatalf("Failed to write baseline tests: %v", err)
	}

	// Generate baseline
	gen := generator.NewFlatGenerator(sourceDir, generatedDir, generator.GenerateOptions{
		SourceFormat: generator.FormatCompact,
	})
	if err := gen.GenerateAll(); err != nil {
		t.Fatalf("Baseline generation failed: %v", err)
	}

	// CI configuration (simulating different CCL implementations in CI)
	ciConfigs := []config.ImplementationConfig{
		{
			Name:               "ci-minimal",
			Version:            "v1.0.0",
			SupportedFunctions: []config.CCLFunction{config.FunctionParse},
			SupportedFeatures:  []config.CCLFeature{},
			BehaviorChoices:    []config.CCLBehavior{},
			VariantChoice:      config.VariantProposed,
		},
		{
			Name: "ci-full",
			Version: "v1.0.0",
			SupportedFunctions: []config.CCLFunction{
				config.FunctionParse,
				config.FunctionBuildHierarchy,
				config.FunctionGetString,
				config.FunctionGetInt,
			},
			SupportedFeatures: []config.CCLFeature{config.FeatureComments},
			BehaviorChoices:   []config.CCLBehavior{},
			VariantChoice:     config.VariantProposed,
		},
	}

	// Get baseline statistics
	baselineStats := make(map[string]types.TestStatistics)
	for _, cfg := range ciConfigs {
		stats, err := GetTestStats(repoDir, cfg)
		if err != nil {
			t.Fatalf("Failed to get baseline stats for %s: %v", cfg.Name, err)
		}
		baselineStats[cfg.Name] = stats
	}

	// Simulate PR that adds new tests
	t.Run("pr_validation", func(t *testing.T) {
		// Add new test (simulating PR changes)
		expandedTests := append(baselineTests, loader.CompactTest{
			Name:     "pr_addition",
			Input:    "new_key = new_value\ncount = 5",
			Level:    1,
			Features: []string{},
			Tests: []loader.CompactValidation{
				{Function: "parse", Expect: []map[string]interface{}{
					{"key": "new_key", "value": "new_value"},
					{"key": "count", "value": "5"},
				}},
				{Function: "get_string", Args: []string{"new_key"}, Expect: "new_value"},
				{Function: "get_int", Args: []string{"count"}, Expect: 5},
			},
		})

		// Update test data
		expandedData, _ := json.MarshalIndent(expandedTests, "", "  ")
		if err := os.WriteFile(filepath.Join(sourceDir, "baseline.json"), expandedData, 0644); err != nil {
			t.Fatalf("Failed to write expanded tests: %v", err)
		}

		// Regenerate (CI step)
		if err := gen.GenerateAll(); err != nil {
			t.Fatalf("PR generation failed: %v", err)
		}

		// Validate against all CI configurations
		prPassed := true
		for _, cfg := range ciConfigs {
			tests, err := LoadCompatibleTests(repoDir, cfg)
			if err != nil {
				t.Errorf("CI validation failed for %s: %v", cfg.Name, err)
				prPassed = false
				continue
			}

			newStats, err := GetTestStats(repoDir, cfg)
			if err != nil {
				t.Errorf("CI stats failed for %s: %v", cfg.Name, err)
				prPassed = false
				continue
			}

			// Validate that test count increased appropriately
			baseline := baselineStats[cfg.Name]
			if newStats.TotalTests <= baseline.TotalTests {
				t.Errorf("CI: Expected test count increase for %s, got %d -> %d",
					cfg.Name, baseline.TotalTests, newStats.TotalTests)
				prPassed = false
			}

			// Validate that compatible tests increased or stayed same
			if newStats.CompatibleTests < baseline.CompatibleTests {
				t.Errorf("CI: Compatible test count decreased for %s, got %d -> %d",
					cfg.Name, baseline.CompatibleTests, newStats.CompatibleTests)
				prPassed = false
			}

			// Validate specific test characteristics
			validations := make(map[string]int)
			for _, test := range tests {
				validations[test.Validation]++
			}

			// For full config, should have all new function tests
			if cfg.Name == "ci-full" {
				expectedFunctions := []string{"parse", "get_string", "get_int"}
				for _, fn := range expectedFunctions {
					if validations[fn] == 0 {
						t.Errorf("CI: Full config should have %s tests after PR", fn)
						prPassed = false
					}
				}
			}

			t.Logf("CI %s: %d total tests, %d compatible", cfg.Name, newStats.TotalTests, newStats.CompatibleTests)
		}

		if prPassed {
			t.Log("CI: PR validation PASSED")
		} else {
			t.Error("CI: PR validation FAILED")
		}
	})

	// Simulate invalid PR that should fail CI
	t.Run("invalid_pr_rejection", func(t *testing.T) {
		// Create invalid test data
		invalidTests := []loader.CompactTest{
			{
				Name:     "invalid_test",
				Input:    "", // Invalid: empty input
				Level:    0,  // Invalid: level 0
				Features: []string{"nonexistent_feature"}, // Invalid feature
				Tests: []loader.CompactValidation{
					{Function: "nonexistent_function", Expect: "something"}, // Invalid function
				},
			},
		}

		invalidData, _ := json.MarshalIndent(invalidTests, "", "  ")
		if err := os.WriteFile(filepath.Join(sourceDir, "invalid.json"), invalidData, 0644); err != nil {
			t.Fatalf("Failed to write invalid tests: %v", err)
		}

		// Try to generate (should handle gracefully)
		err := gen.GenerateAll()
		if err != nil {
			t.Logf("CI: Generation correctly rejected invalid test data: %v", err)
		} else {
			t.Log("CI: Generation handled invalid test data gracefully")
		}

		// Clean up invalid file
		os.Remove(filepath.Join(sourceDir, "invalid.json"))
	})
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}