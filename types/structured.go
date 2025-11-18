package types

import (
	"github.com/tylerbu/ccl-test-lib/types/generated"
)

// Convenient type aliases for the generated structs with better naming

// SourceTest represents the structure of source test files (api_*.json)
type SourceTest = generated.SourceFormatJson

// SourceTestCase represents a single test case from source files
type SourceTestCase struct {
	Name      string                 `json:"name"`
	Input     string                 `json:"input"`
	Tests     []SourceTestValidation `json:"tests"`
	Level     *int                   `json:"level,omitempty"`
	Features  []string               `json:"features,omitempty"`
	Behaviors []string               `json:"behaviors,omitempty"`
	Variants  []string               `json:"variants,omitempty"`
	Conflicts *ConflictSpec          `json:"conflicts,omitempty"`
}

// SourceTestValidation represents a validation in source tests
type SourceTestValidation struct {
	Function string      `json:"function"`
	Expect   interface{} `json:"expect"`
	Args     []string    `json:"args,omitempty"`
	Error    bool        `json:"error,omitempty"`
}

// FlatTest represents the structure of flat test files (*-flat.json)
type FlatTest = generated.GeneratedFormatSimpleJson

// FlatTestCase represents a single flattened test case
type FlatTestCase struct {
	Name        string         `json:"name"`
	Input       string         `json:"input"`
	Validation  string         `json:"validation"`
	Expected    ExpectedResult `json:"expected"`
	Args        []string       `json:"args,omitempty"`
	Functions   []string       `json:"functions,omitempty"`
	Behaviors   []string       `json:"behaviors"`
	Variants    []string       `json:"variants"`
	Features    []string       `json:"features"`
	Conflicts   *ConflictSpec  `json:"conflicts,omitempty"`
	Requires    []string       `json:"requires,omitempty"`
	Level       *int           `json:"level,omitempty"`
	SourceTest  *string        `json:"source_test,omitempty"`
	ExpectError bool           `json:"expect_error,omitempty"`
	ErrorType   *string        `json:"error_type,omitempty"`
}

// ExpectedResult represents the expected result in flat format
type ExpectedResult struct {
	Count   int           `json:"count"`
	Entries []Entry       `json:"entries,omitempty"`
	Object  interface{}   `json:"object,omitempty"`
	Value   interface{}   `json:"value,omitempty"`
	List    []interface{} `json:"list,omitempty"`
	Error   bool          `json:"error,omitempty"`
}

// ConflictSpec represents mutually exclusive options by category
type ConflictSpec struct {
	Functions []string `json:"functions,omitempty"`
	Behaviors []string `json:"behaviors,omitempty"`
	Variants  []string `json:"variants,omitempty"`
	Features  []string `json:"features,omitempty"`
}

// Entry is defined in schema.go to avoid duplication

// Enums for better type safety

// CCLFunction represents the available CCL functions
type CCLFunction string

const (
	FunctionParse           CCLFunction = "parse"
	FunctionParseIndented   CCLFunction = "parse_indented"
	FunctionFilter          CCLFunction = "filter"
	FunctionCompose         CCLFunction = "compose"
	FunctionExpandDotted    CCLFunction = "expand_dotted"
	FunctionBuildHierarchy  CCLFunction = "build_hierarchy"
	FunctionGetString       CCLFunction = "get_string"
	FunctionGetInt          CCLFunction = "get_int"
	FunctionGetBool         CCLFunction = "get_bool"
	FunctionGetFloat        CCLFunction = "get_float"
	FunctionGetList         CCLFunction = "get_list"
	FunctionPrettyPrint     CCLFunction = "pretty_print"
	FunctionLoad            CCLFunction = "load"
	FunctionRoundTrip       CCLFunction = "round_trip"
	FunctionCanonicalFormat CCLFunction = "canonical_format"
	FunctionAssociativity   CCLFunction = "associativity"
)

// Behavior represents implementation behavior choices
type Behavior string

const (
	BehaviorBooleanStrict        Behavior = "boolean_strict"
	BehaviorBooleanLenient       Behavior = "boolean_lenient"
	BehaviorCrlfPreserveLiteral  Behavior = "crlf_preserve_literal"
	BehaviorCrlfNormalizeToLf    Behavior = "crlf_normalize_to_lf"
	BehaviorTabsPreserve         Behavior = "tabs_preserve"
	BehaviorTabsToSpaces         Behavior = "tabs_to_spaces"
	BehaviorStrictSpacing        Behavior = "strict_spacing"
	BehaviorLooseSpacing         Behavior = "loose_spacing"
	BehaviorListCoercionEnabled  Behavior = "list_coercion_enabled"
	BehaviorListCoercionDisabled Behavior = "list_coercion_disabled"
)

// Feature represents language features
type Feature string

const (
	FeatureComments               Feature = "comments"
	FeatureEmptyKeys              Feature = "empty_keys"
	FeatureExperimentalDottedKeys Feature = "experimental_dotted_keys"
	FeatureMultiline              Feature = "multiline"
	FeatureUnicode                Feature = "unicode"
	FeatureWhitespace             Feature = "whitespace"
)

// Variant represents specification variants
type Variant string

const (
	VariantProposedBehavior   Variant = "proposed_behavior"
	VariantReferenceCompliant Variant = "reference_compliant"
)
