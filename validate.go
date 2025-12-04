package schema

import (
	"fmt"
	"reflect"
	"unicode"
)

// ValidateStruct checks that spec.Fields align with the Go struct fields.
// This provides early failure detection during service initialization.
//
// Validation rules:
//   - Each field name in spec.Fields must exist as an exported field in obj
//   - The Go field type must be compatible with the spec Value type:
//     - Map2Struct requires map type
//     - MapStruct requires map type
//     - ListStruct requires slice, array, or map type
//     - SingleStruct requires struct, pointer, or interface type
//
// Returns nil if spec is nil, has no fields, or all fields validate successfully.
// Returns an error describing the first validation failure found.
func ValidateStruct(obj any, spec *Struct) error {
	if spec == nil || len(spec.GetFields()) == 0 {
		return nil
	}
	if obj == nil {
		return fmt.Errorf("ValidateStruct: object is nil")
	}

	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct: object must be a struct or pointer to struct, got %v", t.Kind())
	}

	// Build map of struct fields by name
	structFields := make(map[string]reflect.StructField)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Only consider exported fields
		if len(field.Name) > 0 && unicode.IsUpper([]rune(field.Name)[0]) {
			structFields[field.Name] = field
		}
	}

	// Validate each spec field exists and type matches
	for name, value := range spec.GetFields() {
		field, ok := structFields[name]
		if !ok {
			return fmt.Errorf("ValidateStruct: field %q in spec not found in struct %s", name, t.Name())
		}

		if err := validateFieldType(field, value); err != nil {
			return fmt.Errorf("ValidateStruct: field %q: %w", name, err)
		}
	}

	return nil
}

// validateFieldType checks that the Go field type is compatible with the spec Value type.
func validateFieldType(field reflect.StructField, value *Value) error {
	if value == nil {
		return nil
	}

	kind := field.Type.Kind()
	// Dereference pointer to get underlying type
	if kind == reflect.Ptr {
		kind = field.Type.Elem().Kind()
	}

	switch {
	case value.GetMap2Struct() != nil:
		// Map2Struct expects map with composite key (e.g., map[[2]string]T)
		if kind != reflect.Map {
			return fmt.Errorf("Map2Struct requires map type, got %v", kind)
		}

	case value.GetMapStruct() != nil:
		// MapStruct expects map[string]T or similar
		if kind != reflect.Map {
			return fmt.Errorf("MapStruct requires map type, got %v", kind)
		}

	case value.GetListStruct() != nil:
		// ListStruct expects slice, array, or map (map is treated as ordered collection)
		if kind != reflect.Slice && kind != reflect.Array && kind != reflect.Map {
			return fmt.Errorf("ListStruct requires slice, array, or map type, got %v", kind)
		}

	case value.GetSingleStruct() != nil:
		// SingleStruct expects struct or pointer to struct or interface
		if kind != reflect.Struct && kind != reflect.Ptr && kind != reflect.Interface {
			return fmt.Errorf("SingleStruct requires struct, pointer, or interface type, got %v", kind)
		}
	}

	return nil
}
