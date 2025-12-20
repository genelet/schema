package schema

import (
	"testing"
)

func TestDeriveStructWithoutServices_Nil(t *testing.T) {
	result := DeriveStructWithoutServices(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestDeriveStructWithoutServices_SimpleStruct(t *testing.T) {
	// Create a simple struct with ServiceName
	oldStruct, err := NewServiceStruct("provider", "providerService")
	if err != nil {
		t.Fatal(err)
	}

	// Derive new struct without services
	newStruct := DeriveStructWithoutServices(oldStruct)

	// Verify ClassName is preserved
	if newStruct.ClassName != "provider" {
		t.Errorf("expected ClassName 'provider', got '%s'", newStruct.ClassName)
	}

	// Verify ServiceName is empty
	if newStruct.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got '%s'", newStruct.ServiceName)
	}

	// Verify Fields is nil (since original had no fields)
	if newStruct.Fields != nil {
		t.Errorf("expected nil Fields, got %v", newStruct.Fields)
	}
}

func TestDeriveStructWithoutServices_WithEndStructFields(t *testing.T) {
	// Create struct with end-node fields (containing ServiceNames)
	oldStruct, err := NewServiceStruct("Geo", map[string]any{
		"EndString": []string{"Circle1", "service1"},
		"EndList": [][]string{
			{"Circle2", "service2"},
			{"Circle3", "service3"},
		},
		"EndMap": map[string][]string{
			"key1": {"Circle2", "service2"},
			"key2": {"Circle3", "service3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Derive new struct without services
	newStruct := DeriveStructWithoutServices(oldStruct)

	// Verify root ClassName is preserved
	if newStruct.ClassName != "Geo" {
		t.Errorf("expected ClassName 'Geo', got '%s'", newStruct.ClassName)
	}

	// Verify root ServiceName is empty
	if newStruct.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got '%s'", newStruct.ServiceName)
	}

	// Verify EndString field
	endString := newStruct.Fields["EndString"].GetSingleStruct()
	if endString.ClassName != "Circle1" {
		t.Errorf("expected EndString ClassName 'Circle1', got '%s'", endString.ClassName)
	}
	if endString.ServiceName != "" {
		t.Errorf("expected EndString ServiceName to be empty, got '%s'", endString.ServiceName)
	}

	// Verify EndList field
	endList := newStruct.Fields["EndList"].GetListStruct()
	if len(endList.ListFields) != 2 {
		t.Fatalf("expected 2 items in EndList, got %d", len(endList.ListFields))
	}
	if endList.ListFields[0].ClassName != "Circle2" {
		t.Errorf("expected EndList[0] ClassName 'Circle2', got '%s'", endList.ListFields[0].ClassName)
	}
	if endList.ListFields[0].ServiceName != "" {
		t.Errorf("expected EndList[0] ServiceName to be empty, got '%s'", endList.ListFields[0].ServiceName)
	}
	if endList.ListFields[1].ClassName != "Circle3" {
		t.Errorf("expected EndList[1] ClassName 'Circle3', got '%s'", endList.ListFields[1].ClassName)
	}
	if endList.ListFields[1].ServiceName != "" {
		t.Errorf("expected EndList[1] ServiceName to be empty, got '%s'", endList.ListFields[1].ServiceName)
	}

	// Verify EndMap field
	endMap := newStruct.Fields["EndMap"].GetMapStruct()
	if len(endMap.MapFields) != 2 {
		t.Fatalf("expected 2 items in EndMap, got %d", len(endMap.MapFields))
	}
	if endMap.MapFields["key1"].ClassName != "Circle2" {
		t.Errorf("expected EndMap[key1] ClassName 'Circle2', got '%s'", endMap.MapFields["key1"].ClassName)
	}
	if endMap.MapFields["key1"].ServiceName != "" {
		t.Errorf("expected EndMap[key1] ServiceName to be empty, got '%s'", endMap.MapFields["key1"].ServiceName)
	}
	if endMap.MapFields["key2"].ClassName != "Circle3" {
		t.Errorf("expected EndMap[key2] ClassName 'Circle3', got '%s'", endMap.MapFields["key2"].ClassName)
	}
	if endMap.MapFields["key2"].ServiceName != "" {
		t.Errorf("expected EndMap[key2] ServiceName to be empty, got '%s'", endMap.MapFields["key2"].ServiceName)
	}
}

func TestDeriveStructWithoutServices_NestedStructs(t *testing.T) {
	// Create struct with nested middle structures
	oldStruct, err := NewServiceStruct("Geo", map[string]any{
		"TheMiddle": [2]any{"Circle", map[string]any{
			"EndString": []string{"Circle1", "service1"},
			"EndList": [][]string{
				{"Circle2", "service2"},
				{"Circle3", "service3"},
			},
		}},
		"TheMiddleList": [][2]any{
			{"CircleStruct1", map[string]any{
				"EndString": []string{"Circle1", "service1"},
			}},
			{"CircleStruct2", map[string]any{
				"EndString": []string{"Circle4", "service4"},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Derive new struct without services
	newStruct := DeriveStructWithoutServices(oldStruct)

	// Verify root
	if newStruct.ClassName != "Geo" {
		t.Errorf("expected ClassName 'Geo', got '%s'", newStruct.ClassName)
	}
	if newStruct.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got '%s'", newStruct.ServiceName)
	}

	// Verify TheMiddle (nested SingleStruct)
	theMiddle := newStruct.Fields["TheMiddle"].GetSingleStruct()
	if theMiddle.ClassName != "Circle" {
		t.Errorf("expected TheMiddle ClassName 'Circle', got '%s'", theMiddle.ClassName)
	}
	if theMiddle.ServiceName != "" {
		t.Errorf("expected TheMiddle ServiceName to be empty, got '%s'", theMiddle.ServiceName)
	}

	// Verify nested EndString inside TheMiddle
	endString := theMiddle.Fields["EndString"].GetSingleStruct()
	if endString.ClassName != "Circle1" {
		t.Errorf("expected nested EndString ClassName 'Circle1', got '%s'", endString.ClassName)
	}
	if endString.ServiceName != "" {
		t.Errorf("expected nested EndString ServiceName to be empty, got '%s'", endString.ServiceName)
	}

	// Verify nested EndList inside TheMiddle
	endList := theMiddle.Fields["EndList"].GetListStruct()
	if len(endList.ListFields) != 2 {
		t.Fatalf("expected 2 items in nested EndList, got %d", len(endList.ListFields))
	}
	for i, s := range endList.ListFields {
		if s.ServiceName != "" {
			t.Errorf("expected nested EndList[%d] ServiceName to be empty, got '%s'", i, s.ServiceName)
		}
	}

	// Verify TheMiddleList (nested ListStruct)
	theMiddleList := newStruct.Fields["TheMiddleList"].GetListStruct()
	if len(theMiddleList.ListFields) != 2 {
		t.Fatalf("expected 2 items in TheMiddleList, got %d", len(theMiddleList.ListFields))
	}

	for i, circleStruct := range theMiddleList.ListFields {
		if circleStruct.ServiceName != "" {
			t.Errorf("expected TheMiddleList[%d] ServiceName to be empty, got '%s'", i, circleStruct.ServiceName)
		}
		// Check nested EndString
		nestedEnd := circleStruct.Fields["EndString"].GetSingleStruct()
		if nestedEnd.ServiceName != "" {
			t.Errorf("expected TheMiddleList[%d].EndString ServiceName to be empty, got '%s'", i, nestedEnd.ServiceName)
		}
	}
}

func TestDeriveStructWithoutServices_Map2Struct(t *testing.T) {
	// Create struct with Map2Struct field
	oldStruct, err := NewServiceStruct("Container", map[string]any{
		"Nested2DMap": map[[2]string][]string{
			{"region1", "key1"}: {"Service1", "service1"},
			{"region1", "key2"}: {"Service2", "service2"},
			{"region2", "key1"}: {"Service3", "service3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Derive new struct without services
	newStruct := DeriveStructWithoutServices(oldStruct)

	// Verify root
	if newStruct.ClassName != "Container" {
		t.Errorf("expected ClassName 'Container', got '%s'", newStruct.ClassName)
	}
	if newStruct.ServiceName != "" {
		t.Errorf("expected empty ServiceName, got '%s'", newStruct.ServiceName)
	}

	// Verify Nested2DMap
	map2 := newStruct.Fields["Nested2DMap"].GetMap2Struct()
	if len(map2.Map2Fields) != 2 {
		t.Fatalf("expected 2 regions in Map2Fields, got %d", len(map2.Map2Fields))
	}

	// Check region1
	region1 := map2.Map2Fields["region1"]
	if region1 == nil {
		t.Fatal("expected region1 to exist")
	}
	if len(region1.MapFields) != 2 {
		t.Fatalf("expected 2 items in region1, got %d", len(region1.MapFields))
	}

	// Verify all ServiceNames are cleared
	for regionKey, mapStruct := range map2.Map2Fields {
		for innerKey, s := range mapStruct.MapFields {
			if s.ServiceName != "" {
				t.Errorf("expected Map2Fields[%s][%s] ServiceName to be empty, got '%s'", regionKey, innerKey, s.ServiceName)
			}
		}
	}

	// Verify ClassNames are preserved
	if map2.Map2Fields["region1"].MapFields["key1"].ClassName != "Service1" {
		t.Errorf("expected ClassName 'Service1', got '%s'", map2.Map2Fields["region1"].MapFields["key1"].ClassName)
	}
	if map2.Map2Fields["region2"].MapFields["key1"].ClassName != "Service3" {
		t.Errorf("expected ClassName 'Service3', got '%s'", map2.Map2Fields["region2"].MapFields["key1"].ClassName)
	}
}

func TestDeriveStructWithoutServices_DeeplyNested(t *testing.T) {
	// Create a deeply nested structure
	oldStruct, err := NewServiceStruct("Root", map[string]any{
		"Level1": [2]any{"L1Class", map[string]any{
			"Level2": [2]any{"L2Class", map[string]any{
				"Level3": [2]any{"L3Class", map[string]any{
					"End": []string{"EndClass", "endService"},
				}},
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Derive new struct without services
	newStruct := DeriveStructWithoutServices(oldStruct)

	// Navigate to the deepest level
	level1 := newStruct.Fields["Level1"].GetSingleStruct()
	level2 := level1.Fields["Level2"].GetSingleStruct()
	level3 := level2.Fields["Level3"].GetSingleStruct()
	end := level3.Fields["End"].GetSingleStruct()

	// Verify all levels have empty ServiceName
	if newStruct.ServiceName != "" {
		t.Errorf("expected Root ServiceName to be empty, got '%s'", newStruct.ServiceName)
	}
	if level1.ServiceName != "" {
		t.Errorf("expected Level1 ServiceName to be empty, got '%s'", level1.ServiceName)
	}
	if level2.ServiceName != "" {
		t.Errorf("expected Level2 ServiceName to be empty, got '%s'", level2.ServiceName)
	}
	if level3.ServiceName != "" {
		t.Errorf("expected Level3 ServiceName to be empty, got '%s'", level3.ServiceName)
	}
	if end.ServiceName != "" {
		t.Errorf("expected End ServiceName to be empty, got '%s'", end.ServiceName)
	}

	// Verify all ClassNames are preserved
	if level1.ClassName != "L1Class" {
		t.Errorf("expected Level1 ClassName 'L1Class', got '%s'", level1.ClassName)
	}
	if level2.ClassName != "L2Class" {
		t.Errorf("expected Level2 ClassName 'L2Class', got '%s'", level2.ClassName)
	}
	if level3.ClassName != "L3Class" {
		t.Errorf("expected Level3 ClassName 'L3Class', got '%s'", level3.ClassName)
	}
	if end.ClassName != "EndClass" {
		t.Errorf("expected End ClassName 'EndClass', got '%s'", end.ClassName)
	}
}
