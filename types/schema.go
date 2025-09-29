package types

//go:generate go run ../cmd/simplify-schema/main.go ../schemas/generated-format.json ../schemas/generated-format-simple.json
//go:generate go-jsonschema -p generated -o generated/source_format.go ../schemas/source-format.json
//go:generate go-jsonschema -p generated -o generated/flat_format.go ../schemas/generated-format-simple.json

// This file contains only go:generate directives for schema-based type generation.
// All active types are defined in types.go
