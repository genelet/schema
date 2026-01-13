package schema

import (
	"encoding/json"
	"fmt"
)

// jsonSchema represents a subset of JSON Schema for parsing.
type jsonSchema struct {
	ClassName            string                 `json:"className,omitempty"`
	Properties           map[string]*jsonSchema `json:"properties,omitempty"`
	Items                *jsonSchema            `json:"items,omitempty"`
	AdditionalProperties *jsonSchema            `json:"additionalProperties,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	ServiceName          string                 `json:"serviceName,omitempty"`
	XMap2                bool                   `json:"x-map2,omitempty"`
}

const (
	wrapperClassName   = "__schema_wrapper__"
	wrapperFieldName   = "__schema_value__"
	wrapperServiceName = "__schema_wrapper_service__"
)

func wrapValueAsStruct(v *Value) *Struct {
	if v == nil {
		return nil
	}
	return &Struct{
		ClassName:   wrapperClassName,
		ServiceName: wrapperServiceName,
		Fields: map[string]*Value{
			wrapperFieldName: v,
		},
	}
}

func unwrapValueFromStruct(s *Struct) (*Value, bool) {
	if s == nil {
		return nil, false
	}
	if s.ClassName != wrapperClassName || s.ServiceName != wrapperServiceName || len(s.Fields) != 1 {
		return nil, false
	}
	v, ok := s.Fields[wrapperFieldName]
	if !ok {
		return nil, false
	}
	if v == nil || v.GetSingleStruct() != nil {
		return nil, false
	}
	return v, true
}

// JSMServiceStruct creates a Struct from a JSON Schema string.
//
// converting Standard JSON Schema to Genelet Schema Struct:
//   - className: object (optional), properties -> Struct (SingleStruct)
//   - className: array (optional), items -> ListStruct
//   - className: object (optional), additionalProperties -> MapStruct
//   - custom className (MyClass) -> SingleStruct with ClassName = MyClass
//
// Arguments:
//   - className: The name of the root object.
//   - jsonSchemaStr: The JSON Schema string.
//
// Examples of jsonSchemaStr (parallel to NewServiceValue rules):
//
//	╔══════════════════════════════════════════════════════════════════════════════════════════════════════╤══════════════════╤═════════════╤════════════════════════════╗
//	║ JSON Schema                                                                                          │ Conversion       │ ClassName   │ ServiceName                ║
//	╠══════════════════════════════════════════════════════════════════════════════════════════════════════╪══════════════════╪═════════════╪════════════════════════════╣
//	║ {"className": "Circle", "serviceName": "s1"}                                                         │ SingleStruct     │ "Circle"    │ "s1"                       ║
//	║ {"className": "array", "items": {"className": "Circle", "serviceName": "s2"}}                        │ ListStruct       │ n/a         │ n/a                        ║
//	║ {"className": "object", "additionalProperties": {"className": "Circle", "serviceName": "s3"}}        │ MapStruct        │ n/a         │ n/a                        ║
//	║ {"className": "Class1", "properties": {"Field1": {"className": "Circle"}}}                           │ SingleStruct     │ "Class1"    │ ""                         ║
//	║ {"className": "array", "items": {"className": "Class2", "properties": {...}}}                        │ ListStruct       │ n/a         │ n/a                        ║
//	║ {"className": "object", "additionalProperties": {"className": "Class3", "properties": {...}}}        │ MapStruct        │ n/a         │ n/a                        ║
//	║ {"className": "object", "x-map2": true, "properties": {"r1": {"properties": {"k1": T}}}}             │ Map2Struct       │ n/a         │ n/a                        ║
//	╚══════════════════════════════════════════════════════════════════════════════════════════════════════╧══════════════════╧═════════════╧════════════════════════════╝
func JSMServiceStruct(className, jsonSchemaStr string) (*Struct, error) {
	if className == "" {
		return nil, fmt.Errorf("className cannot be empty")
	}

	var schema jsonSchema
	if err := json.Unmarshal([]byte(jsonSchemaStr), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse JSON Schema: %w", err)
	}

	value, err := convertSchemaToValue(&schema)
	if err != nil {
		return nil, err
	}
	if value == nil {
		// If the top-level schema is a primitive, we return an empty Struct with the ClassName.
		// This is debatable, but returning nil might break callers expecting a Struct.
		// Better to return an empty struct than nil.
		return &Struct{ClassName: className}, nil
	}

	if s := value.GetSingleStruct(); s != nil {
		s.ClassName = className
		return s, nil
	}

	return nil, fmt.Errorf("top-level JSON schema must be an object (SingleStruct), got %T", value.Kind)
}

// JSMStruct creates a Struct from a JSON Schema string, ensuring no ServiceName exists.
// It behaves like JSMServiceStruct but strips all service names from the result,
// behaving similarly to NewStruct.
//
// Examples of jsonSchemaStr (parallel to NewStruct rules):
//
//	╔══════════════════════════════════════════════════════════════════════════════════════════════════════╤══════════════════╤═════════════╗
//	║ JSON Schema                                                                                          │ Conversion       │ ClassName   ║
//	╠══════════════════════════════════════════════════════════════════════════════════════════════════════╪══════════════════╪═════════════╣
//	║ {"className": "Circle"}                                                                              │ SingleStruct     │ "Circle"    ║
//	║ {"className": "array", "items": {"className": "Circle"}}                                             │ ListStruct       │ n/a         ║
//	║ {"className": "object", "additionalProperties": {"className": "Circle"}}                             │ MapStruct        │ n/a         ║
//	║ {"className": "object", "x-map2": true, "properties": {"r1": {"properties": {"k1": T}}}}             │ Map2Struct       │ n/a         ║
//	║ {"className": "Class1", "properties": {"Field1": {"className": "Circle"}}}                           │ SingleStruct     │ "Class1"    ║
//	║ {"className": "array", "items": {"className": "Class2", "properties": {...}}}                        │ ListStruct       │ n/a         ║
//	║ {"className": "object", "additionalProperties": {"className": "Class3", "properties": {...}}}        │ MapStruct        │ n/a         ║
//	╚══════════════════════════════════════════════════════════════════════════════════════════════════════╧══════════════════╧═════════════╝
func JSMStruct(className, jsonSchemaStr string) (*Struct, error) {
	s, err := JSMServiceStruct(className, jsonSchemaStr)
	if err != nil {
		return nil, err
	}
	return DeriveStructWithoutServices(s), nil
}

func convertSchemaToValue(js *jsonSchema) (*Value, error) {
	if js == nil {
		return nil, fmt.Errorf("nil schema")
	}

	// 0. If "x-map2" is true, it is a Map2Struct.
	// It relies on 2-layer Properties: Region -> Key -> Service
	if js.XMap2 {
		if js.Properties == nil {
			return nil, fmt.Errorf("x-map2 requires properties")
		}
		map2Fields := make(map[string]*MapStruct)
		for regionKey, regionSchema := range js.Properties {
			// regionSchema should have properties for the inner map
			if regionSchema.Properties == nil {
				return nil, fmt.Errorf("x-map2 region %q missing properties", regionKey)
			}
			innerMapFields := make(map[string]*Struct)
			for innerKey, innerSchema := range regionSchema.Properties {
				val, err := convertSchemaToValue(innerSchema)
				if err != nil {
					return nil, fmt.Errorf("in x-map2 region %q key %q: %w", regionKey, innerKey, err)
				}
				if val == nil {
					continue // Ignore primitives in map2
				}
				innerMapFields[innerKey] = extractStructFromValue(val)
			}
			if len(innerMapFields) > 0 {
				map2Fields[regionKey] = &MapStruct{MapFields: innerMapFields}
			}
		}
		if len(map2Fields) == 0 {
			return nil, nil // If all fields were ignored
		}
		return &Value{Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{Map2Fields: map2Fields}}}, nil
	}

	// 1. If "properties" is present, it is a Struct (SingleStruct).
	if js.Properties != nil {
		fields := make(map[string]*Value)
		for name, prop := range js.Properties {
			val, err := convertSchemaToValue(prop)
			if err != nil {
				return nil, fmt.Errorf("in property %q: %w", name, err)
			}
			if val != nil {
				fields[name] = val
			}
		}
		s := &Struct{Fields: fields, ServiceName: js.ServiceName}
		if js.ClassName != "" {
			s.ClassName = js.ClassName
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: s}}, nil
	}

	// 2. If "additionalProperties" is present, it is a MapStruct.
	if js.AdditionalProperties != nil {
		val, err := convertSchemaToValue(js.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil // Ignore Map of primitives
		}
		targetStruct := extractStructFromValue(val)
		return &Value{Kind: &Value_MapStruct{MapStruct: &MapStruct{MapFields: map[string]*Struct{"*": targetStruct}}}}, nil
	}

	// 3. If "items" is present, it is a ListStruct (Array).
	if js.Items != nil {
		itemVal, err := convertSchemaToValue(js.Items)
		if err != nil {
			return nil, err
		}
		if itemVal == nil {
			return nil, nil // Ignore List of primitives
		}
		itemStruct := extractStructFromValue(itemVal)
		return &Value{Kind: &Value_ListStruct{ListStruct: &ListStruct{ListFields: []*Struct{itemStruct}}}}, nil
	}

	// 4. Custom Class / Leaf
	// Treating as CUSTOM CLASS (opaque class with ClassName = type)
	// This captures "MyType".
	return &Value{Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: js.ClassName, ServiceName: js.ServiceName}}}, nil
}

// extractStructFromValue attempts to get a Struct from a Value.
// If Value is not a SingleStruct, it wraps it or creates a dummy Struct.
func extractStructFromValue(v *Value) *Struct {
	if s := v.GetSingleStruct(); s != nil {
		return s
	}
	// Wrap non-SingleStruct values so nested list/map types can round-trip.
	return wrapValueAsStruct(v)
}

// MarshalJSON implements the json.Marshaler interface.
// It converts the Struct into the Genelet JSON Schema format.
func (s *Struct) MarshalJSON() ([]byte, error) {
	js, err := convertStructToSchema(s)
	if err != nil {
		return nil, err
	}
	return json.Marshal(js)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses the Genelet JSON Schema format into the Struct.
func (s *Struct) UnmarshalJSON(data []byte) error {
	var js jsonSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return err
	}

	val, err := convertSchemaToValue(&js)
	if err != nil {
		return err
	}
	if val == nil {
		// Empty struct if primitive/ignored
		s.ClassName = js.ClassName
		s.ServiceName = js.ServiceName
		s.Fields = nil
		return nil
	}

	// Extract the single struct from the value
	extracted := extractStructFromValue(val)
	s.ClassName = extracted.ClassName
	s.ServiceName = extracted.ServiceName
	s.Fields = extracted.Fields

	// If the top-level schema had a class name, ensure it's preserved
	// (extractStructFromValue might return a wrapper or missing name if it came from non-SingleStruct)
	if js.ClassName != "" {
		s.ClassName = js.ClassName
	}
	// Same for ServiceName
	if js.ServiceName != "" {
		s.ServiceName = js.ServiceName
	}

	return nil
}

func convertStructToSchema(s *Struct) (*jsonSchema, error) {
	if s == nil {
		return nil, nil
	}
	if v, ok := unwrapValueFromStruct(s); ok {
		return convertValueToSchema(v)
	}
	js := &jsonSchema{
		ClassName:   s.ClassName,
		ServiceName: s.ServiceName,
	}

	if len(s.Fields) > 0 {
		js.Properties = make(map[string]*jsonSchema)
		for name, val := range s.Fields {
			propJs, err := convertValueToSchema(val)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", name, err)
			}
			if propJs != nil {
				js.Properties[name] = propJs
			}
		}
	}
	return js, nil
}

func convertValueToSchema(v *Value) (*jsonSchema, error) {
	if v == nil {
		return nil, nil
	}

	switch k := v.Kind.(type) {
	case *Value_SingleStruct:
		return convertStructToSchema(k.SingleStruct)

	case *Value_ListStruct:
		// ListStruct: items -> schema of first element
		ls := k.ListStruct
		if len(ls.ListFields) == 0 {
			// Empty list, cannot determine item schema.
			// Return empty schema or error?
			// Return schema with empty items to indicate array type but unknown element.
			return &jsonSchema{Items: &jsonSchema{}}, nil
		}
		itemJs, err := convertStructToSchema(ls.ListFields[0])
		if err != nil {
			return nil, err
		}
		return &jsonSchema{Items: itemJs}, nil

	case *Value_MapStruct:
		// MapStruct: additionalProperties -> schema of "*" element
		ms := k.MapStruct
		if len(ms.MapFields) == 0 {
			return &jsonSchema{AdditionalProperties: &jsonSchema{}}, nil
		}
		// Look for "*" key specifically? Or just pick one?
		// JSMServiceStruct uses "*" as the key for values.
		// If constructed manually, might be specific keys.
		// We maintain the "MapStruct = Map of T" semantic.
		// If "*" exists, use it. Else pick any (assuming homogeneity).
		var target *Struct
		if s, ok := ms.MapFields["*"]; ok {
			target = s
		} else {
			// iter and pick first
			for _, s := range ms.MapFields {
				target = s
				break
			}
		}
		valJs, err := convertStructToSchema(target)
		if err != nil {
			return nil, err
		}
		return &jsonSchema{AdditionalProperties: valJs}, nil

	case *Value_Map2Struct:
		// Map2Struct: x-map2: true, Nested properties
		m2s := k.Map2Struct
		js := &jsonSchema{XMap2: true, Properties: make(map[string]*jsonSchema)}
		for regionName, mapStruct := range m2s.Map2Fields {
			// Each value in Map2Fields is a MapStruct
			// This MapStruct contains the inner keys -> Structs
			innerProps := make(map[string]*jsonSchema)
			for innerKey, innerStruct := range mapStruct.MapFields {
				innerJs, err := convertStructToSchema(innerStruct)
				if err != nil {
					return nil, err
				}
				innerProps[innerKey] = innerJs
			}
			js.Properties[regionName] = &jsonSchema{Properties: innerProps}
		}
		return js, nil

	default:
		return nil, fmt.Errorf("unknown Value kind: %T", v.Kind)
	}
}
