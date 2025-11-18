package types

import (
	"encoding/json"
	"testing"
)

func TestTestSuite_JSONMarshaling(t *testing.T) {
	suite := TestSuite{
		Suite:       "test-suite",
		Version:     "1.0.0",
		Description: "Test suite description",
		Tests: []TestCase{
			{
				Name:  "test1",
				Input: "key = value",
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(suite)
	if err != nil {
		t.Fatalf("Failed to marshal TestSuite: %v", err)
	}

	// Test unmarshaling
	var unmarshaled TestSuite
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestSuite: %v", err)
	}

	// Verify fields
	if unmarshaled.Suite != suite.Suite {
		t.Errorf("Expected suite %s, got %s", suite.Suite, unmarshaled.Suite)
	}
	if unmarshaled.Version != suite.Version {
		t.Errorf("Expected version %s, got %s", suite.Version, unmarshaled.Version)
	}
	if len(unmarshaled.Tests) != len(suite.Tests) {
		t.Errorf("Expected %d tests, got %d", len(suite.Tests), len(unmarshaled.Tests))
	}
}

func TestTestCase_SourceFormat(t *testing.T) {
	testCase := TestCase{
		Name:  "source_test",
		Input: "key = value",
		Validations: &ValidationSet{
			Parse:          []Entry{{Key: "key", Value: "value"}},
			BuildHierarchy: map[string]interface{}{"key": "value"},
		},
		Functions: []string{"parse", "build_hierarchy"},
		Features:  []string{"comments"},
		Behaviors: []string{"crlf_normalize_to_lf"},
		Variants:  []string{"proposed_behavior"},
		Meta: TestMetadata{
			Tags: []string{"function:parse"},
		},
	}

	// Test marshaling
	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("Failed to marshal TestCase: %v", err)
	}

	// Test unmarshaling
	var unmarshaled TestCase
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestCase: %v", err)
	}

	// Verify critical fields
	if unmarshaled.Name != testCase.Name {
		t.Errorf("Expected name %s, got %s", testCase.Name, unmarshaled.Name)
	}
	if unmarshaled.Validations == nil {
		t.Error("Expected validations to be preserved")
	}
	if len(unmarshaled.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(unmarshaled.Functions))
	}
}

func TestTestCase_FlatFormat(t *testing.T) {
	testCase := TestCase{
		Name:        "flat_test",
		Input:       "key = value",
		Validation:  "parse",
		Expected:    []Entry{{Key: "key", Value: "value"}},
		Args:        []string{},
		ExpectError: false,
		Functions:   []string{"parse"},
		Features:    []string{},
		Behaviors:   []string{},
		Variants:    []string{},
		Meta:        TestMetadata{},
		SourceTest:  "original_test",
	}

	// Test marshaling
	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("Failed to marshal flat TestCase: %v", err)
	}

	// Test unmarshaling
	var unmarshaled TestCase
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal flat TestCase: %v", err)
	}

	// Verify flat format fields
	if unmarshaled.Validation != testCase.Validation {
		t.Errorf("Expected validation %s, got %s", testCase.Validation, unmarshaled.Validation)
	}
	if unmarshaled.SourceTest != testCase.SourceTest {
		t.Errorf("Expected source test %s, got %s", testCase.SourceTest, unmarshaled.SourceTest)
	}
}

func TestConflictSet_JSONHandling(t *testing.T) {
	conflicts := ConflictSet{
		Functions: []string{"get_string", "get_int"},
		Behaviors: []string{"boolean_strict", "boolean_lenient"},
		Variants:  []string{"proposed_behavior"},
		Features:  []string{"comments"},
	}

	// Test marshaling
	data, err := json.Marshal(conflicts)
	if err != nil {
		t.Fatalf("Failed to marshal ConflictSet: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ConflictSet
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ConflictSet: %v", err)
	}

	// Verify all fields
	if len(unmarshaled.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(unmarshaled.Functions))
	}
	if len(unmarshaled.Behaviors) != 2 {
		t.Errorf("Expected 2 behaviors, got %d", len(unmarshaled.Behaviors))
	}
	if len(unmarshaled.Variants) != 1 {
		t.Errorf("Expected 1 variant, got %d", len(unmarshaled.Variants))
	}
	if len(unmarshaled.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(unmarshaled.Features))
	}
}

func TestValidationSet_AllFields(t *testing.T) {
	validations := ValidationSet{
		Parse:          []Entry{{Key: "key", Value: "value"}},
		ParseIndented:  "value",
		Filter:         []Entry{{Key: "key", Value: "value"}},
		Combine:        []Entry{{Key: "key", Value: "combined"}},
		ExpandDotted:   []Entry{{Key: "foo.bar", Value: "expanded"}},
		BuildHierarchy: map[string]interface{}{"foo": map[string]interface{}{"bar": "value"}},
		GetString:      "string_value",
		GetInt:         42,
		GetBool:        true,
		GetFloat:       3.14,
		GetList:        []interface{}{"a", "b", "c"},
		PrettyPrint:    "key = value\n",
		RoundTrip:      "key = value",
		Associativity:  true,
		Canonical:      "canonical_format",
	}

	// Test marshaling
	data, err := json.Marshal(validations)
	if err != nil {
		t.Fatalf("Failed to marshal ValidationSet: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ValidationSet
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ValidationSet: %v", err)
	}

	// Verify key fields exist (detailed type checking would require reflection)
	if unmarshaled.Parse == nil {
		t.Error("Parse validation was lost")
	}
	if unmarshaled.BuildHierarchy == nil {
		t.Error("BuildHierarchy validation was lost")
	}
	if unmarshaled.GetString == nil {
		t.Error("GetString validation was lost")
	}
}

func TestTestStatistics_Basic(t *testing.T) {
	stats := TestStatistics{
		TotalTests:        100,
		TotalAssertions:   150,
		CompatibleTests:   80,
		CompatibleAsserts: 120,
		ByFunction: map[string]int{
			"parse":           50,
			"build_hierarchy": 30,
			"get_string":      20,
		},
		ByFeature: map[string]int{
			"comments":  40,
			"multiline": 25,
		},
		ConflictingSets: []ConflictSummary{
			{
				ConflictType:  "behavior",
				ConflictsWith: []string{"boolean_lenient"},
				TestCount:     10,
				AssertCount:   15,
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal TestStatistics: %v", err)
	}

	// Test unmarshaling
	var unmarshaled TestStatistics
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestStatistics: %v", err)
	}

	// Verify counts
	if unmarshaled.TotalTests != 100 {
		t.Errorf("Expected 100 total tests, got %d", unmarshaled.TotalTests)
	}
	if unmarshaled.CompatibleTests != 80 {
		t.Errorf("Expected 80 compatible tests, got %d", unmarshaled.CompatibleTests)
	}
	if len(unmarshaled.ConflictingSets) != 1 {
		t.Errorf("Expected 1 conflicting set, got %d", len(unmarshaled.ConflictingSets))
	}
}

func TestConflictSummary_Fields(t *testing.T) {
	summary := ConflictSummary{
		ConflictType:  "variant",
		ConflictsWith: []string{"reference_compliant", "experimental"},
		TestCount:     25,
		AssertCount:   35,
	}

	// Test marshaling
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Failed to marshal ConflictSummary: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ConflictSummary
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ConflictSummary: %v", err)
	}

	// Verify fields
	if unmarshaled.ConflictType != "variant" {
		t.Errorf("Expected conflict type 'variant', got %s", unmarshaled.ConflictType)
	}
	if len(unmarshaled.ConflictsWith) != 2 {
		t.Errorf("Expected 2 conflicts, got %d", len(unmarshaled.ConflictsWith))
	}
	if unmarshaled.TestCount != 25 {
		t.Errorf("Expected 25 tests, got %d", unmarshaled.TestCount)
	}
	if unmarshaled.AssertCount != 35 {
		t.Errorf("Expected 35 assertions, got %d", unmarshaled.AssertCount)
	}
}

func TestEntry_BasicFunctionality(t *testing.T) {
	entry := Entry{
		Key:   "test_key",
		Value: "test_value",
	}

	// Test marshaling
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal Entry: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Entry
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Entry: %v", err)
	}

	// Verify fields
	if unmarshaled.Key != entry.Key {
		t.Errorf("Expected key %s, got %s", entry.Key, unmarshaled.Key)
	}
	if unmarshaled.Value != entry.Value {
		t.Errorf("Expected value %s, got %s", entry.Value, unmarshaled.Value)
	}
}

func TestTestMetadata_LegacySupport(t *testing.T) {
	meta := TestMetadata{
		Tags:       []string{"function:parse", "feature:comments", "level:1"},
		Conflicts:  []string{"boolean_strict"},
		Feature:    "comments",
		Difficulty: "easy",
	}

	// Test marshaling
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal TestMetadata: %v", err)
	}

	// Test unmarshaling
	var unmarshaled TestMetadata
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestMetadata: %v", err)
	}

	// Verify legacy support fields
	if len(unmarshaled.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(unmarshaled.Tags))
	}
	if unmarshaled.Feature != "comments" {
		t.Errorf("Expected feature 'comments', got %s", unmarshaled.Feature)
	}
}

// Test edge cases and error conditions

func TestTestCase_EmptySliceFields(t *testing.T) {
	testCase := TestCase{
		Name:      "empty_test",
		Input:     "key = value",
		Functions: []string{},
		Features:  []string{},
		Behaviors: []string{},
		Variants:  []string{},
	}

	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("Failed to marshal TestCase with empty slices: %v", err)
	}

	var unmarshaled TestCase
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestCase with empty slices: %v", err)
	}

	// Verify empty slices are handled correctly
	if len(unmarshaled.Functions) != 0 {
		t.Error("Functions slice should be empty")
	}
}

func TestTestCase_NilConflicts(t *testing.T) {
	testCase := TestCase{
		Name:      "no_conflicts_test",
		Input:     "key = value",
		Conflicts: nil, // Explicitly nil
	}

	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("Failed to marshal TestCase with nil conflicts: %v", err)
	}

	var unmarshaled TestCase
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TestCase with nil conflicts: %v", err)
	}

	// Verify nil conflicts are preserved (omitted in JSON)
	if unmarshaled.Conflicts != nil {
		t.Error("Conflicts should remain nil when omitted")
	}
}

func TestValidationSet_NilFields(t *testing.T) {
	validations := ValidationSet{
		Parse:          []Entry{{Key: "key", Value: "value"}},
		BuildHierarchy: nil, // Explicitly nil
		GetString:      nil, // Explicitly nil
	}

	data, err := json.Marshal(validations)
	if err != nil {
		t.Fatalf("Failed to marshal ValidationSet with nil fields: %v", err)
	}

	var unmarshaled ValidationSet
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ValidationSet with nil fields: %v", err)
	}

	// Verify nil fields are preserved
	if unmarshaled.Parse == nil {
		t.Error("Parse should not be nil")
	}
	if unmarshaled.BuildHierarchy != nil {
		t.Error("BuildHierarchy should remain nil when omitted")
	}
}
