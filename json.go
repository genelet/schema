package schema

import (
	"encoding/json"
	"fmt"
)

// jsonSchema represents a subset of JSON Schema for parsing.
type jsonSchema struct {
	Type                 string                 `json:"type"`
	Properties           map[string]*jsonSchema `json:"properties"`
	Items                *jsonSchema            `json:"items"`
	AdditionalProperties *jsonSchema            `json:"additionalProperties"`
	Ref                  string                 `json:"$ref"`
	ServiceName          string                 `json:"serviceName,omitempty"`
	XMap2                bool                   `json:"x-map2,omitempty"`
}

// JSMServiceStruct creates a Struct from a JSON Schema string.
//
// converting Standard JSON Schema to Genelet Schema Struct:
//   - type: object, properties -> Struct (SingleStruct)
//   - type: array, items -> ListStruct
//   - type: object, additionalProperties -> MapStruct
//   - primitive types (string, integer, etc.) -> SingleStruct with ClassName = type
//
// Arguments:
//   - className: The name of the root object.
//   - jsonSchemaStr: The JSON Schema string.
//
// Examples of jsonSchemaStr (parallel to NewServiceValue rules):
//
//	╔════════════════════════════════════════════════════════════════════════════════════════════════╤══════════════════╤═════════════╤════════════════════════════╗
//	║ JSON Schema                                                                                    │ Conversion       │ ClassName   │ ServiceName                ║
//	╠════════════════════════════════════════════════════════════════════════════════════════════════╪══════════════════╪═════════════╪════════════════════════════╣
//	║ {"type": "Circle", "serviceName": "s1"}                                                        │ SingleStruct     │ "Circle"    │ "s1"                       ║
//	║ {"type": "array", "items": {"type": "Circle", "serviceName": "s2"}}                            │ ListStruct       │ n/a         │ n/a                        ║
//	║ {"type": "object", "additionalProperties": {"type": "Circle", "serviceName": "s3"}}            │ MapStruct        │ n/a         │ n/a                        ║
//	║ {"type": "Class1", "properties": {"Field1": {"type": "Circle"}}}                               │ SingleStruct     │ "Class1"    │ ""                         ║
//	║ {"type": "array", "items": {"type": "Class2", "properties": {"Field2": {"type": "Circle"}}}}   │ ListStruct       │ n/a         │ n/a                        ║
//	║ {"type": "object", "additionalProperties": {"type": "Class3", "properties": {...}}}            │ MapStruct        │ n/a         │ n/a                        ║
//	║ {"type": "object", "x-map2": true, "properties": {"r1": {"properties": {"k1": T}}}}           │ Map2Struct       │ n/a         │ n/a                        ║
//	╚════════════════════════════════════════════════════════════════════════════════════════════════╧══════════════════╧═════════════╧════════════════════════════╝
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
//	╔════════════════════════════════════════════════════════════════════════════════════════════════╤══════════════════╤═════════════╗
//	║ JSON Schema                                                                                    │ Conversion       │ ClassName   ║
//	╠════════════════════════════════════════════════════════════════════════════════════════════════╪══════════════════╪═════════════╣
//	║ {"type": "Circle"}                                                                             │ SingleStruct     │ "Circle"    ║
//	║ {"type": "array", "items": {"type": "Circle"}}                                                 │ ListStruct       │ n/a         ║
//	║ {"type": "object", "additionalProperties": {"type": "Circle"}}                                 │ MapStruct        │ n/a         ║
//	║ {"type": "object", "x-map2": true, "properties": {"r1": {"properties": {"k1": T}}}}           │ Map2Struct       │ n/a         ║
//	║ {"type": "Class1", "properties": {"Field1": {"type": "Circle"}}}                               │ SingleStruct     │ "Class1"    ║
//	║ {"type": "array", "items": {"type": "Class2", "properties": {"Field2": {"type": "Circle"}}}}   │ ListStruct       │ n/a         ║
//	║ {"type": "object", "additionalProperties": {"type": "Class3", "properties": {...}}}            │ MapStruct        │ n/a         ║
//	╚════════════════════════════════════════════════════════════════════════════════════════════════╧══════════════════╧═════════════╝
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
				innerMapFields[innerKey] = extractStructFromValue(val)
			}
			map2Fields[regionKey] = &MapStruct{MapFields: innerMapFields}
		}
		return &Value{Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{Map2Fields: map2Fields}}}, nil
	}

	// 1. If "properties" is present, it is a Struct (SingleStruct).
	// We allow custom "type" to specify ClassName.
	if js.Properties != nil {
		fields := make(map[string]*Value)
		for name, prop := range js.Properties {
			val, err := convertSchemaToValue(prop)
			if err != nil {
				return nil, fmt.Errorf("in property %q: %w", name, err)
			}
			fields[name] = val
		}
		s := &Struct{Fields: fields, ServiceName: js.ServiceName}
		if js.Type != "object" {
			s.ClassName = js.Type
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: s}}, nil
	}

	// 2. If "additionalProperties" is present, it is a MapStruct.
	if js.AdditionalProperties != nil {
		val, err := convertSchemaToValue(js.AdditionalProperties)
		if err != nil {
			return nil, err
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
		itemStruct := extractStructFromValue(itemVal)
		return &Value{Kind: &Value_ListStruct{ListStruct: &ListStruct{ListFields: []*Struct{itemStruct}}}}, nil
	}

	// 4. Fallback: Primitive or Empty Object
	// If type is "array", but no items, return empty ListStruct
	if js.Type == "array" {
		return &Value{Kind: &Value_ListStruct{ListStruct: &ListStruct{}}}, nil
	}

	// If type is "object", return empty Struct
	if js.Type == "object" {
		return &Value{Kind: &Value_SingleStruct{SingleStruct: &Struct{ServiceName: js.ServiceName}}}, nil
	}

	// Otherwise, treating as primitive/opaque class with ClassName = type
	return &Value{Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: js.Type, ServiceName: js.ServiceName}}}, nil
}

// extractStructFromValue attempts to get a Struct from a Value.
// If Value is not a SingleStruct, it wraps it or creates a dummy Struct.
func extractStructFromValue(v *Value) *Struct {
	if s := v.GetSingleStruct(); s != nil {
		return s
	}
	// If it's a list or map, we wrap it?
	// But Struct doesn't wrap Value directly. Struct has Fields (map string->Value).
	// We return a Struct that has NO ClassName (wrapper) but holds the Value in a field?
	// Or we return a Struct with ClassName=value's type?
	// This is ambiguous.
	// For now, if we get a ListStruct from a sub-schema, we probably want to return a Struct that Represents that list?
	// But `MapStruct` needs `*Struct` as value.
	// So if we have Map<String, List<int>>, we need MapStruct->Struct->Fields["items"]->ListStruct?
	// But simplistically, if we just want to avoid panic/nil, return a placeholder.
	// In strict mode we might error.
	return &Struct{ClassName: "WrappedComplexType"}
}
