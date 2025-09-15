# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `ccl-test-lib`, a Go module (`github.com/tylerbu/ccl-test-lib`) that provides shared CCL (Configuration and Command Language) test infrastructure for reducing code duplication across CCL Go projects. It enables type-safe test filtering, dual-format support (source → flat), and capability-driven compatibility checking.

## Development Commands

### Core Development
```bash
# Build and test
go build ./...                    # Build all packages
go test ./...                     # Run all tests
go mod tidy                       # Clean up dependencies

# Run examples
go run examples/basic_usage.go    # Basic library usage patterns
go run examples/ccl-test-data_usage.go  # Test data project integration

# Module management
go mod download                   # Download dependencies
go get -u ./...                   # Update dependencies
```

### Testing Integration
The library is designed to work with external test data:
```bash
# Expected directory structure for integration
../ccl-test-data/tests/           # Source format (maintainable)
../ccl-test-data/generated-tests/ # Flat format (implementation-friendly)
```

## Architecture Overview

### Package Structure
- **`types/`** - Unified data structures supporting both source and flat test formats
- **`config/`** - Type-safe capability declaration system with CCL function/feature/behavior constants
- **`loader/`** - Test loading engine with filtering and compatibility checking
- **`generator/`** - Source-to-flat format transformation utilities
- **Root module** - Convenience functions wrapping common use cases

### Core Design Patterns

#### Dual-Format Architecture
- **Source Format** (`tests/*.json`): Human-maintainable with multiple validations per test case
- **Flat Format** (`generated-tests/*.json`): Implementation-friendly with single validation per test case
- **1:N Transformation**: One source test → multiple flat tests (one per validation type)

#### Type-Safe Capability System
Replace string-based tag parsing with structured configuration:
```go
// Instead of parsing "function:parse" strings
config.ImplementationConfig{
    SupportedFunctions: []config.CCLFunction{
        config.FunctionParse,
        config.FunctionBuildHierarchy,
    },
    SupportedFeatures: []config.CCLFeature{
        config.FeatureComments,
        config.FeatureMultiline,
    },
}
```

#### Progressive Implementation Support
- **Level-based filtering**: Tests organized by implementation complexity (Level 1-5)
- **Function-based filtering**: Load only tests for specific CCL functions
- **Capability-aware compatibility**: Automatic filtering based on declared capabilities

### Key Components

#### Test Loading Pipeline
1. **Format Detection**: Automatically handle both source and flat formats
2. **Capability Filtering**: Filter tests based on implementation config
3. **Conflict Resolution**: Handle mutually exclusive behaviors/variants
4. **Statistics Generation**: Comprehensive coverage analysis

#### Generation Pipeline  
1. **Source Parsing**: Load human-maintainable test suites
2. **Validation Expansion**: Transform multi-validation tests to single-validation flat tests
3. **Metadata Generation**: Extract type-safe metadata from validation types
4. **Filtering Options**: Skip property tests, limit functions, level filtering

### Integration Patterns

#### For CCL Implementations (ccl-go)
```go
// 1. Declare capabilities
impl := config.ImplementationConfig{...}

// 2. Load compatible tests (simple)
tests, err := ccl.LoadCompatibleTests("../ccl-test-data", impl)

// 3. Run tests with direct field access
switch test.Validation {
case "parse": runParseTest(test)
case "build_hierarchy": runBuildHierarchyTest(test)
}
```

#### For Test Data Projects (ccl-test-data)
```go
// 1. Generate flat format from source
err := ccl.GenerateFlat("tests", "generated-tests")

// 2. Get comprehensive statistics
stats, err := ccl.GetTestStats(".", config.ImplementationConfig{...})
```

## Key Files and Entry Points

### Public API Surface
- **`ccl-test-lib.go`** - Main convenience functions (`LoadCompatibleTests`, `GenerateFlat`, `GetTestStats`)
- **`config/config.go`** - Type-safe constants for CCL functions, features, behaviors, variants
- **`types/types.go`** - Unified data structures for both test formats
- **`loader/loader.go`** - Advanced loading with custom filtering options
- **`generator/generator.go`** - Source-to-flat transformation with generation options

### Usage Examples
- **`examples/basic_usage.go`** - Standard implementation integration patterns
- **`examples/ccl-test-data_usage.go`** - Test data project usage patterns

### Migration Support
- **`MIGRATION.md`** - Detailed migration guide for existing CCL projects
- **`README.md`** - Quick start guide and API overview

## Development Notes

### Type Safety Philosophy
This library replaces error-prone string parsing with Go type constants:
- `config.CCLFunction` instead of parsing `"function:parse"` tags
- `config.CCLFeature` instead of parsing `"feature:comments"` tags  
- Direct field access (`test.Validation`) instead of tag string manipulation

### Backward Compatibility
- Supports legacy tag-based filtering during migration periods
- Handles both array and suite JSON formats for flat tests
- Graceful degradation when test data directories are unavailable

### Performance Considerations
- Flat format optimized for direct field access (no string parsing)
- Capability filtering happens at load time, not runtime
- Batch operations for multi-file generation and loading

## Testing Strategy

The library itself has minimal test files but is designed to work with external CCL test suites. Integration testing relies on the `../ccl-test-data` directory structure. Examples demonstrate usage patterns and can serve as integration validation.