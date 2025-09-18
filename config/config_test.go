package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImplementationConfig_JSONMarshaling(t *testing.T) {
	config := ImplementationConfig{
		Name:    "test-implementation",
		Version: "v1.0.0",
		SupportedFunctions: []CCLFunction{
			FunctionParse,
			FunctionBuildHierarchy,
			FunctionGetString,
		},
		SupportedFeatures: []CCLFeature{
			FeatureComments,
			FeatureMultiline,
		},
		BehaviorChoices: []CCLBehavior{
			BehaviorCRLFNormalize,
			BehaviorBooleanLenient,
		},
		VariantChoice: VariantProposed,
		UnsupportedFeatures: []CCLFeature{
			FeatureExperimentalDottedKeys,
		},
	}

	// Test marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal ImplementationConfig: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ImplementationConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ImplementationConfig: %v", err)
	}

	// Verify fields
	if unmarshaled.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, unmarshaled.Name)
	}
	if len(unmarshaled.SupportedFunctions) != 3 {
		t.Errorf("Expected 3 supported functions, got %d", len(unmarshaled.SupportedFunctions))
	}
	if len(unmarshaled.SupportedFeatures) != 2 {
		t.Errorf("Expected 2 supported features, got %d", len(unmarshaled.SupportedFeatures))
	}
	if len(unmarshaled.BehaviorChoices) != 2 {
		t.Errorf("Expected 2 behavior choices, got %d", len(unmarshaled.BehaviorChoices))
	}
	if unmarshaled.VariantChoice != VariantProposed {
		t.Errorf("Expected variant %s, got %s", VariantProposed, unmarshaled.VariantChoice)
	}
}

func TestAllFunctions_Completeness(t *testing.T) {
	functions := AllFunctions()

	// Verify all expected functions are present
	expectedFunctions := []CCLFunction{
		FunctionParse,
		FunctionParseValue,
		FunctionFilter,
		FunctionCombine,
		FunctionExpandDotted,
		FunctionBuildHierarchy,
		FunctionGetString,
		FunctionGetInt,
		FunctionGetBool,
		FunctionGetFloat,
		FunctionGetList,
		FunctionPrettyPrint,
	}

	if len(functions) != len(expectedFunctions) {
		t.Errorf("Expected %d functions, got %d", len(expectedFunctions), len(functions))
	}

	// Check each expected function is present
	for _, expected := range expectedFunctions {
		found := false
		for _, actual := range functions {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing function: %s", expected)
		}
	}
}

func TestAllFeatures_Completeness(t *testing.T) {
	features := AllFeatures()

	expectedFeatures := []CCLFeature{
		FeatureComments,
		FeatureExperimentalDottedKeys,
		FeatureEmptyKeys,
		FeatureMultiline,
		FeatureUnicode,
		FeatureWhitespace,
	}

	if len(features) != len(expectedFeatures) {
		t.Errorf("Expected %d features, got %d", len(expectedFeatures), len(features))
	}

	// Check each expected feature is present
	for _, expected := range expectedFeatures {
		found := false
		for _, actual := range features {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing feature: %s", expected)
		}
	}
}

func TestAllVariants_Completeness(t *testing.T) {
	variants := AllVariants()

	expectedVariants := []CCLVariant{
		VariantProposed,
		VariantReference,
	}

	if len(variants) != len(expectedVariants) {
		t.Errorf("Expected %d variants, got %d", len(expectedVariants), len(variants))
	}

	// Check each expected variant is present
	for _, expected := range expectedVariants {
		found := false
		for _, actual := range variants {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing variant: %s", expected)
		}
	}
}

func TestGetBehaviorConflicts_Structure(t *testing.T) {
	conflicts := GetBehaviorConflicts()

	// Verify expected conflict groups exist
	expectedGroups := []string{
		"crlf_handling",
		"tab_handling",
		"spacing",
		"boolean",
		"list_coercion",
	}

	for _, group := range expectedGroups {
		if _, exists := conflicts[group]; !exists {
			t.Errorf("Missing conflict group: %s", group)
		}
	}

	// Verify CRLF handling conflicts
	crlfConflicts := conflicts["crlf_handling"]
	if len(crlfConflicts) != 2 {
		t.Errorf("Expected 2 CRLF conflicts, got %d", len(crlfConflicts))
	}
	if crlfConflicts[0] != BehaviorCRLFNormalize || crlfConflicts[1] != BehaviorCRLFPreserve {
		t.Error("CRLF conflict behaviors not as expected")
	}

	// Verify boolean handling conflicts
	boolConflicts := conflicts["boolean"]
	if len(boolConflicts) != 2 {
		t.Errorf("Expected 2 boolean conflicts, got %d", len(boolConflicts))
	}
}

func TestImplementationConfig_IsValid_ValidConfig(t *testing.T) {
	config := ImplementationConfig{
		Name:    "valid-config",
		Version: "v1.0.0",
		SupportedFunctions: []CCLFunction{
			FunctionParse,
			FunctionBuildHierarchy,
		},
		SupportedFeatures: []CCLFeature{
			FeatureComments,
		},
		BehaviorChoices: []CCLBehavior{
			BehaviorCRLFNormalize,
			BehaviorBooleanLenient,
			BehaviorTabsPreserve,
		},
		VariantChoice: VariantProposed,
	}

	if err := config.IsValid(); err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}
}

func TestImplementationConfig_IsValid_ConflictingBehaviors(t *testing.T) {
	config := ImplementationConfig{
		Name:    "invalid-config",
		Version: "v1.0.0",
		BehaviorChoices: []CCLBehavior{
			BehaviorCRLFNormalize,
			BehaviorCRLFPreserve, // Conflicting with above
		},
		VariantChoice: VariantProposed,
	}

	err := config.IsValid()
	if err == nil {
		t.Error("Expected error for conflicting behaviors")
	}

	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Errorf("Expected ConfigError, got %T", err)
	}
	if configErr.Type != "conflicting_behaviors" {
		t.Errorf("Expected error type 'conflicting_behaviors', got %s", configErr.Type)
	}
	if !strings.Contains(configErr.Message, "crlf_handling") {
		t.Error("Error message should mention the conflicting group")
	}
}

func TestImplementationConfig_IsValid_MultipleBooleanConflicts(t *testing.T) {
	config := ImplementationConfig{
		Name:    "boolean-conflict-config",
		Version: "v1.0.0",
		BehaviorChoices: []CCLBehavior{
			BehaviorBooleanStrict,
			BehaviorBooleanLenient, // Conflicting with above
		},
		VariantChoice: VariantProposed,
	}

	err := config.IsValid()
	if err == nil {
		t.Error("Expected error for conflicting boolean behaviors")
	}

	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Errorf("Expected ConfigError, got %T", err)
	}
	if !strings.Contains(configErr.Message, "boolean") {
		t.Error("Error message should mention boolean conflict group")
	}
}

func TestConfigError_ErrorMessage(t *testing.T) {
	err := &ConfigError{
		Type:    "test_error",
		Message: "test message",
	}

	expected := "test_error: test message"
	if err.Error() != expected {
		t.Errorf("Expected error message %s, got %s", expected, err.Error())
	}
}

func TestImplementationConfig_HasFunction_Positive(t *testing.T) {
	config := ImplementationConfig{
		SupportedFunctions: []CCLFunction{
			FunctionParse,
			FunctionBuildHierarchy,
			FunctionGetString,
		},
	}

	// Test positive cases
	if !config.HasFunction(FunctionParse) {
		t.Error("Should have parse function")
	}
	if !config.HasFunction(FunctionBuildHierarchy) {
		t.Error("Should have build_hierarchy function")
	}
	if !config.HasFunction(FunctionGetString) {
		t.Error("Should have get_string function")
	}
}

func TestImplementationConfig_HasFunction_Negative(t *testing.T) {
	config := ImplementationConfig{
		SupportedFunctions: []CCLFunction{
			FunctionParse,
		},
	}

	// Test negative cases
	if config.HasFunction(FunctionGetInt) {
		t.Error("Should not have get_int function")
	}
	if config.HasFunction(FunctionBuildHierarchy) {
		t.Error("Should not have build_hierarchy function")
	}
}

func TestImplementationConfig_HasFeature_Positive(t *testing.T) {
	config := ImplementationConfig{
		SupportedFeatures: []CCLFeature{
			FeatureComments,
			FeatureMultiline,
		},
	}

	// Test positive cases
	if !config.HasFeature(FeatureComments) {
		t.Error("Should have comments feature")
	}
	if !config.HasFeature(FeatureMultiline) {
		t.Error("Should have multiline feature")
	}
}

func TestImplementationConfig_HasFeature_Negative(t *testing.T) {
	config := ImplementationConfig{
		SupportedFeatures: []CCLFeature{
			FeatureComments,
		},
	}

	// Test negative cases
	if config.HasFeature(FeatureUnicode) {
		t.Error("Should not have unicode feature")
	}
	if config.HasFeature(FeatureExperimentalDottedKeys) {
		t.Error("Should not have experimental_dotted_keys feature")
	}
}

func TestImplementationConfig_HasFeature_ExplicitlyUnsupported(t *testing.T) {
	config := ImplementationConfig{
		SupportedFeatures: []CCLFeature{
			FeatureComments,
		},
		UnsupportedFeatures: []CCLFeature{
			FeatureUnicode,
		},
	}

	// Feature explicitly listed as unsupported
	if config.HasFeature(FeatureUnicode) {
		t.Error("Should not have explicitly unsupported unicode feature")
	}

	// Feature not listed anywhere (default false)
	if config.HasFeature(FeatureMultiline) {
		t.Error("Should not have unlisted multiline feature")
	}
}

func TestImplementationConfig_HasBehavior_Positive(t *testing.T) {
	config := ImplementationConfig{
		BehaviorChoices: []CCLBehavior{
			BehaviorCRLFNormalize,
			BehaviorBooleanLenient,
		},
	}

	// Test positive cases
	if !config.HasBehavior(BehaviorCRLFNormalize) {
		t.Error("Should have crlf_normalize behavior")
	}
	if !config.HasBehavior(BehaviorBooleanLenient) {
		t.Error("Should have boolean_lenient behavior")
	}
}

func TestImplementationConfig_HasBehavior_Negative(t *testing.T) {
	config := ImplementationConfig{
		BehaviorChoices: []CCLBehavior{
			BehaviorCRLFNormalize,
		},
	}

	// Test negative cases
	if config.HasBehavior(BehaviorCRLFPreserve) {
		t.Error("Should not have crlf_preserve behavior")
	}
	if config.HasBehavior(BehaviorBooleanStrict) {
		t.Error("Should not have boolean_strict behavior")
	}
}

func TestImplementationConfig_HasVariant_Positive(t *testing.T) {
	config := ImplementationConfig{
		VariantChoice: VariantProposed,
	}

	if !config.HasVariant(VariantProposed) {
		t.Error("Should have proposed variant")
	}
}

func TestImplementationConfig_HasVariant_Negative(t *testing.T) {
	config := ImplementationConfig{
		VariantChoice: VariantProposed,
	}

	if config.HasVariant(VariantReference) {
		t.Error("Should not have reference variant")
	}
}

// Test edge cases and boundary conditions

func TestImplementationConfig_EmptyConfig(t *testing.T) {
	config := ImplementationConfig{}

	// Should be valid (no conflicts in empty config)
	if err := config.IsValid(); err != nil {
		t.Errorf("Empty config should be valid: %v", err)
	}

	// Should not have any capabilities
	if config.HasFunction(FunctionParse) {
		t.Error("Empty config should not have any functions")
	}
	if config.HasFeature(FeatureComments) {
		t.Error("Empty config should not have any features")
	}
	if config.HasBehavior(BehaviorCRLFNormalize) {
		t.Error("Empty config should not have any behaviors")
	}
}

func TestImplementationConfig_AllConflictGroups(t *testing.T) {
	// Test each conflict group individually
	conflicts := GetBehaviorConflicts()

	for groupName, conflictingBehaviors := range conflicts {
		if len(conflictingBehaviors) < 2 {
			continue // Skip groups with less than 2 behaviors
		}

		// Create config with all behaviors in the group (should be invalid)
		config := ImplementationConfig{
			BehaviorChoices: conflictingBehaviors,
			VariantChoice:   VariantProposed,
		}

		err := config.IsValid()
		if err == nil {
			t.Errorf("Config with all behaviors from group %s should be invalid", groupName)
		}

		configErr, ok := err.(*ConfigError)
		if !ok {
			t.Errorf("Expected ConfigError for group %s, got %T", groupName, err)
		}
		if configErr.Type != "conflicting_behaviors" {
			t.Errorf("Expected conflicting_behaviors error for group %s", groupName)
		}
	}
}

func TestCCLFunction_StringValues(t *testing.T) {
	// Verify function constants have expected string values
	testCases := []struct {
		function CCLFunction
		expected string
	}{
		{FunctionParse, "parse"},
		{FunctionParseValue, "parse_value"},
		{FunctionFilter, "filter"},
		{FunctionCombine, "combine"},
		{FunctionExpandDotted, "expand_dotted"},
		{FunctionBuildHierarchy, "build_hierarchy"},
		{FunctionGetString, "get_string"},
		{FunctionGetInt, "get_int"},
		{FunctionGetBool, "get_bool"},
		{FunctionGetFloat, "get_float"},
		{FunctionGetList, "get_list"},
		{FunctionPrettyPrint, "pretty_print"},
	}

	for _, tc := range testCases {
		if string(tc.function) != tc.expected {
			t.Errorf("Function %s should have value %s, got %s", tc.expected, tc.expected, string(tc.function))
		}
	}
}

func TestCCLFeature_StringValues(t *testing.T) {
	// Verify feature constants have expected string values
	testCases := []struct {
		feature  CCLFeature
		expected string
	}{
		{FeatureComments, "comments"},
		{FeatureExperimentalDottedKeys, "experimental_dotted_keys"},
		{FeatureEmptyKeys, "empty_keys"},
		{FeatureMultiline, "multiline"},
		{FeatureUnicode, "unicode"},
		{FeatureWhitespace, "whitespace"},
	}

	for _, tc := range testCases {
		if string(tc.feature) != tc.expected {
			t.Errorf("Feature %s should have value %s, got %s", tc.expected, tc.expected, string(tc.feature))
		}
	}
}

func TestCCLBehavior_StringValues(t *testing.T) {
	// Verify behavior constants have expected string values
	testCases := []struct {
		behavior CCLBehavior
		expected string
	}{
		{BehaviorCRLFNormalize, "crlf_normalize_to_lf"},
		{BehaviorCRLFPreserve, "crlf_preserve_literal"},
		{BehaviorTabsPreserve, "tabs_preserve"},
		{BehaviorTabsToSpaces, "tabs_to_spaces"},
		{BehaviorStrictSpacing, "strict_spacing"},
		{BehaviorLooseSpacing, "loose_spacing"},
		{BehaviorBooleanStrict, "boolean_strict"},
		{BehaviorBooleanLenient, "boolean_lenient"},
		{BehaviorListCoercionOn, "list_coercion_enabled"},
		{BehaviorListCoercionOff, "list_coercion_disabled"},
	}

	for _, tc := range testCases {
		if string(tc.behavior) != tc.expected {
			t.Errorf("Behavior %s should have value %s, got %s", tc.expected, tc.expected, string(tc.behavior))
		}
	}
}

func TestCCLVariant_StringValues(t *testing.T) {
	// Verify variant constants have expected string values
	testCases := []struct {
		variant  CCLVariant
		expected string
	}{
		{VariantProposed, "proposed_behavior"},
		{VariantReference, "reference_compliant"},
	}

	for _, tc := range testCases {
		if string(tc.variant) != tc.expected {
			t.Errorf("Variant %s should have value %s, got %s", tc.expected, tc.expected, string(tc.variant))
		}
	}
}