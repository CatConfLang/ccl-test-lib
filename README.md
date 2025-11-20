# CCL Test Library

Shared Go module for CCL test loading infrastructure, providing modern dual-format test suite support with type-safe filtering.

## Overview

This library addresses code duplication between CCL Go projects by providing:
- **Unified test structures** supporting both source and flat formats
- **Type-safe filtering** based on implementation capabilities  
- **Modern test loading** with capability-driven compatibility
- **Flat format generation** for implementation-friendly test consumption

## Package Structure

```
github.com/tylerbu/ccl-test-lib/
├── types/            # Unified test data structures  
├── config/           # Implementation capability declaration
├── loader/           # Test loading and filtering
└── generator/        # Flat format generation utilities
```

## Quick Start

### 1. Declare Implementation Capabilities

```go
import "github.com/tylerbu/ccl-test-lib/config"

impl := config.ImplementationConfig{
    Name:    "my-ccl-impl",
    Version: "v1.0.0",
    SupportedFunctions: []config.CCLFunction{
        config.FunctionParse,
        config.FunctionBuildHierarchy,
        config.FunctionGetString,
        config.FunctionGetInt,
    },
    SupportedFeatures: []config.CCLFeature{
        config.FeatureComments,
        config.FeatureMultiline,
    },
    BehaviorChoices: []config.CCLBehavior{
        config.BehaviorCRLFNormalize,
        config.BehaviorBooleanLenient,
    },
    VariantChoice: config.VariantProposed,
}
```

### 2. Load Compatible Tests

```go
import ccl "github.com/tylerbu/ccl-test-lib"

// Simple approach
tests, err := ccl.LoadCompatibleTests("../ccl-test-data", impl)
if err != nil {
    log.Fatal(err)
}

// Advanced approach with options
loader := ccl.NewLoader("../ccl-test-data", impl)
tests, err := loader.LoadAllTests(loader.LoadOptions{
    Format:     loader.FormatFlat,
    FilterMode: loader.FilterCompatible,
    LevelLimit: 4, // Skip Level 5 tests
})
```

### 3. Run Tests

```go
for _, test := range tests {
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

### 4. Generate Flat Format (for test data projects)

```go
// Simple generation
err := ccl.GenerateFlat("source_tests", "generated_tests")
if err != nil {
    log.Fatal(err)
}

// Advanced generation with options
gen := generator.NewFlatGenerator("source_tests", "generated_tests", generator.GenerateOptions{
    SkipPropertyTests: false,
    OnlyFunctions: []config.CCLFunction{
        config.FunctionParse,
        config.FunctionBuildHierarchy,
    },
    Verbose: true,
})
err := gen.GenerateAll()
```

## Key Benefits

### For ccl-go
- **Modernized to dual-format** architecture (source + flat)
- **Simplified test runners** - switch on `test.Validation` instead of complex tag parsing
- **Better performance** - direct field access vs string parsing
- **Type-safe filtering** - declare capabilities, get compatible tests automatically

### For ccl-test-data  
- **Shared infrastructure** - test loading logic becomes reusable
- **Reduced duplication** - single source of truth for test structures
- **Better maintainability** - changes to test format only need updates in one place

### For Future Implementations
- **Language-agnostic patterns** - architecture translates to other languages
- **Progressive adoption** - clear path from minimal to full CCL support
- **Standardized filtering** - consistent capability declaration across implementations

## Test Formats

### Source Format (source_tests/*.json)
Human-maintainable with multiple validations per test case. Example:
```json
{
  "name": "basic_parsing",
  "input": "key = value",
  "validations": {
    "parse": [{"key": "value"}],
    "get_string": {"args": ["key"], "expected": "value"}
  }
}
```

### Flat Format (generated_tests/*.json)
Implementation-friendly with single validation per test case. Example:
```json
{
  "name": "basic_parsing_parse",
  "input": "key = value", 
  "validation": "parse",
  "expected": [{"key": "value"}]
}
```

## Type-Safe Metadata

Instead of string-based tag parsing, use structured metadata:
```go
test.Functions   []string  // ["parse", "get_string"]
test.Features    []string  // ["comments", "multiline"] 
test.Behaviors   []string  // ["crlf_normalize_to_lf"]
test.Variants    []string  // ["proposed_behavior"]
test.Conflicts   *ConflictSet  // Mutually exclusive requirements
```

## API Reference

### Core Types
- `types.TestSuite` - Test suite container
- `types.TestCase` - Individual test case (source or flat)
- `types.TestStatistics` - Comprehensive test analysis

### Configuration
- `config.ImplementationConfig` - Capability declaration
- `config.CCLFunction` - Type-safe function identifiers
- `config.CCLFeature` - Type-safe feature identifiers
- `config.CCLBehavior` - Type-safe behavior choices

### Loading
- `loader.TestLoader` - Main test loading interface
- `loader.LoadOptions` - Loading behavior control
- `LoadCompatibleTests()` - Convenience function

### Generation
- `generator.FlatGenerator` - Source to flat transformation
- `generator.GenerateOptions` - Generation behavior control
- `GenerateFlat()` - Convenience function

## Validation

Use external tools for JSON schema validation:
```bash
# Validate with jv or similar tools
jv schema.json < test-file.json
```

The library focuses on semantic validation and compatibility checking rather than schema validation.