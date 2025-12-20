package schema

// DeriveStructWithoutServices creates a new Struct from an old Struct (typically generated
// from NewServiceStruct) with all ServiceName fields emptied recursively throughout the
// entire structure.
//
// This function:
//   - Clears the ServiceName field at the root level
//   - Recursively processes all nested Fields to clear ServiceName in nested Structs
//   - Handles all Value types: SingleStruct, ListStruct, MapStruct, and Map2Struct
//   - Returns a deep copy with only ClassName and Fields preserved (ServiceNames removed)
//
// Example:
//
//	oldStruct, _ := NewServiceStruct("provider", map[string]any{
//	    "Database": []string{"db", "dbService"},
//	    "Cache": []string{"redis", "cacheService"},
//	})
//	newStruct := DeriveStructWithoutServices(oldStruct)
//	// newStruct will have ClassName="provider" with empty ServiceName
//	// Nested Database and Cache will also have their ServiceNames cleared
func DeriveStructWithoutServices(old *Struct) *Struct {
	if old == nil {
		return nil
	}

	// Create new struct with ClassName but empty ServiceName
	newStruct := &Struct{
		ClassName: old.ClassName,
		// ServiceName is intentionally left empty
	}

	// Recursively process all fields
	if old.Fields != nil {
		newStruct.Fields = make(map[string]*Value, len(old.Fields))
		for key, value := range old.Fields {
			newStruct.Fields[key] = deriveValueWithoutServices(value)
		}
	}

	return newStruct
}

// deriveValueWithoutServices recursively processes a Value and clears all ServiceNames
// from any nested Structs it contains.
func deriveValueWithoutServices(old *Value) *Value {
	if old == nil {
		return nil
	}

	newValue := &Value{}

	switch kind := old.Kind.(type) {
	case *Value_SingleStruct:
		newValue.Kind = &Value_SingleStruct{
			SingleStruct: DeriveStructWithoutServices(kind.SingleStruct),
		}

	case *Value_ListStruct:
		newValue.Kind = &Value_ListStruct{
			ListStruct: deriveListStructWithoutServices(kind.ListStruct),
		}

	case *Value_MapStruct:
		newValue.Kind = &Value_MapStruct{
			MapStruct: deriveMapStructWithoutServices(kind.MapStruct),
		}

	case *Value_Map2Struct:
		newValue.Kind = &Value_Map2Struct{
			Map2Struct: deriveMap2StructWithoutServices(kind.Map2Struct),
		}
	}

	return newValue
}

// deriveListStructWithoutServices processes a ListStruct and clears ServiceNames from all Structs.
func deriveListStructWithoutServices(old *ListStruct) *ListStruct {
	if old == nil {
		return nil
	}

	newList := &ListStruct{
		ListFields: make([]*Struct, len(old.ListFields)),
	}

	for i, s := range old.ListFields {
		newList.ListFields[i] = DeriveStructWithoutServices(s)
	}

	return newList
}

// deriveMapStructWithoutServices processes a MapStruct and clears ServiceNames from all Structs.
func deriveMapStructWithoutServices(old *MapStruct) *MapStruct {
	if old == nil {
		return nil
	}

	newMap := &MapStruct{
		MapFields: make(map[string]*Struct, len(old.MapFields)),
	}

	for key, s := range old.MapFields {
		newMap.MapFields[key] = DeriveStructWithoutServices(s)
	}

	return newMap
}

// deriveMap2StructWithoutServices processes a Map2Struct and clears ServiceNames from all Structs.
func deriveMap2StructWithoutServices(old *Map2Struct) *Map2Struct {
	if old == nil {
		return nil
	}

	newMap2 := &Map2Struct{
		Map2Fields: make(map[string]*MapStruct, len(old.Map2Fields)),
	}

	for key, ms := range old.Map2Fields {
		newMap2.Map2Fields[key] = deriveMapStructWithoutServices(ms)
	}

	return newMap2
}
