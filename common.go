// Package schema provides unified data structures for describing hierarchical configuration schemas.
// It defines schema types (Struct, Value, ListStruct, MapStruct, Map2Struct)
// that are used across multiple packages for:
//   - Dynamic HCL/JSON unmarshaling (using ClassName for type lookup)
//   - Service orchestration (using ClassName and ServiceName for delegation)
//
// This package consolidates types from horizon/utils, determined/det, and grand/spec
// into a single shared definition.
package schema

import (
	"fmt"
	"unicode/utf8"
)

// typeSpec represents a type specification for a single struct field.
// It's a 2-element array where:
//   - [0]: class name (string)
//   - [1]: optional field specifications (map[string]any) or service name (string)
type typeSpec [2]any

// createTypeSpec creates a typeSpec from a class name and optional field specifications.
func createTypeSpec(className string, fields ...map[string]any) typeSpec {
	spec := typeSpec{className}
	if len(fields) > 0 && fields[0] != nil {
		spec[1] = fields[0]
	}
	return spec
}

// NewValue constructs a Value from a generic Go interface.
//
// This function converts Go types into the Value structure
// that describes interface field types for dynamic unmarshaling.
//
// Conversion rules:
//
//	╔════════════════════════════════════╤══════════════════════════════╗
//	║ Go type                            │ Conversion                   ║
//	╠════════════════════════════════════╪══════════════════════════════╣
//	║ string                             │ SingleStruct (class name)    ║
//	║ []string                           │ ListStruct                   ║
//	║ map[string]string                  │ MapStruct                    ║
//	║ map[[2]string]string               │ Map2Struct                   ║
//	║                                    │                              ║
//	║ [2]any                             │ SingleStruct with fields     ║
//	║ [][2]any                           │ ListStruct with fields       ║
//	║ map[string][2]any                  │ MapStruct with fields        ║
//	║ map[[2]string][2]any               │ Map2Struct with fields       ║
//	║                                    │                              ║
//	║ *Struct                            │ SingleStruct                 ║
//	║ []*Struct                          │ ListStruct            u       ║
//	║ map[string]*Struct                 │ MapStruct                    ║
//	║ map[string]*MapStruct              │ Map2Struct                   ║
//	╚════════════════════════════════════╧══════════════════════════════╝
//
// Returns an error if the input type is not supported.
func NewValue(v any) (*Value, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot create Value from nil")
	}

	switch typedValue := v.(type) {
	case string:
		return newValueFromString(typedValue)

	case [2]any:
		return newValueFromTypeSpec(typedValue)

	case []string:
		return newValueFromStringSlice(typedValue)

	case [][2]any:
		return newValueFromTypeSpecSlice(typedValue)

	case map[string]string:
		return newValueFromStringMap(typedValue)

	case map[string][2]any:
		return newValueFromTypeSpecMap(typedValue)

	case map[[2]string]string:
		return newValueFromString2DMap(typedValue)

	case map[[2]string][2]any:
		return newValueFromTypeSpec2DMap(typedValue)

	case *Struct:
		return &Value{Kind: &Value_SingleStruct{SingleStruct: typedValue}}, nil

	case []*Struct:
		listStruct := &ListStruct{ListFields: typedValue}
		return &Value{Kind: &Value_ListStruct{ListStruct: listStruct}}, nil

	case map[string]*Struct:
		mapStruct := &MapStruct{MapFields: typedValue}
		return &Value{Kind: &Value_MapStruct{MapStruct: mapStruct}}, nil

	case map[string]*MapStruct:
		map2Struct := &Map2Struct{Map2Fields: typedValue}
		return &Value{Kind: &Value_Map2Struct{Map2Struct: map2Struct}}, nil

	default:
		return nil, fmt.Errorf("unsupported type for NewValue: %T", v)
	}
}

// newValueFromString creates a Value from a simple string class name.
func newValueFromString(className string) (*Value, error) {
	spec := createTypeSpec(className)
	return newValueFromTypeSpec(spec)
}

// newValueFromTypeSpec creates a Value from a type specification.
func newValueFromTypeSpec(spec typeSpec) (*Value, error) {
	structSpec, err := newSingleStruct(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create SingleStruct: %w", err)
	}
	return &Value{Kind: &Value_SingleStruct{SingleStruct: structSpec}}, nil
}

// newValueFromStringSlice creates a Value from a slice of string class names.
func newValueFromStringSlice(classNames []string) (*Value, error) {
	specs := make([][2]any, len(classNames))
	for i, className := range classNames {
		specs[i] = createTypeSpec(className)
	}
	return newValueFromTypeSpecSlice(specs)
}

// newValueFromTypeSpecSlice creates a Value from a slice of type specifications.
func newValueFromTypeSpecSlice(specs [][2]any) (*Value, error) {
	listStruct, err := newListStruct(specs)
	if err != nil {
		return nil, fmt.Errorf("failed to create ListStruct: %w", err)
	}
	return &Value{Kind: &Value_ListStruct{ListStruct: listStruct}}, nil
}

// newValueFromStringMap creates a Value from a map of string class names.
func newValueFromStringMap(classNames map[string]string) (*Value, error) {
	specs := make(map[string][2]any, len(classNames))
	for key, className := range classNames {
		specs[key] = createTypeSpec(className)
	}
	return newValueFromTypeSpecMap(specs)
}

// newValueFromTypeSpecMap creates a Value from a map of type specifications.
func newValueFromTypeSpecMap(specs map[string][2]any) (*Value, error) {
	mapStruct, err := newMapStruct(specs)
	if err != nil {
		return nil, fmt.Errorf("failed to create MapStruct: %w", err)
	}
	return &Value{Kind: &Value_MapStruct{MapStruct: mapStruct}}, nil
}

// newValueFromString2DMap creates a Value from a 2D map of string class names.
func newValueFromString2DMap(classNames map[[2]string]string) (*Value, error) {
	specs := make(map[[2]string][2]any, len(classNames))
	for key, className := range classNames {
		specs[key] = createTypeSpec(className)
	}
	return newValueFromTypeSpec2DMap(specs)
}

// newValueFromTypeSpec2DMap creates a Value from a 2D map of type specifications.
func newValueFromTypeSpec2DMap(specs map[[2]string][2]any) (*Value, error) {
	map2Struct, err := newMap2Struct(specs)
	if err != nil {
		return nil, fmt.Errorf("failed to create Map2Struct: %w", err)
	}
	return &Value{Kind: &Value_Map2Struct{Map2Struct: map2Struct}}, nil
}

// NewStruct constructs a Struct specification for dynamic type unmarshaling.
//
// A Struct describes the runtime types of interface fields in a Go struct.
// This is used during HCL/JSON unmarshaling to know which concrete types
// to instantiate for interface fields.
//
// Parameters:
//   - className: The class name (Go struct type name), must be non-empty
//   - fieldSpecs: Optional map specifying field types (field name → type specification)
//
// The map values are converted using NewValue. See NewValue for conversion rules.
//
// Examples:
//
//	NewStruct("geo", map[string]any{
//	    "Shape": "circle",  // Shape field should be a circle
//	})
//
//	NewStruct("child", map[string]any{
//	    "Brand": map[string][2]any{
//	        "abc1": {"toy", map[string]any{"Geo": ...}},
//	    },
//	})
//
// Returns a Struct that can be passed to UnmarshalSpec functions.
// Returns an error if className is empty or field specifications are invalid.
func NewStruct(className string, fieldSpecs ...map[string]any) (*Struct, error) {
	if className == "" {
		return nil, fmt.Errorf("className cannot be empty")
	}

	structSpec := &Struct{ClassName: className}

	// No field specifications provided
	if len(fieldSpecs) == 0 || fieldSpecs[0] == nil {
		return structSpec, nil
	}

	// Convert field specifications to Values
	structSpec.Fields = make(map[string]*Value, len(fieldSpecs[0]))
	for fieldName, fieldSpec := range fieldSpecs[0] {
		if !utf8.ValidString(fieldName) {
			return nil, fmt.Errorf("field name contains invalid UTF-8: %q", fieldName)
		}

		fieldValue, err := NewValue(fieldSpec)
		if err != nil {
			return nil, fmt.Errorf("invalid specification for field %q: %w", fieldName, err)
		}

		structSpec.Fields[fieldName] = fieldValue
	}

	return structSpec, nil
}

// NewServiceStruct constructs a Struct for service orchestration.
//
// This creates a Struct with ClassName and optional ServiceName for
// service delegation in the microservice framework.
//
// Parameters:
//   - className: The class/object type identifier
//   - v: Either a service name (string), field specifications (map[string]any), or a Struct directly (*Struct)
//
// Examples:
//
//	NewServiceStruct("provider", "providerService")  // Delegate to providerService
//	NewServiceStruct("config", map[string]any{...}) // With nested field specs
//	NewServiceStruct("wrapper", existingStruct)     // Use existing Struct
func NewServiceStruct(className string, v any) (*Struct, error) {
	x := &Struct{ClassName: className}
	if v == nil {
		return nil, fmt.Errorf("nil value for service struct")
	}
	switch t := v.(type) {
	case string:
		x.ServiceName = t
	case map[string]any:
		x.Fields = make(map[string]*Value, len(t))
		for key, val := range t {
			if !utf8.ValidString(key) {
				return nil, fmt.Errorf("invalid UTF-8 in key: %q", key)
			}
			var err error
			x.Fields[key], err = NewServiceValue(val)
			if err != nil {
				return nil, err
			}
		}
	case *Struct:
		// Use the provided Struct's fields and service name
		x.Fields = t.Fields
		x.ServiceName = t.ServiceName
	default:
		return nil, fmt.Errorf("invalid type for service struct: %T", v)
	}
	if err := validateServiceEndStruct(x); err != nil {
		return nil, err
	}
	return x, nil
}

func validateServiceEndStruct(s *Struct) error {
	return validateServiceEndStructWithSeen(s, map[*Struct]struct{}{}, "")
}

func validateServiceEndStructWithSeen(s *Struct, seen map[*Struct]struct{}, path string) error {
	if s == nil {
		return nil
	}
	if path == "" {
		if s.ClassName != "" {
			path = s.ClassName
		} else {
			path = "<root>"
		}
	}
	if _, ok := seen[s]; ok {
		return nil
	}
	seen[s] = struct{}{}
	if s.ServiceName != "" && len(s.Fields) > 0 {
		return fmt.Errorf("service name must be on leaf struct at %s", path)
	}
	for name, v := range s.Fields {
		childPath := path + "." + name
		if err := validateServiceEndValueWithSeen(v, seen, childPath); err != nil {
			return err
		}
	}
	return nil
}

func validateServiceEndValue(v *Value) error {
	return validateServiceEndValueWithSeen(v, map[*Struct]struct{}{}, "")
}

func validateServiceEndValueWithSeen(v *Value, seen map[*Struct]struct{}, path string) error {
	if v == nil {
		return nil
	}
	if path == "" {
		path = "<root>"
	}
	switch k := v.Kind.(type) {
	case *Value_SingleStruct:
		return validateServiceEndStructWithSeen(k.SingleStruct, seen, path)
	case *Value_ListStruct:
		if k.ListStruct == nil {
			return nil
		}
		for i, s := range k.ListStruct.ListFields {
			if err := validateServiceEndStructWithSeen(s, seen, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case *Value_MapStruct:
		if k.MapStruct == nil {
			return nil
		}
		for key, s := range k.MapStruct.MapFields {
			if err := validateServiceEndStructWithSeen(s, seen, fmt.Sprintf("%s[%q]", path, key)); err != nil {
				return err
			}
		}
	case *Value_Map2Struct:
		if k.Map2Struct == nil {
			return nil
		}
		for outerKey, ms := range k.Map2Struct.Map2Fields {
			if ms == nil {
				continue
			}
			for innerKey, s := range ms.MapFields {
				if err := validateServiceEndStructWithSeen(s, seen, fmt.Sprintf("%s[%q][%q]", path, outerKey, innerKey)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// newSingleStruct creates a Struct from a type specification.
// The spec is a [2]any where:
//   - [0]: class name (string)
//   - [1]: optional field specifications (map[string]any) or service name (string)
func newSingleStruct(spec typeSpec) (*Struct, error) {
	className, ok := spec[0].(string)
	if !ok {
		return nil, fmt.Errorf("class name must be a string, got %T", spec[0])
	}

	if spec[1] == nil {
		return NewStruct(className)
	}

	// Check if it's a service name (string) or field specs (map)
	switch v := spec[1].(type) {
	case string:
		// Service name - create struct with service
		s := &Struct{ClassName: className, ServiceName: v}
		return s, nil
	case map[string]any:
		return NewStruct(className, v)
	default:
		return nil, fmt.Errorf("field specifications must be map[string]any or string, got %T", spec[1])
	}
}

// newListStruct creates a ListStruct from a slice of type specifications.
func newListStruct(specs [][2]any) (*ListStruct, error) {
	structs := make([]*Struct, len(specs))

	for i, spec := range specs {
		structSpec, err := newSingleStruct(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid specification at index %d: %w", i, err)
		}
		structs[i] = structSpec
	}

	return &ListStruct{ListFields: structs}, nil
}

// newMapStruct creates a MapStruct from a map of type specifications.
func newMapStruct(specs map[string][2]any) (*MapStruct, error) {
	structs := make(map[string]*Struct, len(specs))

	for key, spec := range specs {
		structSpec, err := newSingleStruct(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid specification for key %q: %w", key, err)
		}
		structs[key] = structSpec
	}

	return &MapStruct{MapFields: structs}, nil
}

// newMap2Struct creates a Map2Struct from a 2D map of type specifications.
// This handles nested map structures where the key is a 2-element string array.
func newMap2Struct(specs map[[2]string][2]any) (*Map2Struct, error) {
	// Group specifications by the first key dimension
	groupedSpecs := make(map[string]map[string][2]any)

	for key, spec := range specs {
		firstKey := key[0]
		secondKey := key[1]

		if groupedSpecs[firstKey] == nil {
			groupedSpecs[firstKey] = make(map[string][2]any)
		}
		groupedSpecs[firstKey][secondKey] = spec
	}

	// Convert grouped specifications to MapStructs
	map2Fields := make(map[string]*MapStruct, len(groupedSpecs))

	for firstKey, secondLevelSpecs := range groupedSpecs {
		mapStruct, err := newMapStruct(secondLevelSpecs)
		if err != nil {
			return nil, fmt.Errorf("invalid specification for first-level key %q: %w", firstKey, err)
		}
		map2Fields[firstKey] = mapStruct
	}

	return &Map2Struct{Map2Fields: map2Fields}, nil
}

// --- Service Orchestration Helpers (for grand/spec compatibility) ---

// NewServiceValue constructs a Value for service orchestration from generic Go interface v.
//
// This is similar to NewValue but uses []string format for end-node structs:
//
//	╔═══════════════════════════╤══════════════════╤═════════════╤════════════════════════════╗
//	║ Go type                   │ Conversion       │ First       │ Second                     ║
//	╠═══════════════════════════╪══════════════════╪═════════════╪════════════════════════════╣
//	║ []string                  │ end SingleStruct │ class name  │ service name               ║
//	║ [][]string                │ end ListStruct   │ class name  │ service name               ║
//	║ map[string][]string       │ end MapStruct    │ class name  │ service name               ║
//	║ map[[2]string][]string    │ end Map2Struct   │ class name  │ service name               ║
//	║ [2]any                    │ SingleStruct     │ class name  │ service/fields/*Struct     ║
//	║ [][2]any                  │ ListStruct       │ class name  │ service/fields/*Struct     ║
//	║ map[string][2]any         │ MapStruct        │ class name  │ service/fields/*Struct     ║
//	║ map[[2]string][2]any      │ Map2Struct       │ class name  │ service/fields/*Struct     ║
//	║ *Struct                   │ SingleStruct     │ -           │ -                          ║
//	║ []*Struct                 │ ListStruct       │ -           │ -                          ║
//	║ map[string]*Struct        │ MapStruct        │ -           │ -                          ║
//	║ map[string]*MapStruct     │ Map2Struct       │ -           │ -                          ║
//	╚═══════════════════════════╧══════════════════╧═════════════╧════════════════════════════╝
func NewServiceValue(v any) (*Value, error) {
	switch t := v.(type) {
	case []string:
		v2, err := newEndStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_SingleStruct{SingleStruct: v2}})
	case [][]string:
		v2, err := newEndListStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_ListStruct{ListStruct: v2}})
	case map[string][]string:
		v2, err := newEndMapStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_MapStruct{MapStruct: v2}})
	case map[[2]string][]string:
		v2, err := newEndMap2Struct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_Map2Struct{Map2Struct: v2}})
	case [2]any:
		v2, err := newServiceSingleStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_SingleStruct{SingleStruct: v2}})
	case [][2]any:
		v2, err := newServiceListStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_ListStruct{ListStruct: v2}})
	case map[string][2]any:
		v2, err := newServiceMapStruct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_MapStruct{MapStruct: v2}})
	case map[[2]string][2]any:
		v2, err := newServiceMap2Struct(t)
		if err != nil {
			return nil, err
		}
		return finalizeServiceValue(&Value{Kind: &Value_Map2Struct{Map2Struct: v2}})
	case *Struct:
		return finalizeServiceValue(&Value{Kind: &Value_SingleStruct{SingleStruct: t}})
	case []*Struct:
		return finalizeServiceValue(&Value{Kind: &Value_ListStruct{ListStruct: &ListStruct{ListFields: t}}})
	case map[string]*Struct:
		return finalizeServiceValue(&Value{Kind: &Value_MapStruct{MapStruct: &MapStruct{MapFields: t}}})
	case map[string]*MapStruct:
		return finalizeServiceValue(&Value{Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{Map2Fields: t}}})
	default:
		return nil, fmt.Errorf("unsupported type for NewServiceValue: %T", v)
	}
}

func finalizeServiceValue(v *Value) (*Value, error) {
	if err := validateServiceEndValue(v); err != nil {
		return nil, err
	}
	return v, nil
}

// newServiceSingleStruct creates a Struct from a type specification for service orchestration.
func newServiceSingleStruct(spec typeSpec) (*Struct, error) {
	className, ok := spec[0].(string)
	if !ok {
		return nil, fmt.Errorf("class name must be a string, got %T", spec[0])
	}
	if spec[1] == nil {
		return &Struct{ClassName: className}, nil
	}
	switch v := spec[1].(type) {
	case string:
		return &Struct{ClassName: className, ServiceName: v}, nil
	case map[string]any:
		return NewServiceStruct(className, v)
	case *Struct:
		return &Struct{ClassName: className, Fields: v.Fields, ServiceName: v.ServiceName}, nil
	default:
		return nil, fmt.Errorf("field specifications must be map[string]any, string, or *Struct, got %T", spec[1])
	}
}

// newServiceListStruct creates a ListStruct from a slice of type specifications for service orchestration.
func newServiceListStruct(specs [][2]any) (*ListStruct, error) {
	structs := make([]*Struct, len(specs))
	for i, spec := range specs {
		s, err := newServiceSingleStruct(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid specification at index %d: %w", i, err)
		}
		structs[i] = s
	}
	return &ListStruct{ListFields: structs}, nil
}

// newServiceMapStruct creates a MapStruct from a map of type specifications for service orchestration.
func newServiceMapStruct(specs map[string][2]any) (*MapStruct, error) {
	structs := make(map[string]*Struct, len(specs))
	for key, spec := range specs {
		s, err := newServiceSingleStruct(spec)
		if err != nil {
			return nil, fmt.Errorf("invalid specification for key %q: %w", key, err)
		}
		structs[key] = s
	}
	return &MapStruct{MapFields: structs}, nil
}

// newServiceMap2Struct creates a Map2Struct from a 2D map of type specifications for service orchestration.
func newServiceMap2Struct(specs map[[2]string][2]any) (*Map2Struct, error) {
	groupedSpecs := make(map[string]map[string][2]any)
	for key, spec := range specs {
		if groupedSpecs[key[0]] == nil {
			groupedSpecs[key[0]] = make(map[string][2]any)
		}
		groupedSpecs[key[0]][key[1]] = spec
	}
	map2Fields := make(map[string]*MapStruct, len(groupedSpecs))
	for firstKey, secondLevelSpecs := range groupedSpecs {
		ms, err := newServiceMapStruct(secondLevelSpecs)
		if err != nil {
			return nil, fmt.Errorf("invalid specification for first-level key %q: %w", firstKey, err)
		}
		map2Fields[firstKey] = ms
	}
	return &Map2Struct{Map2Fields: map2Fields}, nil
}

// newEndStruct creates an end-node Struct from a string slice.
// The slice must have at least 2 elements: [className, serviceName].
func newEndStruct(v []string) (*Struct, error) {
	if len(v) < 1 {
		return nil, fmt.Errorf("too few elements: %d (need at least 1)", len(v))
	}
	if len(v) == 1 {
		return &Struct{ClassName: v[0]}, nil
	}
	return &Struct{ClassName: v[0], ServiceName: v[1]}, nil
}

// newEndListStruct creates a ListStruct of end-node Structs from a slice of string slices.
func newEndListStruct(v [][]string) (*ListStruct, error) {
	x := make([]*Struct, len(v))
	for i, u := range v {
		s, err := newEndStruct(u)
		if err != nil {
			return nil, fmt.Errorf("invalid end struct at index %d: %w", i, err)
		}
		x[i] = s
	}
	return &ListStruct{ListFields: x}, nil
}

// newEndMapStruct creates a MapStruct of end-node Structs from a map of string slices.
func newEndMapStruct(v map[string][]string) (*MapStruct, error) {
	x := make(map[string]*Struct)
	for key, val := range v {
		s, err := newEndStruct(val)
		if err != nil {
			return nil, fmt.Errorf("invalid end struct for key %q: %w", key, err)
		}
		x[key] = s
	}
	return &MapStruct{MapFields: x}, nil
}

// newEndMap2Struct creates a Map2Struct of end-node Structs from a two-level map.
func newEndMap2Struct(v map[[2]string][]string) (*Map2Struct, error) {
	// Group by first key
	grouped := make(map[string]map[string][]string)
	for key, val := range v {
		if grouped[key[0]] == nil {
			grouped[key[0]] = make(map[string][]string)
		}
		grouped[key[0]][key[1]] = val
	}

	x := make(map[string]*MapStruct)
	for key0, inner := range grouped {
		ms, err := newEndMapStruct(inner)
		if err != nil {
			return nil, fmt.Errorf("invalid end map struct for key %q: %w", key0, err)
		}
		x[key0] = ms
	}
	return &Map2Struct{Map2Fields: x}, nil
}

// --- Compatibility aliases ---

// GetObjectName returns ClassName (alias for backwards compatibility with grand/spec).
// ObjectName and ClassName are the same concept.
func (x *Struct) GetObjectName() string {
	if x != nil {
		return x.ClassName
	}
	return ""
}
