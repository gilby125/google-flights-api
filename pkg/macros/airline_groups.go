// Package macros provides expansion utilities for region and airline group tokens.
package macros

import (
	"fmt"
	"regexp"
	"strings"
)

// groupPrefix is the prefix for airline group tokens.
const groupPrefix = "GROUP:"

// Airline group tokens supported by the API.
const (
	GroupStarAlliance = "GROUP:STAR_ALLIANCE"
	GroupOneworld     = "GROUP:ONEWORLD"
	GroupSkyTeam      = "GROUP:SKYTEAM"
	GroupLowCost      = "GROUP:LOW_COST"
)

// AirlineGroup represents bitmask flags for airline group membership.
type AirlineGroup uint8

const (
	// GroupFlagStarAlliance indicates membership in Star Alliance.
	GroupFlagStarAlliance AirlineGroup = 1 << iota
	// GroupFlagOneworld indicates membership in Oneworld.
	GroupFlagOneworld
	// GroupFlagSkyTeam indicates membership in SkyTeam.
	GroupFlagSkyTeam
	// GroupFlagLowCost indicates a low-cost carrier.
	GroupFlagLowCost
)

// airlineCodePattern validates IATA airline codes (2 alphanumeric characters).
var airlineCodePattern = regexp.MustCompile(`^[A-Z0-9]{2}$`)

// airlineToGroups maps airline IATA codes to their group memberships (bitmask).
//
// IMPORTANT: Best-effort mapping, may drift over time!
// - Alliance memberships change (airlines join/leave)
// - Low-cost classification is subjective
// - Some legacy carriers may be missing or outdated
//
// Safe for: tagging offers with metadata, UI hints, sorting
// NOT safe for: hard filtering, business logic that assumes completeness
//
// To update: check official alliance websites periodically
var airlineToGroups = map[string]AirlineGroup{
	// Star Alliance members (major carriers)
	"LH": GroupFlagStarAlliance, // Lufthansa
	"UA": GroupFlagStarAlliance, // United Airlines
	"AC": GroupFlagStarAlliance, // Air Canada
	"NH": GroupFlagStarAlliance, // ANA (All Nippon Airways)
	"SQ": GroupFlagStarAlliance, // Singapore Airlines
	"TG": GroupFlagStarAlliance, // Thai Airways
	"SK": GroupFlagStarAlliance, // SAS Scandinavian Airlines
	"OS": GroupFlagStarAlliance, // Austrian Airlines
	"LX": GroupFlagStarAlliance, // Swiss International Air Lines
	"TK": GroupFlagStarAlliance, // Turkish Airlines
	"ET": GroupFlagStarAlliance, // Ethiopian Airlines
	"A3": GroupFlagStarAlliance, // Aegean Airlines
	"OU": GroupFlagStarAlliance, // Croatia Airlines
	"SA": GroupFlagStarAlliance, // South African Airways
	"NZ": GroupFlagStarAlliance, // Air New Zealand
	"BR": GroupFlagStarAlliance, // EVA Air
	"OZ": GroupFlagStarAlliance, // Asiana Airlines
	"CA": GroupFlagStarAlliance, // Air China
	"AI": GroupFlagStarAlliance, // Air India
	"AV": GroupFlagStarAlliance, // Avianca
	"CM": GroupFlagStarAlliance, // Copa Airlines
	"TP": GroupFlagStarAlliance, // TAP Air Portugal
	"MS": GroupFlagStarAlliance, // EgyptAir
	"JP": GroupFlagStarAlliance, // Adria Airways
	"LO": GroupFlagStarAlliance, // LOT Polish Airlines

	// Oneworld members (major carriers)
	"AA": GroupFlagOneworld, // American Airlines
	"BA": GroupFlagOneworld, // British Airways
	"QF": GroupFlagOneworld, // Qantas
	"CX": GroupFlagOneworld, // Cathay Pacific
	"JL": GroupFlagOneworld, // Japan Airlines
	"IB": GroupFlagOneworld, // Iberia
	"AY": GroupFlagOneworld, // Finnair
	"MH": GroupFlagOneworld, // Malaysia Airlines
	"QR": GroupFlagOneworld, // Qatar Airways
	"RJ": GroupFlagOneworld, // Royal Jordanian
	"S7": GroupFlagOneworld, // S7 Airlines
	"UL": GroupFlagOneworld, // SriLankan Airlines
	"FJ": GroupFlagOneworld, // Fiji Airways
	"4M": GroupFlagOneworld, // LATAM Argentina
	"LA": GroupFlagOneworld, // LATAM Airlines
	"AB": GroupFlagOneworld, // American Eagle (operated as AA)

	// SkyTeam members (major carriers)
	"AF": GroupFlagSkyTeam, // Air France
	"KL": GroupFlagSkyTeam, // KLM Royal Dutch Airlines
	"DL": GroupFlagSkyTeam, // Delta Air Lines
	"AM": GroupFlagSkyTeam, // Aeromexico
	"KE": GroupFlagSkyTeam, // Korean Air
	"AZ": GroupFlagSkyTeam, // ITA Airways (formerly Alitalia)
	"CI": GroupFlagSkyTeam, // China Airlines
	"MU": GroupFlagSkyTeam, // China Eastern Airlines
	"SU": GroupFlagSkyTeam, // Aeroflot
	"VN": GroupFlagSkyTeam, // Vietnam Airlines
	"GA": GroupFlagSkyTeam, // Garuda Indonesia
	"ME": GroupFlagSkyTeam, // Middle East Airlines
	"SV": GroupFlagSkyTeam, // Saudia
	"OK": GroupFlagSkyTeam, // Czech Airlines
	"RO": GroupFlagSkyTeam, // TAROM
	"AR": GroupFlagSkyTeam, // Aerolíneas Argentinas
	"UX": GroupFlagSkyTeam, // Air Europa

	// Low-cost carriers (major ones)
	"FR": GroupFlagLowCost, // Ryanair
	"U2": GroupFlagLowCost, // easyJet
	"W6": GroupFlagLowCost, // Wizz Air
	"NK": GroupFlagLowCost, // Spirit Airlines
	"F9": GroupFlagLowCost, // Frontier Airlines
	"WN": GroupFlagLowCost, // Southwest Airlines
	"G4": GroupFlagLowCost, // Allegiant Air
	"B6": GroupFlagLowCost, // JetBlue Airways
	"DY": GroupFlagLowCost, // Norwegian Air Shuttle
	"VY": GroupFlagLowCost, // Vueling
	"PC": GroupFlagLowCost, // Pegasus Airlines
	"AK": GroupFlagLowCost, // AirAsia
	"QZ": GroupFlagLowCost, // Indonesia AirAsia
	"FD": GroupFlagLowCost, // Thai AirAsia
	"TR": GroupFlagLowCost, // Scoot
	"3K": GroupFlagLowCost, // Jetstar Asia Airways
	"JQ": GroupFlagLowCost, // Jetstar Airways
	"GK": GroupFlagLowCost, // Jetstar Japan
	"9C": GroupFlagLowCost, // Spring Airlines
	"5J": GroupFlagLowCost, // Cebu Pacific
	"Z2": GroupFlagLowCost, // Philippines AirAsia
	"ZE": GroupFlagLowCost, // Eastar Jet
	"LJ": GroupFlagLowCost, // Jin Air
	"7C": GroupFlagLowCost, // Jeju Air
	"TW": GroupFlagLowCost, // T'way Air
	"MM": GroupFlagLowCost, // Peach Aviation
	"BC": GroupFlagLowCost, // Skymark Airlines
	"HO": GroupFlagLowCost, // Juneyao Airlines
	"OD": GroupFlagLowCost, // Batik Air Malaysia
	"XT": GroupFlagLowCost, // Indonesia AirAsia X
	"D7": GroupFlagLowCost, // AirAsia X
	"XJ": GroupFlagLowCost, // Thai AirAsia X
	"G9": GroupFlagLowCost, // Air Arabia
	"FZ": GroupFlagLowCost, // flydubai
	"XY": GroupFlagLowCost, // flynas
	"J9": GroupFlagLowCost, // Jazeera Airways
	"8Q": GroupFlagLowCost, // Onur Air
}

// groupTokenToFlag maps group tokens to their bitmask flags.
var groupTokenToFlag = map[string]AirlineGroup{
	GroupStarAlliance: GroupFlagStarAlliance,
	GroupOneworld:     GroupFlagOneworld,
	GroupSkyTeam:      GroupFlagSkyTeam,
	GroupLowCost:      GroupFlagLowCost,
}

// flagToGroupToken maps bitmask flags to their group tokens.
var flagToGroupToken = map[AirlineGroup]string{
	GroupFlagStarAlliance: GroupStarAlliance,
	GroupFlagOneworld:     GroupOneworld,
	GroupFlagSkyTeam:      GroupSkyTeam,
	GroupFlagLowCost:      GroupLowCost,
}

// AllAirlineGroups returns all supported airline group tokens.
func AllAirlineGroups() []string {
	return []string{
		GroupStarAlliance,
		GroupOneworld,
		GroupSkyTeam,
		GroupLowCost,
	}
}

// AirlineGroupInfo contains metadata about an airline group for API responses.
type AirlineGroupInfo struct {
	Token        string   `json:"token"`
	AirlineCount int      `json:"airline_count"`
	SampleCodes  []string `json:"sample_codes"`
}

// GetAllAirlineGroupInfo returns metadata about all airline groups for the metadata endpoint.
func GetAllAirlineGroupInfo() []AirlineGroupInfo {
	// Count airlines per group
	groupCounts := make(map[string][]string)
	for code, groups := range airlineToGroups {
		for flag, token := range flagToGroupToken {
			if groups&flag != 0 {
				groupCounts[token] = append(groupCounts[token], code)
			}
		}
	}

	allGroups := AllAirlineGroups()
	result := make([]AirlineGroupInfo, 0, len(allGroups))

	for _, token := range allGroups {
		codes := groupCounts[token]
		if codes == nil {
			codes = []string{}
		}

		// Get up to 5 sample codes
		sampleCount := 5
		if len(codes) < sampleCount {
			sampleCount = len(codes)
		}
		samples := make([]string, sampleCount)
		copy(samples, codes[:sampleCount])

		result = append(result, AirlineGroupInfo{
			Token:        token,
			AirlineCount: len(codes),
			SampleCodes:  samples,
		})
	}

	return result
}

// GetAirlineGroupsForCode returns the list of group tokens for a given airline code.
// Returns an empty slice if the airline is not in any group.
func GetAirlineGroupsForCode(code string) []string {
	code = strings.ToUpper(strings.TrimSpace(code))
	groups, ok := airlineToGroups[code]
	if !ok {
		return nil
	}

	result := make([]string, 0, 4)
	for flag, token := range flagToGroupToken {
		if groups&flag != 0 {
			result = append(result, token)
		}
	}
	return result
}

// AirlineGroupsForCodes tags a list of airline codes with group labels.
// Example return: ["GROUP:STAR_ALLIANCE"] or ["GROUP:LOW_COST","GROUP:ONEWORLD"].
// Returns unique group labels across all provided codes.
func AirlineGroupsForCodes(codes []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, 4)

	for _, code := range codes {
		groups := GetAirlineGroupsForCode(code)
		for _, group := range groups {
			if !seen[group] {
				seen[group] = true
				result = append(result, group)
			}
		}
	}

	return result
}

// IsAirlineGroupToken returns true if the input looks like an airline group token.
func IsAirlineGroupToken(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), groupPrefix)
}

// GetGroupAirlines returns the list of airline codes for a group token.
// Returns nil if the group is not recognized.
func GetGroupAirlines(group string) []string {
	flag, ok := groupTokenToFlag[group]
	if !ok {
		return nil
	}

	result := make([]string, 0)
	for code, groups := range airlineToGroups {
		if groups&flag != 0 {
			result = append(result, code)
		}
	}
	return result
}

// ExpandAirlineTokens expands airline IATA codes + GROUP:* tokens into airline codes.
// - uppercases, trims, dedupes
// - validates airlines: ^[A-Z0-9]{2}$ (digits allowed)
// - unknown token => error
func ExpandAirlineTokens(inputs []string) (codes []string, warnings []string, err error) {
	seen := make(map[string]bool)
	result := make([]string, 0, len(inputs))

	for _, input := range inputs {
		token := strings.ToUpper(strings.TrimSpace(input))
		if token == "" {
			continue
		}

		if IsAirlineGroupToken(token) {
			// Expand group token
			groupAirlines := GetGroupAirlines(token)
			if groupAirlines == nil {
				return nil, nil, fmt.Errorf("unknown airline group token: %s", token)
			}

			if len(groupAirlines) == 0 {
				warnings = append(warnings, fmt.Sprintf("airline group %s contains no airlines", token))
				continue
			}

			for _, code := range groupAirlines {
				if !seen[code] {
					seen[code] = true
					result = append(result, code)
				}
			}
		} else {
			// Validate as IATA airline code
			if !airlineCodePattern.MatchString(token) {
				return nil, nil, fmt.Errorf("invalid airline code format: %s (expected 2 alphanumeric characters)", input)
			}

			if !seen[token] {
				seen[token] = true
				result = append(result, token)
			}
		}
	}

	return result, warnings, nil
}

// ExtractAirlineCodeFromFlightNumber extracts the airline code prefix from a flight number.
// Flight numbers typically start with the 2-character airline code followed by digits.
// Examples: "UA123" -> "UA", "AA1234" -> "AA", "FR8001" -> "FR"
func ExtractAirlineCodeFromFlightNumber(flightNumber string) string {
	if len(flightNumber) < 2 {
		return ""
	}

	// Try 2-character prefix first (most common)
	prefix := strings.ToUpper(flightNumber[:2])
	if airlineCodePattern.MatchString(prefix) {
		return prefix
	}

	// Some flight numbers might have 3-character ICAO codes followed by numbers
	// but we only support 2-character IATA codes
	return ""
}
