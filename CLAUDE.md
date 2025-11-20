# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `ccl-test-lib`, a Go module (`github.com/tylerbu/ccl-test-lib`) that provides shared CCL (Configuration and Command Language) test infrastructure for reducing code duplication across CCL Go projects. It enables type-safe test filtering, dual-format support (source → flat), and capability-driven compatibility checking.

## Development Commands

### Essential Commands (use `just` for optimal workflow)
```bash
# Quick development workflow
just dev                          # Format, lint, generate, build, test
just ci                           # Full CI pipeline validation

# Core development
just build                        # Generate types and build all packages
just test                         # Run tests with gotestsum
just deps                         # Install dependencies and tools
just generate                     # Sync schemas and generate Go types

# Examples and validation
just run-examples                 # Run both usage examples
just run-basic                    # Basic library usage patterns
just run-ccl-data                 # Test data project integration
```

### Manual Commands (when just is unavailable)
```bash
# Build and test
go generate ./...                 # Generate types from schemas
go build ./...                    # Build all packages
gotestsum                         # Run tests (requires gotestsum)
go mod tidy                       # Clean up dependencies

# Schema and type generation
go run cmd/schema-sync/main.go schemas  # Sync schemas from ccl-test-data
go run cmd/simplify-schema/main.go <input> <output>  # Create go-jsonschema compatible schemas
```

### Integration Requirements
The library requires external test data and tools:
```bash
# Expected directory structure for integration
../ccl-test-data/source_tests/    # Source format (maintainable)
../ccl-test-data/generated_tests/ # Flat format (implementation-friendly)

# Required tools (installed via `just deps`)
gotestsum                         # Enhanced test runner
go-jsonschema                     # Generate Go types from JSON schemas
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
- **Source Format** (`source_tests/*.json`): Human-maintainable with multiple validations per test case
- **Flat Format** (`generated_tests/*.json`): Implementation-friendly with single validation per test case
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
err := ccl.GenerateFlat("source_tests", "generated_tests")

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

### Schema and Code Generation
- **`schemas/`** - JSON Schema definitions for test formats
  - `source-format.json` - Human-maintainable multi-validation format
  - `generated-format.json` - Implementation-friendly single-validation format
  - `generated-format-simple.json` - Simplified schema for go-jsonschema compatibility
- **`types/generated/`** - Auto-generated Go types from JSON schemas
  - `source_format.go` - Types for source test format
  - `flat_format.go` - Types for flat test format (simplified schema)
- **`cmd/schema-sync/`** - Tool to sync schemas from ccl-test-data repository
- **`cmd/simplify-schema/`** - Tool to create go-jsonschema compatible schemas

### Usage Examples
- **`examples/basic/basic_usage.go`** - Standard implementation integration patterns
- **`examples/ccl-test-data/ccl-test-data_usage.go`** - Test data project usage patterns

### Development Tools
- **`justfile`** - Complete development workflow automation
- **`tools.go`** - Go tool dependency management and installation
- **`MIGRATION.md`** - Detailed migration guide for existing CCL projects
- **`README.md`** - Quick start guide and API overview

## Development Notes

### Schema-Driven Development Workflow
1. **Schema Definition**: JSON schemas in `schemas/` define test format contracts
2. **Type Generation**: `go generate ./...` creates Go types from schemas using go-jsonschema
3. **Dual Schema Strategy**:
   - `generated-format.json` - Full schema with strict enum validation
   - `generated-format-simple.json` - Simplified schema for go-jsonschema compatibility
4. **Automatic Sync**: `cmd/schema-sync` pulls latest schemas from ccl-test-data repository

### Type Safety Philosophy
This library replaces error-prone string parsing with Go type constants:
- `config.CCLFunction` instead of parsing `"function:parse"` tags
- `config.CCLFeature` instead of parsing `"feature:comments"` tags
- Direct field access (`test.Validation`) instead of tag string manipulation
- Auto-generated types ensure compile-time validation of JSON schema compliance

### Schema Simplification for Go Generation
The `cmd/simplify-schema` tool addresses go-jsonschema limitations:
- Removes `allOf`, `anyOf`, `oneOf` conditional logic
- Strips `if`/`then`/`else` conditional validation
- Converts strict enum arrays to plain string arrays for broader compatibility
- Preserves core validation while enabling clean Go struct generation

### Backward Compatibility
- Supports legacy tag-based filtering during migration periods
- Handles both array and suite JSON formats for flat tests
- Graceful degradation when test data directories are unavailable
- Dual schema approach maintains strict validation while supporting code generation

### Performance Considerations
- Flat format optimized for direct field access (no string parsing)
- Capability filtering happens at load time, not runtime
- Batch operations for multi-file generation and loading
- Generated types eliminate runtime reflection and string parsing overhead

## Testing Strategy

The library itself has minimal test files but is designed to work with external CCL test suites. Integration testing relies on the `../ccl-test-data` directory structure. Examples demonstrate usage patterns and can serve as integration validation.

### Running Tests
- Use `just test` for enhanced test output with gotestsum
- Use `just test-coverage` for coverage analysis with HTML reports
- Use `just run-examples` to validate both usage patterns
- Integration tests verify compatibility with external ccl-test-data repository