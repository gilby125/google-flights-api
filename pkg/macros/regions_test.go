package macros

import (
	"testing"
)

func TestExpandAirportTokens_BasicIATA(t *testing.T) {
	inputs := []string{"JFK", "lax", " ORD ", "jfk"} // mixed case, duplicates, whitespace

	airports, warnings, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Should be deduped and uppercased
	if len(airports) != 3 {
		t.Errorf("expected 3 airports, got %d: %v", len(airports), airports)
	}

	expected := map[string]bool{"JFK": true, "LAX": true, "ORD": true}
	for _, code := range airports {
		if !expected[code] {
			t.Errorf("unexpected airport code: %s", code)
		}
	}
}

func TestExpandAirportTokens_RegionExpansion(t *testing.T) {
	inputs := []string{"REGION:MIDDLE_EAST"}

	airports, warnings, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Should contain Middle East airports from Top100Airports
	// DXB, DOH, AUH based on our mapping
	if len(airports) == 0 {
		t.Error("expected some airports for MIDDLE_EAST region")
	}

	// Verify at least DXB is present
	found := false
	for _, code := range airports {
		if code == "DXB" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected DXB in MIDDLE_EAST region airports")
	}
}

func TestExpandAirportTokens_WorldRegion(t *testing.T) {
	inputs := []string{"REGION:WORLD"}

	airports, warnings, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Should contain all airports from Top100Airports (95 airports)
	if len(airports) < 90 {
		t.Errorf("expected at least 90 airports for WORLD region, got %d", len(airports))
	}
}

func TestExpandAirportTokens_MixedInputs(t *testing.T) {
	inputs := []string{"JFK", "REGION:MIDDLE_EAST", "DXB"} // DXB is in MIDDLE_EAST, should be deduped

	airports, _, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count DXB occurrences
	dxbCount := 0
	for _, code := range airports {
		if code == "DXB" {
			dxbCount++
		}
	}

	if dxbCount != 1 {
		t.Errorf("expected DXB to appear exactly once, appeared %d times", dxbCount)
	}
}

func TestExpandAirportTokens_InvalidCode(t *testing.T) {
	inputs := []string{"JFKX"} // 4 characters

	_, _, err := ExpandAirportTokens(inputs)
	if err == nil {
		t.Error("expected error for invalid airport code")
	}
}

func TestExpandAirportTokens_UnknownRegion(t *testing.T) {
	inputs := []string{"REGION:UNKNOWN"}

	_, _, err := ExpandAirportTokens(inputs)
	if err == nil {
		t.Error("expected error for unknown region")
	}
}

func TestValidateNoRegionTokens_NoTokens(t *testing.T) {
	err := ValidateNoRegionTokens("JFK", "LAX")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateNoRegionTokens_WithToken(t *testing.T) {
	err := ValidateNoRegionTokens("JFK", "REGION:EUROPE")
	if err == nil {
		t.Error("expected error for region token in single-route context")
	}
}

func TestGetRegionAirports_AllRegions(t *testing.T) {
	// ALL advertised regions should return non-nil (recognized), even if empty
	for _, region := range AllRegions() {
		airports := GetRegionAirports(region)
		if airports == nil {
			t.Errorf("expected non-nil airports for advertised region %s (should be recognized even if empty)", region)
		}
	}

	// Regions with known airports should have some
	regionsWithAirports := []string{
		RegionEurope,
		RegionNorthAmerica,
		RegionAsia,
		RegionOceania,
		RegionMiddleEast,
		RegionAfrica,
		RegionSouthAmerica,
		RegionWorld,
	}

	for _, region := range regionsWithAirports {
		airports := GetRegionAirports(region)
		if len(airports) == 0 {
			t.Errorf("expected some airports for region %s", region)
		}
	}
}

func TestExpandAirportTokens_EmptyRegion(t *testing.T) {
	// REGION:CARIBBEAN is advertised but has zero airports in Top100
	// It should NOT error - it should return empty with a warning
	inputs := []string{RegionCaribbean}

	airports, warnings, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("%s should be recognized even with zero airports, got error: %v", RegionCaribbean, err)
	}

	if len(airports) != 0 {
		t.Errorf("expected 0 airports for %s, got %d", RegionCaribbean, len(airports))
	}

	if len(warnings) == 0 {
		t.Fatalf("expected warning for empty region %s, got none", RegionCaribbean)
	}
}

func TestGetAllRegionInfo(t *testing.T) {
	info := GetAllRegionInfo()

	if len(info) != len(AllRegions()) {
		t.Errorf("expected %d regions, got %d", len(AllRegions()), len(info))
	}

	for _, r := range info {
		if r.Token == "" {
			t.Error("expected non-empty token")
		}
		if r.AirportCount > 0 && len(r.SampleAirports) == 0 {
			t.Errorf("expected sample airports for region %s with count %d", r.Token, r.AirportCount)
		}
	}
}

func TestIsRegionToken(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"REGION:EUROPE", true},
		{"region:europe", true},
		{"Region:Europe", true},
		{"JFK", false},
		{"GROUP:STAR_ALLIANCE", false},
		{"", false},
	}

	for _, tc := range tests {
		result := IsRegionToken(tc.input)
		if result != tc.expected {
			t.Errorf("IsRegionToken(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestExpandAirportTokens_EmptyInput(t *testing.T) {
	inputs := []string{"", "  ", "JFK"}

	airports, _, err := ExpandAirportTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(airports) != 1 {
		t.Errorf("expected 1 airport, got %d: %v", len(airports), airports)
	}
}
