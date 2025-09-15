# Generated Types Package

This package contains Go structs generated from JSON Schema using `github.com/atombender/go-jsonschema`.

## Structure

- `generated/` - Raw generated structs from go-jsonschema
  - `source_format.go` - Generated from `schemas/source-format.json`
  - `flat_format.go` - Generated from `schemas/generated-format.json`
- `structured.go` - Convenient type aliases and enums with better naming
- `schema.go` - Legacy types (DEPRECATED, kept for backward compatibility)

## Usage

### For new code:

```go
import "github.com/ccl-test-data/test-runner/internal/types"

// Use the convenient type aliases
var sourceTests types.SourceTest
var flatTests types.FlatTest

// Or use the typed enums
function := types.FunctionParse
behavior := types.BehaviorBooleanStrict
feature := types.FeatureComments
```

### For working with source test files (api_*.json):

```go
// Load source test data
var tests types.SourceTest
err := json.Unmarshal(data, &tests)

// Iterate through test cases
for _, testCase := range tests {
    fmt.Printf("Test: %s\n", testCase.Name)
    for _, validation := range testCase.Tests {
        fmt.Printf("  Function: %s\n", validation.Function)
    }
}
```

### For working with flat test files (*-flat.json):

```go
// Load flat test data
var tests types.FlatTest
err := json.Unmarshal(data, &tests)

// Iterate through test cases
for _, testCase := range tests {
    fmt.Printf("Test: %s, Function: %s\n", testCase.Name, testCase.Validation)
    fmt.Printf("  Expected count: %d\n", testCase.Expected.Count)
}
```

## Regenerating Types

To regenerate the structs after schema changes:

```bash
# From project root
go-jsonschema --output internal/types/generated/source_format.go --package generated schemas/source-format.json
go-jsonschema --output internal/types/generated/flat_format.go --package generated schemas/generated-format.json
```

## Migration from Legacy Types

Old code using the deprecated types in `schema.go` should be migrated:

| Old Type | New Type |
|----------|----------|
| `types.TestCase` | `types.SourceTestCase` |
| `types.ValidationSet` | `types.SourceTestValidation` |
| `types.TestMetadata` | Use fields in `types.SourceTestCase` directly |
| Raw validation types | `types.ExpectedResult` |

The generated types provide:
- Better type safety with enums
- Automatic JSON validation
- Clear structure matching the schemas
- Generated documentation from schema descriptions