# Migration Guide

This guide shows how to migrate existing CCL Go projects to use the shared `ccl-test-lib` infrastructure.

## Migration Overview

### For ccl-go Projects
**Current state**: Single-format test loading with string-based tag parsing  
**Target state**: Dual-format support with type-safe filtering

### For ccl-test-data Projects  
**Current state**: Dual-format but with duplicated test loading logic  
**Target state**: Shared infrastructure with reduced duplication

## ccl-go Migration

### Step 1: Add ccl-test-lib Dependency

```bash
go mod edit -require github.com/tylerbu/ccl-test-lib@v0.1.0
go mod download
```

### Step 2: Replace Existing Test Loading

**Before** (old ccl-go approach):
```go
// Old string-based tag parsing
func loadTests(testDir string) []TestCase {
    files := findTestFiles(testDir)
    var tests []TestCase
    for _, file := range files {
        // Custom JSON parsing
        // String-based tag filtering: "function:parse"
        tests = append(tests, parseFile(file)...)
    }
    return filterByTags(tests, supportedTags)
}
```

**After** (using ccl-test-lib):
```go
import ccl "github.com/tylerbu/ccl-test-lib"
import "github.com/tylerbu/ccl-test-lib/config"

// Declare implementation capabilities
impl := config.ImplementationConfig{
    Name: "ccl-go",
    Version: "v1.0.0",
    SupportedFunctions: []config.CCLFunction{
        config.FunctionParse,
        config.FunctionBuildHierarchy,
        config.FunctionGetString,
        // ... other functions
    },
    SupportedFeatures: []config.CCLFeature{
        config.FeatureComments,
        config.FeatureMultiline,
        // ... other features
    },
    BehaviorChoices: []config.CCLBehavior{
        config.BehaviorCRLFNormalize,
        config.BehaviorBooleanLenient,
    },
    VariantChoice: config.VariantProposed,
}

// Load compatible tests with type-safe filtering
tests, err := ccl.LoadCompatibleTests("../ccl-test-data", impl)
```

### Step 3: Update Test Runner Logic

**Before** (complex tag parsing):
```go
func runTest(test TestCase) {
    // Parse tags: "function:parse", "level:1", "feature:comments"
    tags := parseTagString(test.Tags)
    
    switch tags["function"] {
    case "parse":
        runParseTest(test)
    case "build_hierarchy":
        runBuildHierarchyTest(test)
    }
}
```

**After** (simple validation field):
```go
func runTest(test types.TestCase) {
    // Direct field access - no parsing needed
    switch test.Validation {
    case "parse":
        runParseTest(test)
    case "build_hierarchy":
        runBuildHierarchyTest(test)
    case "get_string":
        runGetStringTest(test)
    }
}
```

### Step 4: Remove Custom Test Infrastructure

Delete or deprecate:
- Custom JSON test parsing code
- String-based tag parsing logic
- Manual test filtering functions
- Custom test statistics generation

Replace with ccl-test-lib equivalents.

## ccl-test-data Migration

### Step 1: Add ccl-test-lib Dependency

```bash
go mod edit -require github.com/tylerbu/ccl-test-lib@v0.1.0
go mod download
```

### Step 2: Replace Generation Logic

**Before** (custom flat format generation):
```go
// Custom source-to-flat transformation
func generateFlatFormat(sourceDir, outputDir string) {
    // Custom logic for transforming tests/*.json to generated-tests/
    // Manual validation expansion
    // Custom metadata extraction
}
```

**After** (using ccl-test-lib):
```go
import ccl "github.com/tylerbu/ccl-test-lib"

// Simple generation with shared logic
err := ccl.GenerateFlat("tests", "generated-tests")
if err != nil {
    log.Fatal(err)
}

// Or with advanced options
gen := generator.NewFlatGenerator("tests", "generated-tests", generator.GenerateOptions{
    SkipPropertyTests: false,
    Verbose: true,
})
err := gen.GenerateAll()
```

### Step 3: Update Test Statistics

**Before** (custom statistics):
```go
func getTestStats(testDir string) TestStats {
    // Custom logic for counting tests, assertions, functions
    // Manual analysis of test coverage
}
```

**After** (using ccl-test-lib):
```go
// Universal test loader for statistics
loader := ccl.NewLoader(".", config.ImplementationConfig{
    // Minimal config for statistics gathering
})
tests, _ := loader.LoadAllTests(loader.LoadOptions{
    Format:     loader.FormatFlat,
    FilterMode: loader.FilterAll,
})
stats := loader.GetTestStatistics(tests)
```

### Step 4: Simplify Test Validation

Replace custom test validation with standard JSON schema tools:

```bash
# Instead of custom Go validation code
jv schema.json < test-file.json
```

## Benefits After Migration

### Immediate Benefits
- **Reduced code duplication** between CCL projects
- **Type-safe filtering** instead of error-prone string parsing
- **Better performance** with direct field access
- **Simplified test runners** with clear validation switching

### Long-term Benefits
- **Shared maintenance** - fixes and improvements benefit all projects
- **Consistent patterns** across CCL ecosystem
- **Easier implementation development** with clear capability declaration
- **Language-agnostic architecture** for future implementations

## Migration Checklist

### For ccl-go:
- [ ] Add ccl-test-lib dependency
- [ ] Create ImplementationConfig with your capabilities
- [ ] Replace test loading with ccl.LoadCompatibleTests()
- [ ] Update test runners to switch on test.Validation
- [ ] Remove custom test infrastructure
- [ ] Update documentation and examples
- [ ] Test compatibility with existing test suite

### For ccl-test-data:
- [ ] Add ccl-test-lib dependency  
- [ ] Replace generation logic with ccl.GenerateFlat()
- [ ] Update statistics generation with shared loader
- [ ] Replace custom validation with standard tools (jv)
- [ ] Update documentation and build scripts
- [ ] Verify generated format compatibility

## Rollback Plan

If issues arise during migration:

1. **Keep old code temporarily** - don't delete until migration is confirmed working
2. **Use feature flags** - toggle between old and new systems during transition
3. **Gradual migration** - migrate one component at a time
4. **Extensive testing** - ensure test suite compatibility before fully switching

## Support

The shared library provides:
- **Backward compatibility** - supports both old and new test formats during transition
- **Progressive adoption** - use only the features you need initially
- **Clear interfaces** - well-documented APIs for common use cases
- **Examples** - working examples for both ccl-go and ccl-test-data patterns

Migration should be straightforward with significant long-term benefits for the CCL ecosystem.