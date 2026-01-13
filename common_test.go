package schema

import (
	"strings"
	"testing"
)

func TestCommonString(t *testing.T) {
	// In the Geo struct,
	//  - EndString is the field name,
	//  - Circle1 is the object (or Struct) name, pointed by EndString.
	//  - service1 is the service name, to serve Circle1.
	//
	// EndList is another field name for list of services.
	// EndMap is another field name for map of services.
	//
	// And line with ',' means an object or Struct
	//  - the first is the object name
	//  - the second could be map[string]any, means all fields in this middle object (or Struct)
	//  - the second could be string or list of strings, means the ending service name and optional labels
	//
	// Any line with ':' means a field name to Struct mapping.
	//  - a Struct pointer means TBD struct
	//  - a slice of string, end struct with service name and optional labels
	//  - a slice of slice of string, end list struct with service name and optional labels
	//  - [2]any means a single middle Struct
	//  - [][2]any means a list of middle Struct
	//
	sp, err := NewServiceStruct(
		"Geo", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
			"EndMap": map[string][]string{
				"key1": {"Circle2", "service2"},
				"key2": {"Circle3", "service3"},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	testSpec(t, sp)
}

func testSpec(t *testing.T, sp *Struct) {
	fields := sp.GetFields()

	s := fields["EndString"].GetSingleStruct()
	slist := fields["EndList"].GetListStruct()
	smap := fields["EndMap"].GetMapStruct()

	if s.ClassName != "Circle1" || s.ServiceName != "service1" {
		t.Errorf("%#v", s.Fields)
	}
	s = slist.ListFields[0]
	if s.ClassName != "Circle2" || s.ServiceName != "service2" {
		t.Errorf("%#v", s.Fields)
	}
	s = slist.ListFields[1]
	if s.ClassName != "Circle3" || s.ServiceName != "service3" {
		t.Errorf("%#v", s.Fields)
	}
	s = smap.MapFields["key1"]
	if s.ClassName != "Circle2" || s.ServiceName != "service2" {
		t.Errorf("%#v", s.Fields)
	}
	s = smap.MapFields["key2"]
	if s.ClassName != "Circle3" || s.ServiceName != "service3" {
		t.Errorf("%#v", s.Fields)
	}
}

func TestCommonInterface(t *testing.T) {
	// In the Geo struct, TheMiddle and TheMiddleList are the field names.
	// Circle is the object (or Struct) name, pointed by TheMiddle.
	// In Cirle, there are three fields with names EndString, EndList and EndMap.
	//
	sp, err := NewServiceStruct(
		"Geo", map[string]any{
			"TheMiddle": [2]any{"Circle", map[string]any{
				"EndString": []string{"Circle1", "service1"},
				"EndList": [][]string{
					{"Circle2", "service2"},
					{"Circle3", "service3"},
				},
				"EndMap": map[string][]string{
					"key1": {"Circle2", "service2"},
					"key2": {"Circle3", "service3"},
				},
			}},
			"TheMiddleList": [][2]any{
				{"CircleStruct1", map[string]any{
					"EndString": []string{"Circle1", "service1"},
					"EndList": [][]string{
						{"Circle2", "service2"},
						{"Circle3", "service3"},
					},
					"EndMap": map[string][]string{
						"key1": {"Circle2", "service2"},
						"key2": {"Circle3", "service3"},
					},
				}},
				{"CircleStruct2", map[string]any{
					"EndString": []string{"Circle1", "service1"},
					"EndList": [][]string{
						{"Circle2", "service2"},
						{"Circle3", "service3"},
					},
					"EndMap": map[string][]string{
						"key1": {"Circle2", "service2"},
						"key2": {"Circle3", "service3"},
					},
				}},
			},
			"TheMiddleMap": map[string][2]any{
				"key3": {"CircleStruct1", map[string]any{
					"EndString": []string{"Circle1", "service1"},
					"EndList": [][]string{
						{"Circle2", "service2"},
						{"Circle3", "service3"},
					},
					"EndMap": map[string][]string{
						"key1": {"Circle2", "service2"},
						"key2": {"Circle3", "service3"},
					},
				}},
				"key4": {"CircleStruct2", map[string]any{
					"EndString": []string{"Circle1", "service1"},
					"EndList": [][]string{
						{"Circle2", "service2"},
						{"Circle3", "service3"},
					},
					"EndMap": map[string][]string{
						"key1": {"Circle2", "service2"},
						"key2": {"Circle3", "service3"},
					},
				}},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fields := sp.GetFields()

	middle := fields["TheMiddle"].GetSingleStruct()
	if middle.ClassName != "Circle" {
		t.Errorf("%#v", middle.ClassName)
	}
	testSpec(t, middle)

	middleList := fields["TheMiddleList"].GetListStruct()
	if middleList.ListFields[0].ClassName != "CircleStruct1" {
		t.Errorf("%#v", middleList.ListFields[0].ClassName)
	}
	testSpec(t, middleList.ListFields[0])
	if middleList.ListFields[1].ClassName != "CircleStruct2" {
		t.Errorf("%#v", middleList.ListFields[1].ClassName)
	}
	testSpec(t, middleList.ListFields[1])

	middleMap := fields["TheMiddleMap"].GetMapStruct()
	mapFields := middleMap.MapFields
	if mapFields["key3"].ClassName != "CircleStruct1" {
		t.Errorf("%#v", mapFields["key3"].ClassName)
	}
	testSpec(t, mapFields["key3"])
	if mapFields["key4"].ClassName != "CircleStruct2" {
		t.Errorf("%#v", mapFields["key4"].ClassName)
	}
	testSpec(t, mapFields["key4"])
}

func TestCommonStruct(t *testing.T) {
	spec1, err := NewServiceStruct(
		"Class1", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
			"EndMap": map[string][]string{
				"key1": {"Circle2", "service2"},
				"key2": {"Circle3", "service3"},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	spec2, err := NewServiceStruct(
		"Class2", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
			"EndMap": map[string][]string{
				"key1": {"Circle2", "service2"},
				"key2": {"Circle3", "service3"},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	spec3, err := NewServiceStruct(
		"Class3", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
			"EndMap": map[string][]string{
				"key1": {"Circle2", "service2"},
				"key2": {"Circle3", "service3"},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	sp, err := NewStruct(
		"Geo", map[string]any{
			"Shape1": spec1,
			"Shape2": []*Struct{spec1, spec2, spec3},
			"Shape3": map[string]*Struct{"key1": spec1, "key2": spec2, "key3": spec3},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fields := sp.GetFields()

	s := fields["Shape1"].GetSingleStruct()
	if s.ClassName != "Class1" {
		t.Errorf("%#v", s.ClassName)
	}
	testSpec(t, s)
	slist := fields["Shape2"].GetListStruct()
	if slist.ListFields[0].ClassName != "Class1" {
		t.Errorf("%#v", slist.ListFields[0].ClassName)
	}
	testSpec(t, slist.ListFields[0])
	if slist.ListFields[1].ClassName != "Class2" {
		t.Errorf("%#v", slist.ListFields[1].ClassName)
	}
	testSpec(t, slist.ListFields[1])
	if slist.ListFields[2].ClassName != "Class3" {
		t.Errorf("%#v", slist.ListFields[2].ClassName)
	}
	testSpec(t, slist.ListFields[2])
	smap := fields["Shape3"].GetMapStruct()
	mapFields := smap.MapFields
	if mapFields["key1"].ClassName != "Class1" {
		t.Errorf("%#v", mapFields["key1"].ClassName)
	}
	testSpec(t, mapFields["key1"])
	if mapFields["key2"].ClassName != "Class2" {
		t.Errorf("%#v", mapFields["key2"].ClassName)
	}
	testSpec(t, mapFields["key2"])
	if mapFields["key3"].ClassName != "Class3" {
		t.Errorf("%#v", mapFields["key3"].ClassName)
	}
	testSpec(t, mapFields["key3"])
}

// TestServiceValueWithStructInSpec tests that NewServiceValue can accept *Struct
// directly as the second element in [2]any.
func TestServiceValueWithStructInSpec(t *testing.T) {
	// Create an inner struct
	innerStruct, err := NewServiceStruct(
		"InnerClass", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
			"EndMap": map[string][]string{
				"key1": {"Circle2", "service2"},
				"key2": {"Circle3", "service3"},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Now use the struct directly in [2]any for NewServiceValue
	sp, err := NewServiceStruct(
		"Geo", map[string]any{
			// Test single struct using [2]any with *Struct as second element
			"TheMiddle": [2]any{"WrapperClass", innerStruct},
			// Test list of structs using [][2]any with *Struct as second element
			"TheMiddleList": [][2]any{
				{"WrapperClass1", innerStruct},
				{"WrapperClass2", innerStruct},
			},
			// Test map of structs using map[string][2]any with *Struct as second element
			"TheMiddleMap": map[string][2]any{
				"key1": {"WrapperClass1", innerStruct},
				"key2": {"WrapperClass2", innerStruct},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fields := sp.GetFields()

	// Verify TheMiddle field
	middle := fields["TheMiddle"].GetSingleStruct()
	if middle.ClassName != "WrapperClass" {
		t.Errorf("expected ClassName 'WrapperClass', got %q", middle.ClassName)
	}
	// Verify that the inner struct's fields were copied
	if middle.Fields == nil {
		t.Error("expected Fields to be non-nil")
	}
	testSpec(t, middle)

	// Verify TheMiddleList field
	middleList := fields["TheMiddleList"].GetListStruct()
	if len(middleList.ListFields) != 2 {
		t.Errorf("expected 2 list items, got %d", len(middleList.ListFields))
	}
	if middleList.ListFields[0].ClassName != "WrapperClass1" {
		t.Errorf("expected ClassName 'WrapperClass1', got %q", middleList.ListFields[0].ClassName)
	}
	testSpec(t, middleList.ListFields[0])
	if middleList.ListFields[1].ClassName != "WrapperClass2" {
		t.Errorf("expected ClassName 'WrapperClass2', got %q", middleList.ListFields[1].ClassName)
	}
	testSpec(t, middleList.ListFields[1])

	// Verify TheMiddleMap field
	middleMap := fields["TheMiddleMap"].GetMapStruct()
	mapFields := middleMap.MapFields
	if mapFields["key1"].ClassName != "WrapperClass1" {
		t.Errorf("expected ClassName 'WrapperClass1', got %q", mapFields["key1"].ClassName)
	}
	testSpec(t, mapFields["key1"])
	if mapFields["key2"].ClassName != "WrapperClass2" {
		t.Errorf("expected ClassName 'WrapperClass2', got %q", mapFields["key2"].ClassName)
	}
	testSpec(t, mapFields["key2"])
}

func TestNewServiceStruct_ServiceNameMustBeLeaf(t *testing.T) {
	bad := &Struct{
		ClassName:   "Inner",
		ServiceName: "svc",
		Fields: map[string]*Value{
			"Field": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "Leaf"}}},
		},
	}

	_, err := NewServiceStruct("Wrapper", bad)
	if err == nil {
		t.Fatal("expected error for service name on non-leaf struct")
	}
}

func TestNewServiceValue_ServiceNameMustBeLeaf(t *testing.T) {
	bad := &Struct{
		ClassName:   "Inner",
		ServiceName: "svc",
		Fields: map[string]*Value{
			"Field": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "Leaf"}}},
		},
	}

	if _, err := NewServiceValue(bad); err == nil {
		t.Fatal("expected error for service name on non-leaf struct")
	}
}

func TestNewServiceStruct_ServiceNameLeafAllowed(t *testing.T) {
	leaf := &Struct{
		ClassName:   "Inner",
		ServiceName: "svc",
	}

	got, err := NewServiceStruct("Wrapper", leaf)
	if err != nil {
		t.Fatalf("expected leaf service struct to pass, got %v", err)
	}
	if got.ServiceName != "svc" {
		t.Fatalf("expected ServiceName to be preserved, got %q", got.ServiceName)
	}
}

func TestNewServiceStruct_ServiceNameMustBeLeafErrorMessage(t *testing.T) {
	bad := &Struct{
		ClassName:   "Inner",
		ServiceName: "svc",
		Fields: map[string]*Value{
			"Field": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "Leaf"}}},
		},
	}

	_, err := NewServiceStruct("Wrapper", bad)
	if err == nil {
		t.Fatal("expected error for service name on non-leaf struct")
	}
	if !strings.Contains(err.Error(), "service name must be on leaf struct") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewServiceValueDirectTypes(t *testing.T) {
	// Test *Struct
	structSpec := &Struct{ClassName: "DirectStruct"}
	v, err := NewServiceValue(structSpec)
	if err != nil {
		t.Fatal(err)
	}
	if v.GetSingleStruct().ClassName != "DirectStruct" {
		t.Errorf("expected DirectStruct, got %s", v.GetSingleStruct().ClassName)
	}

	// Test []*Struct
	listSpec := []*Struct{{ClassName: "List1"}, {ClassName: "List2"}}
	v, err = NewServiceValue(listSpec)
	if err != nil {
		t.Fatal(err)
	}
	ls := v.GetListStruct()
	if len(ls.ListFields) != 2 || ls.ListFields[0].ClassName != "List1" {
		t.Error("list struct mismatch")
	}

	// Test map[string]*Struct
	mapSpec := map[string]*Struct{"k1": {ClassName: "Map1"}}
	v, err = NewServiceValue(mapSpec)
	if err != nil {
		t.Fatal(err)
	}
	ms := v.GetMapStruct()
	if ms.MapFields["k1"].ClassName != "Map1" {
		t.Error("map struct mismatch")
	}

	// Test map[string]*MapStruct
	map2Spec := map[string]*MapStruct{"k1": {MapFields: map[string]*Struct{"k2": {ClassName: "Map2"}}}}
	v, err = NewServiceValue(map2Spec)
	if err != nil {
		t.Fatal(err)
	}
	m2s := v.GetMap2Struct()
	if m2s.Map2Fields["k1"].MapFields["k2"].ClassName != "Map2" {
		t.Error("map2 struct mismatch")
	}
}
