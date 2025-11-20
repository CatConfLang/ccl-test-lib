# GitHub Copilot Instructions for ccl-test-lib

This file provides comprehensive onboarding instructions for GitHub Copilot agents working with the ccl-test-lib repository.

## Project Overview

**ccl-test-lib** is a Go module (`github.com/tylerbu/ccl-test-lib`) that provides shared CCL (Configuration and Command Language) test infrastructure. It serves as a reusable library to reduce code duplication across CCL Go projects by providing:

- **Dual-format test support**: Source format (human-maintainable) and flat format (implementation-friendly)
- **Type-safe capability system**: Structured configuration instead of string-based tag parsing
- **Intelligent test filtering**: Load only tests compatible with declared implementation capabilities
- **Test generation utilities**: Transform source format to flat format for easier consumption

### Primary Use Cases

1. **CCL Implementations (like ccl-go)**: Load compatible tests based on implementation capabilities
2. **Test Data Projects (like ccl-test-data)**: Generate flat format tests and get comprehensive statistics

## Repository Structure

```
ccl-test-lib/
├── .github/                    # GitHub configuration (CI, Copilot instructions)
├── cmd/                        # Command-line tools
│   ├── schema-sync/           # Sync schemas from ccl-test-data repository
│   └── simplify-schema/       # Create go-jsonschema compatible schemas
├── config/                     # Type-safe capability declaration system
│   ├── config.go              # CCL function/feature/behavior constants
│   └── config_test.go         # Configuration tests
├── examples/                   # Usage examples
│   ├── basic/                 # Standard implementation integration patterns
│   └── ccl-test-data/         # Test data project usage patterns
├── generator/                  # Source-to-flat format transformation
│   ├── generator.go           # Core generation logic
│   └── generator_test.go      # Generator tests
├── loader/                     # Test loading and filtering engine
│   ├── loader.go              # Loading, filtering, compatibility checking
│   └── loader_test.go         # Loader tests
├── schemas/                    # JSON Schema definitions
│   ├── source-format.json     # Multi-validation source format
│   ├── generated-format.json  # Single-validation flat format
│   └── generated-format-simple.json  # Simplified for go-jsonschema
├── types/                      # Unified data structures
│   ├── generated/             # Auto-generated Go types from schemas
│   ├── types.go               # Core type definitions
│   ├── structured.go          # Structured test representations
│   └── schema.go              # Schema handling utilities
├── ccl-test-lib.go            # Main convenience functions (public API)
├── *_test.go                  # Integration and unit tests
├── justfile                   # Development workflow automation (preferred)
├── tools.go                   # Go tool dependency management
├── go.mod / go.sum            # Go module dependencies
├── README.md                  # Quick start guide and API overview
├── CLAUDE.md                  # Detailed guidance for Claude Code AI
└── INTEGRATION_TESTS.md       # Integration test documentation
```

## Development Workflow

### Prerequisites

- **Go 1.25.1+**: The project requires Go 1.25.1 or later
- **just (optional)**: Command runner for simplified workflows (install: `brew install just` or `cargo install just`)
- **gotestsum (auto-installed)**: Enhanced test runner with better output
- **go-jsonschema (auto-installed)**: Generate Go types from JSON schemas

### Essential Commands

#### With `just` (Recommended)

```bash
# Quick development workflow
just dev                # Format, lint, generate, build, test (all-in-one)
just ci                 # Full CI pipeline validation

# Individual tasks
just build              # Generate types and build all packages
just test               # Run tests with gotestsum
just test-verbose       # Run tests with detailed output
just test-coverage      # Generate coverage reports (HTML available)
just lint               # Format and vet code
just generate           # Sync schemas and generate Go types
just run-examples       # Run both usage examples

# Dependencies and cleanup
just deps               # Install dependencies and tools
just clean              # Clean build artifacts
```

#### Without `just` (Manual Commands)

```bash
# Build and test
go generate ./...       # Generate types from schemas
go build ./...          # Build all packages
go test ./...           # Run tests
go test -v ./...        # Verbose test output
go test -cover ./...    # Test with coverage

# Schema and type generation
go run cmd/schema-sync/main.go schemas  # Sync schemas from ccl-test-data
go run cmd/simplify-schema/main.go <input> <output>  # Simplify schemas

# Code quality
go fmt ./...            # Format code
go vet ./...            # Static analysis

# Dependencies
go mod download         # Download dependencies
go mod tidy             # Clean up dependencies
go generate tools.go    # Install dev tools (gotestsum, go-jsonschema)
```

### Development Best Practices

1. **Always use `just dev` or `just ci` before committing** to ensure code quality
2. **Run tests after code changes** to catch regressions early
3. **Use `just generate` after schema changes** to regenerate types
4. **Follow existing code patterns** - examine similar code before implementing new features
5. **Keep changes minimal** - make surgical, focused modifications

## Core Concepts

### 1. Dual-Format Architecture

**Source Format** (`source_tests/*.json`)
- Human-maintainable JSON files
- Multiple validations per test case
- Example:
  ```json
  {
    "name": "basic_test",
    "input": "key = value",
    "validations": {
      "parse": [{"key": "value"}],
      "get_string": {"args": ["key"], "expected": "value"}
    }
  }
  ```

**Flat Format** (`generated_tests/*.json`)
- Implementation-friendly JSON files
- Single validation per test case (1:N transformation)
- Direct field access, no string parsing needed
- Example:
  ```json
  {
    "name": "basic_test_parse",
    "input": "key = value",
    "validation": "parse",
    "expected": [{"key": "value"}]
  }
  ```

### 2. Type-Safe Capability System

Replace error-prone string parsing with Go type constants:

```go
config.ImplementationConfig{
    Name: "my-ccl-impl",
    SupportedFunctions: []config.CCLFunction{
        config.FunctionParse,
        config.FunctionBuildHierarchy,
    },
    SupportedFeatures: []config.CCLFeature{
        config.FeatureComments,
        config.FeatureMultiline,
    },
    BehaviorChoices: []config.CCLBehavior{
        config.BehaviorCRLFNormalize,
    },
    VariantChoice: config.VariantProposed,
}
```

Key constants defined in `config/config.go`:
- `CCLFunction`: parse, build_hierarchy, get_string, get_int, etc.
- `CCLFeature`: comments, multiline, experimental_dotted_keys, etc.
- `CCLBehavior`: crlf_normalize_to_lf, boolean_lenient, etc.
- `CCLVariant`: proposed_behavior, legacy_behavior, etc.

### 3. Test Loading and Filtering

The loader automatically filters tests based on declared capabilities:

```go
// Simple approach
tests, err := ccl.LoadCompatibleTests("../ccl-test-data", implConfig)

// Advanced approach with options
loader := ccl.NewLoader("../ccl-test-data", implConfig)
tests, err := loader.LoadAllTests(loader.LoadOptions{
    Format:     loader.FormatFlat,      // or FormatSource
    FilterMode: loader.FilterCompatible, // or FilterAll
    LevelLimit: 4,                       // Skip Level 5 tests
})
```

### 4. Test Generation

Transform source format to flat format:

```go
// Simple generation
err := ccl.GenerateFlat("source_tests", "generated_tests")

// Advanced with options
gen := generator.NewFlatGenerator("source_tests", "generated_tests", 
    generator.GenerateOptions{
        SkipPropertyTests: false,
        OnlyFunctions: []config.CCLFunction{
            config.FunctionParse,
            config.FunctionBuildHierarchy,
        },
        Verbose: true,
    })
err := gen.GenerateAll()
```

## Common Tasks

### Adding a New CCL Function

1. Add constant to `config/config.go`:
   ```go
   const FunctionNewFunction CCLFunction = "new_function"
   ```

2. Add to `AllCCLFunctions` slice in `config/config.go`

3. Update validation handling in `loader/loader.go` if needed

4. Run `just dev` to ensure tests pass

### Adding a New Feature or Behavior

1. Add constant to `config/config.go`:
   ```go
   const FeatureNewFeature CCLFeature = "new_feature"
   const BehaviorNewBehavior CCLBehavior = "new_behavior"
   ```

2. Add to respective `All*` slices

3. Update compatibility checking in `loader/loader.go` if needed

4. Run `just dev` to validate

### Updating Schemas

1. Modify JSON schemas in `schemas/` directory

2. Run schema sync and generation:
   ```bash
   just generate
   # Or manually:
   go run cmd/schema-sync/main.go schemas
   go generate ./...
   ```

3. Review generated types in `types/generated/`

4. Update any code that depends on changed types

### Adding Tests

Follow existing test patterns in `*_test.go` files:

- **Unit tests**: Test individual functions in their respective `*_test.go` files
- **Integration tests**: Use `integration_*_test.go` files for cross-package workflows
- **Test data setup**: Use `t.TempDir()` for temporary test directories
- **Test naming**: Use descriptive names like `TestLoader_FilterCompatible`

Example test structure:
```go
func TestMyFeature(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    
    // Test
    result, err := MyFunction(input)
    
    // Assertions
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

## Code Patterns and Conventions

### Error Handling

```go
// Prefer detailed error messages
if err != nil {
    return nil, fmt.Errorf("failed to load tests from %s: %w", path, err)
}

// Use early returns for errors
if something == nil {
    return fmt.Errorf("something is required")
}
```

### Test Filtering Logic

```go
// Check function support
if !impl.SupportsFunction(test.Validation) {
    continue // Skip unsupported tests
}

// Check feature requirements
for _, feature := range test.Features {
    if !impl.SupportsFeature(feature) {
        skip = true
        break
    }
}
```

### Type Assertions and Conversions

```go
// Safe type assertions with ok check
if strVal, ok := value.(string); ok {
    // Use strVal
}

// Type conversions for JSON data
expected, err := json.Marshal(test.Expected)
```

## Integration Points

### External Dependencies

**ccl-test-data repository** (sibling directory expected):
- Location: `../ccl-test-data/` relative to this repo
- Contains: Test data in both source and flat formats
- Used by: Integration tests and examples
- Note: Tests gracefully handle missing test data

### Tools and Libraries

- **gotestsum**: Enhanced test runner with better output formatting
- **go-jsonschema**: Generate Go structs from JSON schemas
- **Standard library**: Extensive use of `encoding/json`, `os`, `path/filepath`, etc.

## Troubleshooting

### Common Issues

**"No such file or directory" for ccl-test-data**
- Expected: Integration tests may fail if `../ccl-test-data/` doesn't exist
- Solution: This is normal; the library handles missing test data gracefully
- Not your responsibility: Don't fix unrelated test failures

**"gotestsum: command not found"**
- Solution: Run `just deps` or `go generate tools.go` to install tools
- Alternative: Use `go test ./...` directly

**Type generation issues**
- Check schemas in `schemas/` are valid JSON
- Run `just generate` to regenerate types
- Verify `go-jsonschema` is installed

**Test failures**
- Run `just test-verbose` for detailed output
- Check if failures are pre-existing (not your responsibility)
- Focus only on failures related to your changes

## Testing Philosophy

- **Pre-existing failures**: Ignore test failures unrelated to your task
- **Minimal testing**: Only test the specific changes you make
- **Integration tests**: The repository has comprehensive integration tests in `integration_*_test.go` files
- **External test suites**: Designed to work with external CCL test data

## Key Files to Understand

### Must Read
- **`ccl-test-lib.go`**: Main public API - start here for convenience functions
- **`config/config.go`**: All type-safe constants and capability system
- **`README.md`**: Quick start guide with usage examples
- **`CLAUDE.md`**: Detailed technical guidance (complementary to this file)

### For Specific Tasks
- **`loader/loader.go`**: When working with test loading and filtering
- **`generator/generator.go`**: When working with format transformation
- **`types/types.go`**: When working with data structures
- **`justfile`**: When understanding build/test workflows

### For Reference
- **`INTEGRATION_TESTS.md`**: Comprehensive integration test documentation
- **`examples/`**: Real-world usage patterns
- **`types/generated/`**: Auto-generated types (don't edit manually)

## Working with This Repository

### Making Changes

1. **Understand the task**: Read the issue/PR description carefully
2. **Explore first**: Use `view`, examine existing code patterns
3. **Build and test**: Run `just dev` or manual commands to understand current state
4. **Plan changes**: Create minimal, surgical modifications
5. **Test frequently**: Run `just test` after each change
6. **Report progress**: Use report_progress tool to commit changes
7. **Verify**: Review committed files, ensure no unintended changes

### What NOT to Change

- **`types/generated/`**: Auto-generated from schemas, regenerate with `just generate`
- **Unrelated tests**: Don't fix pre-existing test failures
- **Schema files**: Only change with clear requirements
- **`.gitignore`**: Already properly configured

### Git and Commits

- **Don't use git commands directly**: Use report_progress tool for commits
- **Branch**: Work on feature branches (already set up for you)
- **Commit messages**: Should be concise and descriptive
- **Scope**: Keep changes focused and minimal

## Additional Resources

### Documentation Files
- `README.md`: User-facing quick start guide
- `CLAUDE.md`: AI agent guidance (Claude Code specific)
- `INTEGRATION_TESTS.md`: Integration test patterns and coverage
- `types/README.md`: Type system documentation

### Code Examples
- `examples/basic/basic_usage.go`: Standard implementation patterns
- `examples/ccl-test-data/ccl-test-data_usage.go`: Test data project patterns
- `integration_workflow_test.go`: Real-world workflow simulations

### External References
- Go modules: `go.mod` defines all dependencies
- Just command runner: `justfile` defines all automation
- CCL specification: Referenced in test data structures

## Tips for Copilot Agents

1. **Read existing code first**: Always examine similar code before implementing new features
2. **Follow established patterns**: This codebase has consistent patterns - use them
3. **Type safety**: Use constants from `config/config.go` instead of strings
4. **Error messages**: Provide context in error messages (include file paths, operation names)
5. **Test naming**: Use descriptive test names like `Test<Package>_<Function>_<Scenario>`
6. **Comments**: Match the existing comment style (minimal but clear)
7. **Minimal changes**: Make the smallest change that accomplishes the goal
8. **Integration tests**: Learn from `integration_*_test.go` files for realistic usage patterns

## Quick Reference

### Public API Functions (ccl-test-lib.go)
- `LoadCompatibleTests(testDataPath, config)`: Load tests compatible with config
- `GenerateFlat(sourceDir, outputDir)`: Generate flat format from source
- `GetTestStats(testDataPath, config)`: Get comprehensive test statistics
- `NewLoader(testDataPath, config)`: Create advanced loader
- `NewGenerator(sourceDir, outputDir)`: Create advanced generator

### Key Types (types/)
- `TestCase`: Individual test with validation and metadata
- `TestSuite`: Collection of tests
- `TestStatistics`: Coverage analysis and compatibility stats
- `ImplementationConfig`: Capability declaration

### Loader Options (loader/)
- `FormatSource` / `FormatFlat`: Choose input format
- `FilterCompatible` / `FilterAll`: Compatibility filtering
- `LevelLimit`: Skip high-complexity tests

### Generator Options (generator/)
- `SkipPropertyTests`: Exclude property-based tests
- `OnlyFunctions`: Generate only specific function tests
- `Verbose`: Enable detailed generation logging

---

**Remember**: This is a library project focused on test infrastructure. The goal is to make it easy for CCL implementations to load and run compatible tests with minimal code duplication.
