package schema

import (
	"strings"
	"testing"
)

// Test structs for validation
type validStruct struct {
	Name     string
	Age      int
	Child    *childStruct
	Children []*childStruct
	ByName   map[string]*childStruct
	ByKeys   map[[2]string]*childStruct
}

type childStruct struct {
	Value string
}

func TestValidateStruct_NilSpec(t *testing.T) {
	obj := &validStruct{}
	err := ValidateStruct(obj, nil)
	if err != nil {
		t.Errorf("ValidateStruct with nil spec should return nil, got %v", err)
	}
}

func TestValidateStruct_EmptyFields(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{ClassName: "validStruct"}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct with empty fields should return nil, got %v", err)
	}
}

func TestValidateStruct_NilObject(t *testing.T) {
	spec := &Struct{
		ClassName: "test",
		Fields:    map[string]*Value{"Name": {}},
	}
	err := ValidateStruct(nil, spec)
	if err == nil {
		t.Error("ValidateStruct with nil object should return error")
	}
}

func TestValidateStruct_NonStruct(t *testing.T) {
	spec := &Struct{
		ClassName: "test",
		Fields:    map[string]*Value{"Name": {}},
	}
	err := ValidateStruct("not a struct", spec)
	if err == nil {
		t.Error("ValidateStruct with non-struct object should return error")
	}
}

func TestValidateStruct_MissingField(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"NonExistent": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "foo"}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err == nil {
		t.Error("ValidateStruct should return error for non-existent field")
	}
	if !strings.Contains(err.Error(), "NonExistent") {
		t.Errorf("error should mention field name, got: %v", err)
	}
}

func TestValidateStruct_SingleStruct_Valid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Child": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "childStruct"}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for valid SingleStruct field, got %v", err)
	}
}

func TestValidateStruct_SingleStruct_Invalid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Name": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "foo"}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err == nil {
		t.Error("ValidateStruct should return error for SingleStruct on string field")
	}
	if !strings.Contains(err.Error(), "SingleStruct") {
		t.Errorf("error should mention SingleStruct, got: %v", err)
	}
}

func TestValidateStruct_ListStruct_Valid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Children": {Kind: &Value_ListStruct{ListStruct: &ListStruct{
				ListFields: []*Struct{{ClassName: "childStruct"}},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for valid ListStruct field, got %v", err)
	}
}

func TestValidateStruct_ListStruct_Invalid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Name": {Kind: &Value_ListStruct{ListStruct: &ListStruct{
				ListFields: []*Struct{{ClassName: "foo"}},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err == nil {
		t.Error("ValidateStruct should return error for ListStruct on string field")
	}
	if !strings.Contains(err.Error(), "ListStruct") {
		t.Errorf("error should mention ListStruct, got: %v", err)
	}
}

func TestValidateStruct_MapStruct_Valid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"ByName": {Kind: &Value_MapStruct{MapStruct: &MapStruct{
				MapFields: map[string]*Struct{"key": {ClassName: "childStruct"}},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for valid MapStruct field, got %v", err)
	}
}

func TestValidateStruct_MapStruct_Invalid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Age": {Kind: &Value_MapStruct{MapStruct: &MapStruct{
				MapFields: map[string]*Struct{"key": {ClassName: "foo"}},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err == nil {
		t.Error("ValidateStruct should return error for MapStruct on int field")
	}
	if !strings.Contains(err.Error(), "MapStruct") {
		t.Errorf("error should mention MapStruct, got: %v", err)
	}
}

func TestValidateStruct_Map2Struct_Valid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"ByKeys": {Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{
				Map2Fields: map[string]*MapStruct{
					"key1": {MapFields: map[string]*Struct{"key2": {ClassName: "childStruct"}}},
				},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for valid Map2Struct field, got %v", err)
	}
}

func TestValidateStruct_Map2Struct_Invalid(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Children": {Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{
				Map2Fields: map[string]*MapStruct{
					"key1": {MapFields: map[string]*Struct{"key2": {ClassName: "foo"}}},
				},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err == nil {
		t.Error("ValidateStruct should return error for Map2Struct on slice field")
	}
	if !strings.Contains(err.Error(), "Map2Struct") {
		t.Errorf("error should mention Map2Struct, got: %v", err)
	}
}

func TestValidateStruct_ListStructOnMap_Valid(t *testing.T) {
	// ListStruct on map is valid (map treated as ordered collection)
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"ByName": {Kind: &Value_ListStruct{ListStruct: &ListStruct{
				ListFields: []*Struct{{ClassName: "childStruct"}},
			}}},
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for ListStruct on map field, got %v", err)
	}
}

func TestValidateStruct_NilValue(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Name": nil,
		},
	}
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for nil Value, got %v", err)
	}
}

func TestValidateStruct_PointerObject(t *testing.T) {
	obj := &validStruct{}
	spec := &Struct{
		ClassName: "validStruct",
		Fields: map[string]*Value{
			"Child": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "childStruct"}}},
		},
	}
	// Test with pointer (normal case)
	err := ValidateStruct(obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for pointer object, got %v", err)
	}

	// Test with value (should also work)
	err = ValidateStruct(*obj, spec)
	if err != nil {
		t.Errorf("ValidateStruct should pass for value object, got %v", err)
	}
}
