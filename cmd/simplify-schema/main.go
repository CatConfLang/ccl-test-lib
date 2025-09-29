package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// SchemaSimplifier removes go-jsonschema incompatible features from JSON schemas
type SchemaSimplifier struct {
	inputFile  string
	outputFile string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <input-schema.json> <output-schema.json>\n", os.Args[0])
		fmt.Println("Removes go-jsonschema incompatible features from JSON schemas")
		os.Exit(1)
	}

	simplifier := &SchemaSimplifier{
		inputFile:  os.Args[1],
		outputFile: os.Args[2],
	}

	if err := simplifier.simplify(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully simplified schema: %s -> %s\n", simplifier.inputFile, simplifier.outputFile)
}

func (s *SchemaSimplifier) simplify() error {
	// Read input schema
	data, err := os.ReadFile(s.inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Simplify schema by removing problematic features
	simplified := s.removeIncompatibleFeatures(schema)

	// Write output schema
	output, err := json.MarshalIndent(simplified, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(s.outputFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func (s *SchemaSimplifier) removeIncompatibleFeatures(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			switch key {
			case "allOf", "anyOf", "oneOf":
				// Skip conditional logic - go-jsonschema can't handle these well
				continue
			case "if", "then", "else":
				// Skip conditional validation - incompatible with go-jsonschema
				continue
			case "additionalProperties":
				// Skip for now - can cause issues in some contexts
				continue
			case "pattern":
				// Skip regex patterns - not needed for type generation
				continue
			case "description":
				// Skip descriptions to reduce output size
				continue
			case "default":
				// Skip defaults - go-jsonschema doesn't use them
				continue
			case "enum":
				// Keep enum values but in simplified form
				if arr, ok := value.([]interface{}); ok {
					result[key] = arr
				}
			case "items":
				// Recursively process array items
				result[key] = s.removeIncompatibleFeatures(value)
			case "properties":
				// Recursively process object properties
				result[key] = s.removeIncompatibleFeatures(value)
			default:
				// Keep other properties, recursively processing complex values
				result[key] = s.removeIncompatibleFeatures(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = s.removeIncompatibleFeatures(item)
		}
		return result
	default:
		// Primitive values (string, number, bool) - keep as-is
		return v
	}
}