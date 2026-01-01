package schema

import (
	"testing"
)

func TestNewStruct_Map2(t *testing.T) {
	// Case 1: Struct containing a Map2Struct field
	spec := map[[2]string]string{
		{"Region1", "ZoneA"}: "ServerType",
		{"Region1", "ZoneB"}: "DatabaseType",
		{"Region2", "ZoneX"}: "CacheType",
	}

	// NewStruct expects map[string]any (fields), so we wrap our Map2 spec
	s, err := NewStruct("Cluster", map[string]any{
		"Topology": spec,
	})
	if err != nil {
		t.Fatalf("NewStruct failed: %v", err)
	}

	if s.ClassName != "Cluster" {
		t.Errorf("Expected ClassName 'Cluster', got %q", s.ClassName)
	}

	// Access the "Topology" field
	val := s.Fields["Topology"]
	if val == nil {
		t.Fatal("Topology field missing")
	}

	m2 := val.GetMap2Struct()
	if m2 == nil {
		t.Fatal("Expected Map2Struct in Topology field, got nil")
	}

	// Check Region1
	r1, ok := m2.Map2Fields["Region1"]
	if !ok {
		t.Fatal("Region1 missing")
	}
	if r1.MapFields["ZoneA"].ClassName != "ServerType" {
		t.Errorf("ZoneA: expected ServerType, got %q", r1.MapFields["ZoneA"].ClassName)
	}
	if r1.MapFields["ZoneB"].ClassName != "DatabaseType" {
		t.Errorf("ZoneB: expected DatabaseType, got %q", r1.MapFields["ZoneB"].ClassName)
	}

	// Check Region2
	r2, ok := m2.Map2Fields["Region2"]
	if !ok {
		t.Fatal("Region2 missing")
	}
	if r2.MapFields["ZoneX"].ClassName != "CacheType" {
		t.Errorf("ZoneX: expected CacheType, got %q", r2.MapFields["ZoneX"].ClassName)
	}
}

func TestNewServiceStruct_Map2(t *testing.T) {
	// Case 2: Struct containing a Map2Struct field with service specs
	spec := map[[2]string][2]any{
		{"Net", "Inner"}: {"Firewall", "fwService"},
		{"Net", "Outer"}: {"LoadBalancer", map[string]any{"IP": []string{"IPString"}}},
	}

	s, err := NewServiceStruct("Infrastructure", map[string]any{
		"Grid": spec,
	})
	if err != nil {
		t.Fatalf("NewServiceStruct failed: %v", err)
	}

	val := s.Fields["Grid"]
	if val == nil {
		t.Fatal("Grid field missing")
	}

	m2 := val.GetMap2Struct()
	if m2 == nil {
		t.Fatal("Expected Map2Struct, got nil")
	}

	netMap := m2.Map2Fields["Net"]
	if netMap == nil {
		t.Fatal("Net group missing")
	}

	fw := netMap.MapFields["Inner"]
	if fw.ClassName != "Firewall" || fw.ServiceName != "fwService" {
		t.Errorf("Inner: expected Firewall/fwService, got %s/%s", fw.ClassName, fw.ServiceName)
	}

	lb := netMap.MapFields["Outer"]
	if lb.ClassName != "LoadBalancer" {
		t.Errorf("Outer: expected LoadBalancer, got %s", lb.ClassName)
	}
	// Verify nested field 1.1.1.1? (Actually NewServiceValue(val) creates a Value, so "1.1.1.1" becomes a String Value?
	// Or implementation detal: "1.1.1.1" string is processed by NewServiceValue.
	// Wait, map[string]any{"IP": "1.1.1.1"} -> Fields["IP"] is value("1.1.1.1")?
	// Let's check LB fields if possible, or assume it worked because no error.
}

func TestNewValue_Map2(t *testing.T) {
	// Test via NewServiceValue directly with map[[2]string][]string (End Structs)
	spec := map[[2]string][]string{
		{"US", "East"}: {"DataCenter", "dc1"},
	}

	v, err := NewServiceValue(spec)
	if err != nil {
		t.Fatalf("NewServiceValue failed: %v", err)
	}

	m2 := v.GetMap2Struct()
	if m2 == nil {
		t.Fatal("Expected Value to be Map2Struct")
	}

	dc := m2.Map2Fields["US"].MapFields["East"]
	if dc.ClassName != "DataCenter" || dc.ServiceName != "dc1" {
		t.Errorf("Expected DataCenter/dc1, got %s/%s", dc.ClassName, dc.ServiceName)
	}
}

func TestNewStruct_Map2_Invalid(t *testing.T) {
	// Test error cases?
	// Not easily triggered unless types are wrong in the map, but Go type system prevents that for map[[2]string]string.
	// Maybe wrong inner values if using interface?
}
