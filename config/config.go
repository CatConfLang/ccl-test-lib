// Package config provides implementation capability declaration system
// for type-safe CCL test filtering and compatibility checking.
package config

// ImplementationConfig declares what an implementation supports
type ImplementationConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	// Supported CCL functions
	SupportedFunctions []CCLFunction `json:"supported_functions"`

	// Supported optional features
	SupportedFeatures []CCLFeature `json:"supported_features"`

	// Behavioral choices (mutually exclusive)
	BehaviorChoices []CCLBehavior `json:"behavior_choices"`

	// Specification variant
	VariantChoice CCLVariant `json:"variant_choice"`

	// Explicit exclusions (optional)
	UnsupportedFeatures  []CCLFeature  `json:"unsupported_features,omitempty"`
	UnsupportedFunctions []CCLFunction `json:"unsupported_functions,omitempty"`
}

// CCLFunction represents type-safe CCL function identifiers
type CCLFunction string

const (
	FunctionParse          CCLFunction = "parse"
	FunctionParseDedented  CCLFunction = "parse_dedented"
	FunctionFilter         CCLFunction = "filter"
	FunctionCombine        CCLFunction = "combine"
	FunctionExpandDotted   CCLFunction = "expand_dotted"
	FunctionBuildHierarchy CCLFunction = "build_hierarchy"
	FunctionGetString      CCLFunction = "get_string"
	FunctionGetInt         CCLFunction = "get_int"
	FunctionGetBool        CCLFunction = "get_bool"
	FunctionGetFloat       CCLFunction = "get_float"
	FunctionGetList        CCLFunction = "get_list"
	FunctionPrettyPrint    CCLFunction = "pretty_print"
)

// AllFunctions returns all valid CCL functions
func AllFunctions() []CCLFunction {
	return []CCLFunction{
		FunctionParse,
		FunctionParseDedented,
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
}

// CCLFeature represents type-safe CCL feature identifiers
type CCLFeature string

const (
	FeatureComments               CCLFeature = "comments"
	FeatureExperimentalDottedKeys CCLFeature = "experimental_dotted_keys"
	FeatureEmptyKeys              CCLFeature = "empty_keys"
	FeatureMultiline              CCLFeature = "multiline"
	FeatureUnicode                CCLFeature = "unicode"
	FeatureWhitespace             CCLFeature = "whitespace"
)

// AllFeatures returns all valid CCL features
func AllFeatures() []CCLFeature {
	return []CCLFeature{
		FeatureComments,
		FeatureExperimentalDottedKeys,
		FeatureEmptyKeys,
		FeatureMultiline,
		FeatureUnicode,
		FeatureWhitespace,
	}
}

// CCLBehavior represents type-safe CCL behavior choices
type CCLBehavior string

const (
	BehaviorCRLFNormalize   CCLBehavior = "crlf_normalize_to_lf"
	BehaviorCRLFPreserve    CCLBehavior = "crlf_preserve_literal"
	BehaviorTabsPreserve    CCLBehavior = "tabs_preserve"
	BehaviorTabsToSpaces    CCLBehavior = "tabs_to_spaces"
	BehaviorStrictSpacing   CCLBehavior = "strict_spacing"
	BehaviorLooseSpacing    CCLBehavior = "loose_spacing"
	BehaviorBooleanStrict   CCLBehavior = "boolean_strict"
	BehaviorBooleanLenient  CCLBehavior = "boolean_lenient"
	BehaviorListCoercionOn  CCLBehavior = "list_coercion_enabled"
	BehaviorListCoercionOff CCLBehavior = "list_coercion_disabled"
)

// GetBehaviorConflicts returns mutually exclusive behavior groups
func GetBehaviorConflicts() map[string][]CCLBehavior {
	return map[string][]CCLBehavior{
		"crlf_handling": {BehaviorCRLFNormalize, BehaviorCRLFPreserve},
		"tab_handling":  {BehaviorTabsPreserve, BehaviorTabsToSpaces},
		"spacing":       {BehaviorStrictSpacing, BehaviorLooseSpacing},
		"boolean":       {BehaviorBooleanStrict, BehaviorBooleanLenient},
		"list_coercion": {BehaviorListCoercionOn, BehaviorListCoercionOff},
	}
}

// CCLVariant represents type-safe CCL specification variants
type CCLVariant string

const (
	VariantProposed  CCLVariant = "proposed_behavior"
	VariantReference CCLVariant = "reference_compliant"
)

// AllVariants returns all valid CCL variants
func AllVariants() []CCLVariant {
	return []CCLVariant{
		VariantProposed,
		VariantReference,
	}
}

// IsValid validates the implementation configuration
func (c ImplementationConfig) IsValid() error {
	// Validate behavior choices don't conflict
	conflicts := GetBehaviorConflicts()
	choicesMap := make(map[CCLBehavior]bool)
	for _, choice := range c.BehaviorChoices {
		choicesMap[choice] = true
	}

	for group, behaviors := range conflicts {
		count := 0
		for _, behavior := range behaviors {
			if choicesMap[behavior] {
				count++
			}
		}
		if count > 1 {
			return &ConfigError{
				Type:    "conflicting_behaviors",
				Message: "multiple conflicting behaviors in group: " + group,
			}
		}
	}

	return nil
}

// ConfigError represents configuration validation errors
type ConfigError struct {
	Type    string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Type + ": " + e.Message
}

// HasFunction checks if implementation supports a function
func (c ImplementationConfig) HasFunction(fn CCLFunction) bool {
	for _, supported := range c.SupportedFunctions {
		if supported == fn {
			return true
		}
	}
	return false
}

// HasFeature checks if implementation supports a feature
func (c ImplementationConfig) HasFeature(feature CCLFeature) bool {
	for _, supported := range c.SupportedFeatures {
		if supported == feature {
			return true
		}
	}
	// Check if explicitly unsupported
	for _, unsupported := range c.UnsupportedFeatures {
		if unsupported == feature {
			return false
		}
	}
	return false
}

// HasBehavior checks if implementation uses a behavior
func (c ImplementationConfig) HasBehavior(behavior CCLBehavior) bool {
	for _, choice := range c.BehaviorChoices {
		if choice == behavior {
			return true
		}
	}
	return false
}

// HasVariant checks if implementation uses a variant
func (c ImplementationConfig) HasVariant(variant CCLVariant) bool {
	return c.VariantChoice == variant
}
