package macros

import (
	"testing"
)

func TestExpandAirlineTokens_BasicIATA(t *testing.T) {
	inputs := []string{"UA", "aa", " DL ", "ua"} // mixed case, duplicates, whitespace

	codes, warnings, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Should be deduped and uppercased
	if len(codes) != 3 {
		t.Errorf("expected 3 airline codes, got %d: %v", len(codes), codes)
	}

	expected := map[string]bool{"UA": true, "AA": true, "DL": true}
	for _, code := range codes {
		if !expected[code] {
			t.Errorf("unexpected airline code: %s", code)
		}
	}
}

func TestExpandAirlineTokens_GroupExpansion(t *testing.T) {
	inputs := []string{"GROUP:LOW_COST"}

	codes, warnings, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Should contain low-cost carriers
	if len(codes) == 0 {
		t.Error("expected some airline codes for LOW_COST group")
	}

	// Verify FR (Ryanair) is present
	found := false
	for _, code := range codes {
		if code == "FR" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected FR (Ryanair) in LOW_COST group airlines")
	}
}

func TestExpandAirlineTokens_StarAlliance(t *testing.T) {
	inputs := []string{"GROUP:STAR_ALLIANCE"}

	codes, _, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain Star Alliance members
	if len(codes) < 10 {
		t.Errorf("expected at least 10 airlines for STAR_ALLIANCE, got %d", len(codes))
	}

	// Verify LH (Lufthansa) is present
	found := false
	for _, code := range codes {
		if code == "LH" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected LH (Lufthansa) in STAR_ALLIANCE group")
	}
}

func TestExpandAirlineTokens_MixedInputs(t *testing.T) {
	inputs := []string{"FR", "GROUP:LOW_COST"} // FR is in LOW_COST, should be deduped

	codes, _, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Count FR occurrences
	frCount := 0
	for _, code := range codes {
		if code == "FR" {
			frCount++
		}
	}

	if frCount != 1 {
		t.Errorf("expected FR to appear exactly once, appeared %d times", frCount)
	}
}

func TestExpandAirlineTokens_InvalidCode(t *testing.T) {
	inputs := []string{"UAA"} // 3 characters

	_, _, err := ExpandAirlineTokens(inputs)
	if err == nil {
		t.Error("expected error for invalid airline code")
	}
}

func TestExpandAirlineTokens_UnknownGroup(t *testing.T) {
	inputs := []string{"GROUP:UNKNOWN"}

	_, _, err := ExpandAirlineTokens(inputs)
	if err == nil {
		t.Error("expected error for unknown airline group")
	}
}

func TestAirlineGroupsForCodes_SingleCarrier(t *testing.T) {
	groups := AirlineGroupsForCodes([]string{"UA"})

	// UA (United) is Star Alliance
	found := false
	for _, g := range groups {
		if g == GroupStarAlliance {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected STAR_ALLIANCE for UA, got %v", groups)
	}
}

func TestAirlineGroupsForCodes_LowCostCarrier(t *testing.T) {
	groups := AirlineGroupsForCodes([]string{"FR"}) // Ryanair

	found := false
	for _, g := range groups {
		if g == GroupLowCost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected LOW_COST for FR, got %v", groups)
	}
}

func TestAirlineGroupsForCodes_MixedCarriers(t *testing.T) {
	groups := AirlineGroupsForCodes([]string{"UA", "AA", "FR"})

	// Should have STAR_ALLIANCE (UA), ONEWORLD (AA), LOW_COST (FR)
	expected := map[string]bool{
		GroupStarAlliance: true,
		GroupOneworld:     true,
		GroupLowCost:      true,
	}

	for _, g := range groups {
		delete(expected, g)
	}

	if len(expected) != 0 {
		t.Errorf("missing groups: %v", expected)
	}
}

func TestAirlineGroupsForCodes_UnknownCarrier(t *testing.T) {
	groups := AirlineGroupsForCodes([]string{"ZZ"}) // Unknown

	if len(groups) != 0 {
		t.Errorf("expected empty groups for unknown carrier, got %v", groups)
	}
}

func TestAirlineGroupsForCodes_Deduplication(t *testing.T) {
	// Both LH and UA are Star Alliance
	groups := AirlineGroupsForCodes([]string{"LH", "UA"})

	count := 0
	for _, g := range groups {
		if g == GroupStarAlliance {
			count++
		}
	}

	if count != 1 {
		t.Errorf("expected STAR_ALLIANCE to appear once, appeared %d times", count)
	}
}

func TestGetAllAirlineGroupInfo(t *testing.T) {
	info := GetAllAirlineGroupInfo()

	if len(info) != 4 {
		t.Errorf("expected 4 groups, got %d", len(info))
	}

	tokens := make(map[string]bool)
	for _, g := range info {
		tokens[g.Token] = true
		if g.AirlineCount == 0 {
			t.Errorf("expected non-zero airline count for group %s", g.Token)
		}
		if len(g.SampleCodes) == 0 {
			t.Errorf("expected sample codes for group %s", g.Token)
		}
	}

	// Verify all 4 groups are present
	expectedTokens := []string{GroupStarAlliance, GroupOneworld, GroupSkyTeam, GroupLowCost}
	for _, expected := range expectedTokens {
		if !tokens[expected] {
			t.Errorf("missing group token: %s", expected)
		}
	}
}

func TestIsAirlineGroupToken(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"GROUP:STAR_ALLIANCE", true},
		{"group:oneworld", true},
		{"Group:SkyTeam", true},
		{"UA", false},
		{"REGION:EUROPE", false},
		{"", false},
	}

	for _, tc := range tests {
		result := IsAirlineGroupToken(tc.input)
		if result != tc.expected {
			t.Errorf("IsAirlineGroupToken(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestExtractAirlineCodeFromFlightNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UA123", "UA"},
		{"AA1234", "AA"},
		{"FR8001", "FR"},
		{"DL302", "DL"},
		{"U21234", "U2"},
		{"9C8512", "9C"},
		{"A", ""},
		{"", ""},
	}

	for _, tc := range tests {
		result := ExtractAirlineCodeFromFlightNumber(tc.input)
		if result != tc.expected {
			t.Errorf("ExtractAirlineCodeFromFlightNumber(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestExpandAirlineTokens_EmptyInput(t *testing.T) {
	inputs := []string{"", "  ", "UA"}

	codes, _, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 1 {
		t.Errorf("expected 1 airline code, got %d: %v", len(codes), codes)
	}
}

func TestExpandAirlineTokens_AlphanumericCodes(t *testing.T) {
	// Some airline codes contain digits (e.g., U2, 9C)
	inputs := []string{"U2", "9C"}

	codes, _, err := ExpandAirlineTokens(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(codes) != 2 {
		t.Errorf("expected 2 airline codes, got %d: %v", len(codes), codes)
	}
}
