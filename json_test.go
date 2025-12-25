package schema

import (
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
		"type": "object",
		"properties": {
			"TheString888": { "type": "Circle" },
			"TheString":    { "type": "Circle" },
			"TheList888":   { "type": "array", "items": { "type": "CircleClass1" } },
			"TheList":      { "type": "array", "items": { "type": "CircleClass1" } },
			"TheMap888":    { "type": "object", "additionalProperties": { "type": "CircleClass1" } },
			"TheMap":       { "type": "object", "additionalProperties": { "type": "CircleClass1" } }
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
		"type": "object",
		"properties": {
			"Shape1": {
				"type": "Class1",
				"properties": {
					"Field1": { "type": "Circle1" }
				}
			},
			"Shape2": {
				"type": "Class2",
				"properties": {
					"Field2": { "type": "array", "items": { "type": "Circle2" } }
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
		"type": "object",
		"properties": {
			"ListShapes": {
				"type": "array",
				"items": { 
					"type": "Class2",
					"properties": { "Field3": { "type": "Circle" } }
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
		"type": "object",
		"properties": {
			"HashShapes": {
				"type": "object",
				"additionalProperties": {
					"type": "Class5",
					"properties": { "Field4": { "type": "Circle" } }
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
}
