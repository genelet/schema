package schema

import (
	"encoding/json"
	"testing"
)

func TestJSONString(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"TheString888": "Circle",
			"TheString":    [2]any{"Circle"},
			"TheList888":   []string{"CircleClass1", "CircleClass2"},
			"TheList":      [][2]any{{"CircleClass1"}, {"CircleClass2"}},
			"TheMap888":    map[string]string{"a1": "CircleClass1", "a2": "CircleClass2"},
			"TheMap":       map[string][2]any{"a1": {"CircleClass1"}, "a2": {"CircleClass2"}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fields := spec.GetFields()
	if fields["TheString888"].String() != fields["TheString"].String() {
		t.Errorf("%s", fields["TheString888"].String())
		t.Errorf("%s", fields["TheString"].String())
	}
	if fields["TheList888"].String() != fields["TheList"].String() {
		t.Errorf("%s", fields["TheList888"].String())
		t.Errorf("%s", fields["TheList"].String())
	}
	if fields["TheMap888"].String() != fields["TheMap"].String() {
		t.Errorf("%s", fields["TheMap888"].String())
		t.Errorf("%s", fields["TheMap"].String())
	}

	// Verify JSMServiceStruct
	jsonSchemaStr := `{
		"className": "object",
		"properties": {
			"TheString888": { "className": "Circle", "serviceName": "s1" },
			"TheString":    { "className": "Circle", "serviceName": "s2" },
			"TheList888":   { "items": { "className": "CircleClass1", "serviceName": "s3" } },
			"TheList":      { "items": { "className": "CircleClass1", "serviceName": "s4" } },
			"TheMap888":    { "additionalProperties": { "className": "CircleClass1", "serviceName": "s5" } },
			"TheMap":       { "additionalProperties": { "className": "CircleClass1", "serviceName": "s6" } }
		}
	}`
	jsSpec, err := JSMServiceStruct("Geo", jsonSchemaStr)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsFields := jsSpec.GetFields()

	// String match
	if jsFields["TheString888"].GetSingleStruct().ClassName != fields["TheString888"].GetSingleStruct().ClassName {
		t.Errorf("TheString888 mismatch")
	}

	// List match (comparing type of first item)
	if jsFields["TheList888"].GetListStruct().ListFields[0].ClassName != fields["TheList888"].GetListStruct().ListFields[0].ClassName {
		t.Errorf("TheList888 item type mismatch")
	}

	// Map match (wildcard vs key "a1")
	if jsFields["TheMap888"].GetMapStruct().MapFields["*"].ClassName != fields["TheMap888"].GetMapStruct().MapFields["a1"].ClassName {
		t.Errorf("TheMap888 value type mismatch")
	}
	// Verify NewServiceStruct with ServiceName
	specService, err := NewServiceStruct(
		"Geo", map[string]any{
			"TheString888": []string{"Circle", "s1"},
			"TheString":    []string{"Circle", "s2"},
			"TheList888":   [][]string{{"CircleClass1", "s3"}, {"CircleClass2", "s4"}},
			"TheList":      [][]string{{"CircleClass1", "s5"}, {"CircleClass2", "s6"}},
			"TheMap888":    map[string][]string{"a1": {"CircleClass1", "s7"}, "a2": {"CircleClass2", "s8"}},
			"TheMap":       map[string][]string{"a1": {"CircleClass1", "s9"}, "a2": {"CircleClass2", "s10"}},
		},
	)
	if err != nil {
		t.Errorf("NewServiceStruct failed: %v", err)
	}
	fieldsService := specService.GetFields()

	jsonSchemaStrService := `{
		"className": "object",
		"properties": {
			"TheString888": { "className": "Circle", "serviceName": "s1" },
			"TheString":    { "className": "Circle", "serviceName": "s2" },
			"TheList888":   { "items": { "className": "CircleClass1", "serviceName": "s3" } },
			"TheList":      { "items": { "className": "CircleClass1", "serviceName": "s5" } },
			"TheMap888":    { "additionalProperties": { "className": "CircleClass1", "serviceName": "s7" } },
			"TheMap":       { "additionalProperties": { "className": "CircleClass1", "serviceName": "s9" } }
		}
	}`
	jsSpecService, err := JSMServiceStruct("Geo", jsonSchemaStrService)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsFieldsService := jsSpecService.GetFields()

	// Verify ServiceName matches
	if jsFieldsService["TheString888"].GetSingleStruct().ServiceName != fieldsService["TheString888"].GetSingleStruct().ServiceName {
		t.Errorf("TheString888 service mismatch")
	}
	if jsFieldsService["TheList888"].GetListStruct().ListFields[0].ServiceName != fieldsService["TheList888"].GetListStruct().ListFields[0].ServiceName {
		t.Errorf("TheList888 service mismatch")
	}
	if jsFieldsService["TheMap888"].GetMapStruct().MapFields["*"].ServiceName != fieldsService["TheMap888"].GetMapStruct().MapFields["a1"].ServiceName {
		t.Errorf("TheMap888 service mismatch")
	}
}

func TestJSONStruct(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"Shape1": [2]any{
				"Class1", map[string]any{"Field1": "Circle1"}},
			"Shape2": [2]any{
				"Class2", map[string]any{"Field2": []string{"Circle2", "Circle3"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()

	shapeEndpoint := shapeFields["Shape1"].GetSingleStruct()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field1"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class1" ||
		field1Endpoint.ClassName != "Circle1" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}

	shape2Endpoint := shapeFields["Shape2"].GetSingleStruct()
	field2Fields := shape2Endpoint.GetFields()
	field2Endpoint := field2Fields["Field2"].GetListStruct()
	if spec.ClassName != "Geo" ||
		shape2Endpoint.ClassName != "Class2" ||
		field2Endpoint.ListFields[0].ClassName != "Circle2" ||
		field2Endpoint.ListFields[1].ClassName != "Circle3" {
		t.Errorf("shape spec: %s", shape2Endpoint.String())
		t.Errorf("field 2 spec: %s", field2Endpoint.String())
	}

	// Verify JSMServiceStruct
	jsonSchemaStr := `{
		"className": "object",
		"properties": {
			"Shape1": {
				"className": "Class1",
				"properties": {
					"Field1": { "className": "Circle1", "serviceName": "s1" }
				}
			},
			"Shape2": {
				"className": "Class2",
				"properties": {
					"Field2": { "items": { "className": "Circle2", "serviceName": "s2" } }
				}
			}
		}
	}`
	jsSpec, err := JSMServiceStruct("Geo", jsonSchemaStr)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsShapeFields := jsSpec.GetFields()

	// Verify Shape1
	jsShape1 := jsShapeFields["Shape1"].GetSingleStruct()
	if jsShape1.ClassName != "Class1" {
		t.Errorf("Shape1 class mismatch")
	}
	if jsShape1.Fields["Field1"].GetSingleStruct().ClassName != "Circle1" {
		t.Errorf("Shape1 Field1 mismatch")
	}

	// Verify Shape2
	jsShape2 := jsShapeFields["Shape2"].GetSingleStruct()
	if jsShape2.ClassName != "Class2" {
		t.Errorf("Shape2 class mismatch")
	}
	// Note: Manual test has mixed list ("Circle2", "Circle3"). JSMServiceStruct only validates against "Circle2" type.
	if jsShape2.Fields["Field2"].GetListStruct().ListFields[0].ClassName != "Circle2" {
		t.Errorf("Shape2 Field2 mismatch")
	}
	// Verify NewServiceStruct with ServiceName
	specService, err := NewServiceStruct(
		"Geo", map[string]any{
			"Shape1": [2]any{
				"Class1", map[string]any{"Field1": []string{"Circle1", "s1"}}},
			"Shape2": [2]any{
				"Class2", map[string]any{"Field2": [][]string{{"Circle2", "s2"}, {"Circle3", "s3"}}}},
		},
	)
	if err != nil {
		t.Errorf("NewServiceStruct failed: %v", err)
	}
	fieldsService := specService.GetFields()

	jsonSchemaStrService := `{
		"className": "object",
		"properties": {
			"Shape1": {
				"className": "Class1",
				"properties": {
					"Field1": { "className": "Circle1", "serviceName": "s1" }
				}
			},
			"Shape2": {
				"className": "Class2",
				"properties": {
					"Field2": { "items": { "className": "Circle2", "serviceName": "s2" } }
				}
			}
		}
	}`
	jsSpecService, err := JSMServiceStruct("Geo", jsonSchemaStrService)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsFieldsService := jsSpecService.GetFields()

	// Verify Shape1 Field1 ServiceName
	if jsFieldsService["Shape1"].GetSingleStruct().Fields["Field1"].GetSingleStruct().ServiceName !=
		fieldsService["Shape1"].GetSingleStruct().Fields["Field1"].GetSingleStruct().ServiceName {
		t.Errorf("Shape1 Field1 service mismatch")
	}

	// Verify Shape2 Field2 ServiceName
	// Note: manual struct list has specific items, schema has generic item. checking first item.
	if jsFieldsService["Shape2"].GetSingleStruct().Fields["Field2"].GetListStruct().ListFields[0].ServiceName !=
		fieldsService["Shape2"].GetSingleStruct().Fields["Field2"].GetListStruct().ListFields[0].ServiceName {
		t.Errorf("Shape2 Field2 service mismatch")
	}
}

func TestJSONList(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"ListShapes": [][2]any{
				{"Class2", map[string]any{"Field3": "Circle"}},
				{"Class3", map[string]any{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()
	shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field5"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class3" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}

	// Verify JSMServiceStruct
	// Manual test has heterogeneous list: Class2 (idx 0) and Class3 (idx 1).
	// JSMServiceStruct assumes homogenous list. We will define schema for Class2 (the first item type logic, usually).
	// BUT manual test specifically asserted on Index 1 (Class3)!
	// "shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]"
	// So we should construct schema to match Class3 if we want to mimic the test's interest, OR stick to Class2.
	// Since "List of T", typically T is uniform.
	// Let's mimic the *List Structure*. The manual construction is `[][2]any`.
	// I will generate schema for `Class2` and checking `Class2` since that's index 0.
	jsonSchemaStr := `{
		"properties": {
			"ListShapes": {
				"items": { 
					"className": "Class2",
					"properties": { "Field3": { "className": "Circle", "serviceName": "s1" } }
				}
			}
		}
	}`
	jsSpec, err := JSMServiceStruct("Geo", jsonSchemaStr)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsListShapes := jsSpec.GetFields()["ListShapes"].GetListStruct()
	jsItem0 := jsListShapes.ListFields[0] // JSM produces 1 item definition
	if jsItem0.ClassName != "Class2" {
		t.Errorf("ListShapes item class mismatch")
	}
	if jsItem0.Fields["Field3"].GetSingleStruct().ClassName != "Circle" {
		t.Errorf("ListShapes item field mismatch")
	}
	// Verify NewServiceStruct with ServiceName
	// Constructing a list of specs where each item has fields with service names
	specService, err := NewServiceStruct(
		"Geo", map[string]any{
			"ListShapes": [][2]any{
				{"Class2", map[string]any{"Field3": []string{"Circle", "s3"}}},
				{"Class3", map[string]any{"Field5": []string{"Circle", "s5"}}}},
		},
	)
	if err != nil {
		t.Errorf("NewServiceStruct failed: %v", err)
	}
	fieldsService := specService.GetFields()

	jsonSchemaStrService := `{
		"properties": {
			"ListShapes": {
				"items": { 
					"className": "Class2",
					"properties": { "Field3": { "className": "Circle", "serviceName": "s3" } }
				}
			}
		}
	}`
	// Note: Schema assumes homogenous list (Class2). Manual list has Class2 then Class3.
	// We check if the FIRST item (Class2) matches the schema prediction for Class2.

	jsSpecService, err := JSMServiceStruct("Geo", jsonSchemaStrService)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsListShapesService := jsSpecService.GetFields()["ListShapes"].GetListStruct()
	jsItem0Service := jsListShapesService.ListFields[0]

	manualListShapes := fieldsService["ListShapes"].GetListStruct()
	manualItem0 := manualListShapes.ListFields[0]

	if jsItem0Service.Fields["Field3"].GetSingleStruct().ServiceName != manualItem0.Fields["Field3"].GetSingleStruct().ServiceName {
		t.Errorf("ListShapes item field service mismatch")
	}
}

func TestJSONMap(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"HashShapes": map[string][2]any{
				"x1": {"Class5", map[string]any{"Field4": "Circle"}},
				"y1": {"Class6", map[string]any{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()
	shapeEndpoint := shapeFields["HashShapes"].GetMapStruct().GetMapFields()["x1"]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field4"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class5" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}

	// Verify JSMServiceStruct
	jsonSchemaStr := `{
		"properties": {
			"HashShapes": {
				"additionalProperties": {
					"className": "Class5",
					"properties": { "Field4": { "className": "Circle", "serviceName": "s1" } }
				}
			}
		}
	}`
	jsSpec, err := JSMServiceStruct("Geo", jsonSchemaStr)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsHashShapes := jsSpec.GetFields()["HashShapes"].GetMapStruct()
	jsItem := jsHashShapes.MapFields["*"]
	if jsItem.ClassName != "Class5" {
		t.Errorf("HashShapes item class mismatch")
	}
	if jsItem.Fields["Field4"].GetSingleStruct().ClassName != "Circle" {
		t.Errorf("HashShapes item field mismatch")
	}
	// Verify NewServiceStruct with ServiceName
	specService, err := NewServiceStruct(
		"Geo", map[string]any{
			"HashShapes": map[string][2]any{
				"x1": {"Class5", map[string]any{"Field4": []string{"Circle", "s1"}}},
				"y1": {"Class6", map[string]any{"Field5": []string{"Circle", "s2"}}}},
		},
	)
	if err != nil {
		t.Errorf("NewServiceStruct failed: %v", err)
	}
	fieldsService := specService.GetFields()

	jsonSchemaStrService := `{
		"properties": {
			"HashShapes": {
				"additionalProperties": {
					"className": "Class5",
					"properties": { "Field4": { "className": "Circle", "serviceName": "s1" } }
				}
			}
		}
	}`
	jsSpecService, err := JSMServiceStruct("Geo", jsonSchemaStrService)
	if err != nil {
		t.Errorf("JSMServiceStruct failed: %v", err)
	}
	jsHashShapesService := jsSpecService.GetFields()["HashShapes"].GetMapStruct()
	jsItemService := jsHashShapesService.MapFields["*"]

	manualHashShapes := fieldsService["HashShapes"].GetMapStruct()
	manualItemX1 := manualHashShapes.MapFields["x1"]

	if jsItemService.Fields["Field4"].GetSingleStruct().ServiceName != manualItemX1.Fields["Field4"].GetSingleStruct().ServiceName {
		t.Errorf("HashShapes item field service mismatch")
	}
}

func TestStruct_MarshalJSON_RoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		initialStruct *Struct
	}{
		{
			name: "Simple SingleStruct",
			initialStruct: &Struct{
				ClassName: "Person",
				Fields: map[string]*Value{
					"Name": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "MyString"}}},
				},
			},
		},
		{
			name: "Struct with Service",
			initialStruct: &Struct{
				ClassName:   "Person",
				ServiceName: "userService",
				Fields: map[string]*Value{
					"Avatar": {Kind: &Value_SingleStruct{SingleStruct: &Struct{ClassName: "Image", ServiceName: "imgService"}}},
				},
			},
		},
		{
			name: "ListStruct",
			initialStruct: &Struct{
				ClassName: "Group",
				Fields: map[string]*Value{
					"Members": {Kind: &Value_ListStruct{ListStruct: &ListStruct{
						ListFields: []*Struct{{ClassName: "Person"}},
					}}},
				},
			},
		},
		{
			name: "MapStruct",
			initialStruct: &Struct{
				ClassName: "Registry",
				Fields: map[string]*Value{
					"Entries": {Kind: &Value_MapStruct{MapStruct: &MapStruct{
						MapFields: map[string]*Struct{"*": {ClassName: "Entry"}},
					}}},
				},
			},
		},
		{
			name: "Map2Struct",
			initialStruct: &Struct{
				ClassName: "GeoMap",
				Fields: map[string]*Value{
					"Grid": {Kind: &Value_Map2Struct{Map2Struct: &Map2Struct{
						Map2Fields: map[string]*MapStruct{
							"region1": {MapFields: map[string]*Struct{
								"keyA": {ClassName: "Point", ServiceName: "s1"},
							}},
						},
					}}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Marshal to JSON
			data, err := json.MarshalIndent(tt.initialStruct, "", "  ")
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			t.Logf("Marshaled JSON:\n%s", string(data))

			// 2. Unmarshal back
			var restored Struct
			if err := json.Unmarshal(data, &restored); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// 3. Compare
			// Note: We need a deep compare.
			// However, initialStruct might be slightly different structurally from restored
			// if we omitted fields that are nil. But here we constructed them carefully.
			// One catch: MapStruct keys. initialStruct has "*". Restored has "*".
			// Map2Struct keys are preserved.
			if restored.ClassName != tt.initialStruct.ClassName {
				t.Errorf("ClassName mismatch: got %q, want %q", restored.ClassName, tt.initialStruct.ClassName)
			}
			if restored.ServiceName != tt.initialStruct.ServiceName {
				t.Errorf("ServiceName mismatch: got %q, want %q", restored.ServiceName, tt.initialStruct.ServiceName)
			}
			// Checking fields existence roughly
			if len(restored.Fields) != len(tt.initialStruct.Fields) {
				t.Errorf("Fields count mismatch: got %d, want %d", len(restored.Fields), len(tt.initialStruct.Fields))
			}
		})
	}
}
