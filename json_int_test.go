package schema

import (
	"testing"
)

func TestJSMServiceStruct(t *testing.T) {
	tests := []struct {
		name       string
		className  string
		jsonSchema string
		check      func(*testing.T, *Struct)
	}{
		{
			name:      "Simple Object",
			className: "Person",
			jsonSchema: `{
				"type": "object",
				"properties": {
					"name": { "type": "string" },
					"age": { "type": "integer" }
				}
			}`,
			check: func(t *testing.T, s *Struct) {
				if s.ClassName != "Person" {
					t.Errorf("expected Person, got %s", s.ClassName)
				}
				if len(s.Fields) != 2 {
					t.Errorf("expected 2 fields, got %d", len(s.Fields))
				}
				if s.Fields["name"].GetSingleStruct().ClassName != "string" {
					t.Errorf("expected name to be string")
				}
				if s.Fields["age"].GetSingleStruct().ClassName != "integer" {
					t.Errorf("expected age to be integer")
				}
			},
		},
		{
			name:      "Nested Object",
			className: "Container",
			jsonSchema: `{
				"type": "object",
				"properties": {
					"child": {
						"type": "object",
						"properties": {
							"val": { "type": "string" }
						}
					}
				}
			}`,
			check: func(t *testing.T, s *Struct) {
				child := s.Fields["child"].GetSingleStruct()
				if child == nil {
					t.Fatal("child is nil")
				}
				if child.Fields["val"].GetSingleStruct().ClassName != "string" {
					t.Errorf("expected val to be string")
				}
			},
		},
		{
			name:      "Array of Strings",
			className: "TagList",
			jsonSchema: `{
				"type": "object",
				"properties": {
					"tags": {
						"type": "array",
						"items": { "type": "string" }
					}
				}
			}`,
			check: func(t *testing.T, s *Struct) {
				tags := s.Fields["tags"].GetListStruct()
				if tags == nil {
					t.Fatal("tags is nil")
				}
				// items should be a SINGLE struct describing the type of elements
				if len(tags.ListFields) != 1 {
					t.Errorf("expected 1 item in ListStruct describing the type, got %d", len(tags.ListFields))
				}
				if tags.ListFields[0].ClassName != "string" {
					t.Errorf("expected item type string")
				}
			},
		},
		{
			name:      "Object with Map Field",
			className: "ContainerWithMap",
			jsonSchema: `{
				"type": "object",
				"properties": {
					"myMap": {
						"type": "object",
						"additionalProperties": {
							"type": "integer"
						}
					}
				}
			}`,
			check: func(t *testing.T, s *Struct) {
				val := s.Fields["myMap"]
				if val == nil {
					t.Fatal("field myMap is nil")
				}
				ms := val.GetMapStruct()
				if ms == nil {
					t.Fatalf("expected MapStruct for myMap, got %v", val.Kind)
				}

				// Check for wildcard key
				elem := ms.MapFields["*"]
				if elem == nil {
					t.Fatal("expected wildcard '*' key in MapStruct")
				}
				if elem.ClassName != "integer" {
					t.Errorf("expected value type integer, got %s", elem.ClassName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := JSMServiceStruct(tt.className, tt.jsonSchema)
			if err != nil {
				t.Fatalf("JSMServiceStruct failed: %v", err)
			}
			tt.check(t, s)
		})
	}
}

func TestJSMServiceStructWithService(t *testing.T) {
	className := "MyClass"
	serviceName := "MyService"
	jsonSchemaStr := `{"type": "string", "serviceName": "MyService"}`

	s, err := JSMServiceStruct(className, jsonSchemaStr)
	if err != nil {
		t.Fatalf("JSMServiceStruct failed: %v", err)
	}

	if s.ClassName != className {
		t.Errorf("expected ClassName %q, got %q", className, s.ClassName)
	}
	if s.ServiceName != serviceName {
		t.Errorf("expected ServiceName %q, got %q", serviceName, s.ServiceName)
	}
}

func TestJSMServiceStruct_Map2(t *testing.T) {
	className := "Map2Container"
	jsonSchemaStr := `{
		"type": "object",
		"properties": {
			"myMap2": {
				"type": "object",
				"x-map2": true,
				"properties": {
					"region1": {
						"type": "object",
						"properties": {
							"key1": { "type": "ServiceA" },
							"key2": { "type": "ServiceB" }
						}
					},
					"region2": {
						"type": "object",
						"properties": {
							"key3": { "type": "ServiceC" }
						}
					}
				}
			}
		}
	}`

	s, err := JSMServiceStruct(className, jsonSchemaStr)
	if err != nil {
		t.Fatalf("JSMServiceStruct failed: %v", err)
	}

	// Navigate to myMap2
	map2Val := s.Fields["myMap2"]
	if map2Val == nil {
		t.Fatal("expected field myMap2")
	}
	map2 := map2Val.GetMap2Struct()
	if map2 == nil {
		t.Fatalf("expected Map2Struct, got %v", map2Val.Kind)
	}

	if len(map2.Map2Fields) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(map2.Map2Fields))
	}

	// Verify region1
	r1, ok := map2.Map2Fields["region1"]
	if !ok {
		t.Fatal("expected region1")
	}
	if len(r1.MapFields) != 2 {
		t.Fatalf("expected 2 keys in region1, got %d", len(r1.MapFields))
	}
	if r1.MapFields["key1"].ClassName != "ServiceA" {
		t.Errorf("expected ServiceA, got %s", r1.MapFields["key1"].ClassName)
	}
	if r1.MapFields["key2"].ClassName != "ServiceB" {
		t.Errorf("expected ServiceB, got %s", r1.MapFields["key2"].ClassName)
	}

	// Verify region2
	r2, ok := map2.Map2Fields["region2"]
	if !ok {
		t.Fatal("expected region2")
	}
	if len(r2.MapFields) != 1 {
		t.Fatalf("expected 1 key in region2, got %d", len(r2.MapFields))
	}
	if r2.MapFields["key3"].ClassName != "ServiceC" {
		t.Errorf("expected ServiceC, got %s", r2.MapFields["key3"].ClassName)
	}
}

func TestJSMStruct(t *testing.T) {
	className := "NoServiceClass"
	// Schema with serviceName
	jsonSchemaStr := `{
		"type": "object",
		"properties": {
			"f1": { "type": "string", "serviceName": "shouldBeGone" }
		}
	}`

	s, err := JSMStruct(className, jsonSchemaStr)
	if err != nil {
		t.Fatalf("JSMStruct failed: %v", err)
	}

	if s.ClassName != className {
		t.Errorf("expected %s, got %s", className, s.ClassName)
	}
	if s.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got %s", s.ServiceName)
	}

	f1 := s.Fields["f1"].GetSingleStruct()
	if f1.ClassName != "string" {
		t.Errorf("expected string class, got %s", f1.ClassName)
	}
	if f1.ServiceName != "" {
		t.Errorf("expected empty nested ServiceName, got %s", f1.ServiceName)
	}
}
