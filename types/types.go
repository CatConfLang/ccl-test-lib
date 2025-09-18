// Package types defines unified data structures for CCL test suites
// supporting both source and flat formats with type-safe metadata.
package types

// TestSuite represents both source and generated test suites
type TestSuite struct {
	Suite       string     `json:"suite"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	Tests       []TestCase `json:"tests"`
}

// TestCase supports both source (multi-validation) and flat (single-validation) formats
type TestCase struct {
	Name   string `json:"name"`
	Input  string `json:"input,omitempty"`
	Input1 string `json:"input1,omitempty"` // For composition tests
	Input2 string `json:"input2,omitempty"`
	Input3 string `json:"input3,omitempty"`

	// Source format: multiple validations
	Validations *ValidationSet `json:"validations,omitempty"`

	// Flat format: single validation
	Validation  string      `json:"validation,omitempty"`
	Expected    interface{} `json:"expected,omitempty"`
	Args        []string    `json:"args,omitempty"`
	ExpectError bool        `json:"expect_error,omitempty"`

	// Type-safe metadata (replaces string tag parsing)
	Functions []string `json:"functions,omitempty"`
	Features  []string `json:"features"`
	Behaviors []string `json:"behaviors"`
	Variants  []string `json:"variants"`

	// Conflict resolution
	Conflicts *ConflictSet `json:"conflicts,omitempty"`

	Meta TestMetadata `json:"meta"`

	// Flat format traceability
	SourceTest string `json:"source_test,omitempty"`
}

// ConflictSet provides structured conflict resolution
type ConflictSet struct {
	Functions []string `json:"functions"`
	Behaviors []string `json:"behaviors"`
	Variants  []string `json:"variants"`
	Features  []string `json:"features"`
}

// ValidationSet contains all possible validations (source format)
type ValidationSet struct {
	Parse          interface{} `json:"parse,omitempty"`
	ParseValue     interface{} `json:"parse_value,omitempty"`
	Filter         interface{} `json:"filter,omitempty"`
	Combine        interface{} `json:"combine,omitempty"`
	ExpandDotted   interface{} `json:"expand_dotted,omitempty"`
	BuildHierarchy interface{} `json:"build_hierarchy,omitempty"`
	GetString      interface{} `json:"get_string,omitempty"`
	GetInt         interface{} `json:"get_int,omitempty"`
	GetBool        interface{} `json:"get_bool,omitempty"`
	GetFloat       interface{} `json:"get_float,omitempty"`
	GetList        interface{} `json:"get_list,omitempty"`
	PrettyPrint    interface{} `json:"pretty_print,omitempty"`
	RoundTrip      interface{} `json:"round_trip,omitempty"`
	Associativity  interface{} `json:"associativity,omitempty"`
	Canonical      interface{} `json:"canonical_format,omitempty"`
}

// TestMetadata contains categorization and legacy tag support
type TestMetadata struct {
	Tags       []string `json:"tags,omitempty"` // Legacy support
	Conflicts  []string `json:"conflicts,omitempty"`
	Level      int      `json:"level"`
	Feature    string   `json:"feature,omitempty"`
	Difficulty string   `json:"difficulty,omitempty"`
}

// TestStatistics provides comprehensive test suite analysis
type TestStatistics struct {
	TotalTests        int
	TotalAssertions   int
	CompatibleTests   int
	CompatibleAsserts int

	ByLevel    map[int]int
	ByFunction map[string]int
	ByFeature  map[string]int

	ConflictingSets []ConflictSummary
}

// ConflictSummary provides analysis of conflicting test sets
type ConflictSummary struct {
	ConflictType  string // "behavior", "variant", "feature"
	ConflictsWith []string
	TestCount     int
	AssertCount   int
}

// Entry represents a key-value pair from CCL parsing
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
