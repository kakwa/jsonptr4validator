package jsonptr4validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

// ValidationDetail represents a single validation error with a JSON pointer and the original FieldError.
type ValidationDetail struct {
	JSONPointer string
	FieldError  validator.FieldError
}

// JSONPtrValidationError represents a global validation error
// that includes a list of JSON pointers and their corresponding FieldErrors.
type JSONPtrValidationError struct {
	ValidatorError error
	Errors         []ValidationDetail
}

// Error implements the error interface, providing a summary of all validation errors.
func (e *JSONPtrValidationError) Error() string {
	var sb strings.Builder
	for _, detail := range e.Errors {
		sb.WriteString(fmt.Sprintf("Error at '%s': %s\n", detail.JSONPointer, detail.FieldError.Error()))
	}
	return sb.String()
}

func JSONPtrValidate(validator *validator.Validate, data interface{}) JSONPtrValidationError {
}

func ResolveJSONPtr(root interface{}, fe validator.FieldError) string {
	// Split the namespace into field names
	fields := strings.Split(fe.Namespace(), ".")

	var jsonPath []string
	current := reflect.TypeOf(root)

	for _, field := range fields {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			break
		}

		// Check if the current field is an array or slice (e.g., "PublicIPs[2]")
		arrayIndex := ""
		if strings.Contains(field, "[") {
			parts := strings.Split(field, "[")
			field = parts[0]
			arrayIndex = "[" + parts[1]
		}

		// Find the field in the struct
		sf, found := current.FieldByName(field)
		if !found {
			// Fall back to the raw field name if no struct field is found
			jsonPath = append(jsonPath, field+arrayIndex)
			break
		}

		// Check for a `json` tag
		jsonTag := sf.Tag.Get("json")
		if jsonTag == "-" {
			break // Skip fields explicitly marked as `-`
		}

		// Use the JSON tag if available, otherwise the struct field name
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		if jsonFieldName == "" {
			jsonFieldName = sf.Name
		}

		jsonPath = append(jsonPath, jsonFieldName+arrayIndex)
		current = sf.Type
	}

	// Construct the JSON pointer
	return "/" + strings.Join(jsonPath, "/")
}
