// Package macros provides expansion utilities for region and airline group tokens.
//
// IMPORTANT: v1 Limitations
//   - Region expansion is a curated set (primarily db.Top100Airports, plus small
//     explicit lists for regions that are underrepresented in that list).
//   - REGION:* is not a complete list of all airports in that region.
//   - Airline group membership is best-effort and may drift over time; safe for tagging, not for filtering
package macros

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gilby125/google-flights-api/db"
)

// regionPrefix is the prefix for region tokens.
const regionPrefix = "REGION:"

// Region tokens supported by the API.
// NOTE: These expand to airports from db.Top100Airports plus small explicit lists
// for regions that are underrepresented in that list (v1 constraint).
const (
	RegionEurope       = "REGION:EUROPE"
	RegionNorthAmerica = "REGION:NORTH_AMERICA"
	RegionSouthAmerica = "REGION:SOUTH_AMERICA"
	RegionAsia         = "REGION:ASIA"
	RegionCaribbean    = "REGION:CARIBBEAN"
	RegionOceania      = "REGION:OCEANIA"
	RegionMiddleEast   = "REGION:MIDDLE_EAST"
	RegionAfrica       = "REGION:AFRICA"
	RegionWorld        = "REGION:WORLD"
	// RegionWorldAll expands to *all* airports supported by this server (backed by Postgres airports table).
	// NOTE: This token requires a server-side override list; it is not included in the canonical Top100Airports set.
	RegionWorldAll = "REGION:WORLD_ALL"
)

// airportCodePattern validates IATA airport codes (3 uppercase letters).
var airportCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

// countryToRegion maps ISO 3166-1 alpha-2 country codes to region tokens.
// This mapping is based on common geographic classifications.
// NOTE: Only countries present in db.Top100Airports will have any effect.
var countryToRegion = map[string]string{
	// North America
	"US": RegionNorthAmerica,
	"CA": RegionNorthAmerica,
	"MX": RegionNorthAmerica, // Mexico is geographically North America

	// Europe
	"GB": RegionEurope,
	"FR": RegionEurope,
	"DE": RegionEurope,
	"NL": RegionEurope,
	"ES": RegionEurope,
	"IT": RegionEurope,
	"CH": RegionEurope,
	"AT": RegionEurope,
	"PT": RegionEurope,
	"NO": RegionEurope,
	"DK": RegionEurope,
	"SE": RegionEurope,
	"IE": RegionEurope,
	"GR": RegionEurope,
	"BE": RegionEurope,
	"PL": RegionEurope,
	"TR": RegionEurope, // Turkey spans Europe/Asia, often grouped with Europe for flights
	"RU": RegionEurope, // Russia spans Europe/Asia, Moscow grouped with Europe

	// Asia
	"JP": RegionAsia,
	"KR": RegionAsia,
	"CN": RegionAsia,
	"HK": RegionAsia,
	"TW": RegionAsia,
	"SG": RegionAsia,
	"TH": RegionAsia,
	"MY": RegionAsia,
	"PH": RegionAsia,
	"ID": RegionAsia,
	"IN": RegionAsia,

	// Middle East
	"AE": RegionMiddleEast,
	"QA": RegionMiddleEast,

	// Oceania
	"AU": RegionOceania,
	"NZ": RegionOceania,

	// South America
	"BR": RegionSouthAmerica,
	"CO": RegionSouthAmerica,

	// Africa
	"ZA": RegionAfrica,
	"MA": RegionAfrica,
	"EG": RegionAfrica,

	// Caribbean (not currently in Top100Airports, but defined for completeness)
	"PR": RegionCaribbean, // Puerto Rico
	"DO": RegionCaribbean, // Dominican Republic
	"JM": RegionCaribbean, // Jamaica
	"BS": RegionCaribbean, // Bahamas
	"BB": RegionCaribbean, // Barbados
	"TT": RegionCaribbean, // Trinidad and Tobago
	"AW": RegionCaribbean, // Aruba
}

// regionAirportsCache stores precomputed airport lists by region.
// Initialized once via sync.Once for thread-safety.
var (
	regionAirportsCache map[string][]string
	regionCacheOnce     sync.Once
)

// explicitRegionAirports augments region expansion beyond db.Top100Airports for regions that
// would otherwise be empty or severely underrepresented.
var explicitRegionAirports = map[string][]string{
	RegionCaribbean: {
		// ABC islands + southern Caribbean
		"AUA", // Aruba (Queen Beatrix)
		"BON", // Bonaire (Flamingo)
		"CUR", // Curaçao (Hato)
		// French Antilles / Leeward islands
		"SBH", // Saint Barthélemy (Gustaf III)
		"SFG", // Saint Martin (Grand Case; French side)
		"SXM", // Sint Maarten (Princess Juliana; Dutch side)
		"PTP", // Guadeloupe (Pointe-à-Pitre)
		"FDF", // Martinique (Fort-de-France)
		// Northern / Eastern Caribbean
		"AXA", // Anguilla (Clayton J. Lloyd)
		"ANU", // Antigua (V. C. Bird)
		"BBQ", // Barbuda (Codrington)
		"SKB", // St Kitts & Nevis (Robert L. Bradshaw)
		"NEV", // Nevis (Vance W. Amory)
		"MNI", // Montserrat (John A. Osborne)
		"SAB", // Saba (Juancho E. Yrausquin)
		"EUX", // Sint Eustatius (F.D. Roosevelt)
		"DOM", // Dominica (Douglas-Charles)
		"UVF", // Saint Lucia (Hewanorra)
		"SLU", // Saint Lucia (George F. L. Charles; intra-island)
		"SVD", // St Vincent & the Grenadines (Argyle)
		"GND", // Grenada (Maurice Bishop)
		"BGI", // Barbados (Grantley Adams)
		// Greater Antilles
		"HAV", // Cuba (Havana)
		"SJU", // Puerto Rico (San Juan)
		"STT", // U.S. Virgin Islands (St Thomas)
		"STX", // U.S. Virgin Islands (St Croix)
		"EIS", // British Virgin Islands (Tortola)
		"KIN", // Jamaica (Kingston)
		"MBJ", // Jamaica (Montego Bay)
		"PAP", // Haiti (Port-au-Prince)
		"CAP", // Haiti (Cap-Haïtien)
		"SDQ", // Dominican Republic (Santo Domingo)
		"PUJ", // Dominican Republic (Punta Cana)
		"POP", // Dominican Republic (Puerto Plata)
		// Bahamas + nearby island territories
		"NAS", // Bahamas (Nassau)
		"FPO", // Bahamas (Freeport; Grand Bahama)
		"GGT", // Bahamas (Exuma)
		"MHH", // Bahamas (Marsh Harbour; Abaco)
		"ELH", // Bahamas (North Eleuthera)
		"ZSA", // Bahamas (San Salvador)
		"PLS", // Turks & Caicos (Providenciales)
		"GDT", // Turks & Caicos (Grand Turk)
		"GCM", // Cayman Islands (Grand Cayman)
		"CYB", // Cayman Islands (Cayman Brac)
		"LYB", // Cayman Islands (Little Cayman)
		"BDA", // Bermuda (often grouped with Caribbean travel)
		// Wider Caribbean basin islands (commonly searched as "Caribbean")
		"ADZ", // Colombia (San Andrés Island)
		"BOC", // Panama (Bocas del Toro)
		"CZM", // Mexico (Cozumel)
		"SPR", // Belize (San Pedro; Ambergris Caye)
		"GJA", // Honduras (Guanaja)
		"RTB", // Honduras (Roatán)
		"UII", // Honduras (Utila)
		"RNI", // Nicaragua (Corn Islands)
		"PMV", // Venezuela (Margarita / Porlamar)
		"POS", // Trinidad (Port of Spain)
		"TAB", // Tobago (A.N.R. Robinson)
	},
	RegionOceania: {
		"ADL", // Adelaide
		"AKL", // Auckland
		"BNE", // Brisbane
		"CHC", // Christchurch
		"GUM", // Guam
		"MEL", // Melbourne
		"NAN", // Fiji (Nadi)
		"PER", // Perth
		"PPT", // Tahiti
		"SYD", // Sydney
		"WLG", // Wellington
	},
}

// initRegionCache builds the region -> airports mapping from Top100Airports + explicitRegionAirports.
// Thread-safe via sync.Once.
func initRegionCache() {
	regionCacheOnce.Do(func() {
		regionAirportsCache = make(map[string][]string)

		// Pre-initialize ALL advertised regions with empty slices
		// This ensures regions with zero airports (e.g., CARIBBEAN) are recognized
		// as valid tokens that expand to empty, rather than "unknown region" errors
		for _, region := range AllRegions() {
			regionAirportsCache[region] = []string{}
		}

		allAirports := make([]string, 0, len(db.Top100Airports))

		for _, airport := range db.Top100Airports {
			allAirports = append(allAirports, airport.Code)

			region, ok := countryToRegion[airport.Country]
			if ok {
				regionAirportsCache[region] = append(regionAirportsCache[region], airport.Code)
			}
		}

		// Merge explicit region airports.
		for region, airports := range explicitRegionAirports {
			for _, code := range airports {
				normalized := strings.ToUpper(strings.TrimSpace(code))
				if !airportCodePattern.MatchString(normalized) {
					continue
				}
				regionAirportsCache[region] = append(regionAirportsCache[region], normalized)
			}
		}

		// Deduplicate each region in insertion order.
		for region, airports := range regionAirportsCache {
			if len(airports) == 0 {
				continue
			}
			seen := make(map[string]struct{}, len(airports))
			out := make([]string, 0, len(airports))
			for _, code := range airports {
				if _, ok := seen[code]; ok {
					continue
				}
				seen[code] = struct{}{}
				out = append(out, code)
			}
			regionAirportsCache[region] = out
		}

		// REGION:WORLD expands to the union of all known region airports (excluding WORLD_ALL).
		worldSeen := make(map[string]struct{})
		world := make([]string, 0, len(allAirports))
		for _, code := range allAirports {
			if _, ok := worldSeen[code]; ok {
				continue
			}
			worldSeen[code] = struct{}{}
			world = append(world, code)
		}
		for region, airports := range regionAirportsCache {
			if region == RegionWorld || region == RegionWorldAll {
				continue
			}
			for _, code := range airports {
				if _, ok := worldSeen[code]; ok {
					continue
				}
				worldSeen[code] = struct{}{}
				world = append(world, code)
			}
		}

		regionAirportsCache[RegionWorld] = world
	})
}

// GetRegionAirports returns the list of airport codes for a region token.
// Returns nil if the region is not recognized.
func GetRegionAirports(region string) []string {
	initRegionCache()
	airports, ok := regionAirportsCache[region]
	if !ok {
		return nil
	}
	// Return a copy to prevent mutation
	result := make([]string, len(airports))
	copy(result, airports)
	return result
}

// AllRegions returns all supported region tokens.
func AllRegions() []string {
	return []string{
		RegionEurope,
		RegionNorthAmerica,
		RegionSouthAmerica,
		RegionAsia,
		RegionCaribbean,
		RegionOceania,
		RegionMiddleEast,
		RegionAfrica,
		RegionWorld,
		RegionWorldAll,
	}
}

// RegionInfo contains metadata about a region for API responses.
type RegionInfo struct {
	Token          string   `json:"token"`
	AirportCount   int      `json:"airport_count"`
	SampleAirports []string `json:"sample_airports"`
}

// GetAllRegionInfo returns metadata about all regions for the metadata endpoint.
func GetAllRegionInfo() []RegionInfo {
	initRegionCache()

	regions := AllRegions()
	result := make([]RegionInfo, 0, len(regions))

	for _, region := range regions {
		airports := GetRegionAirports(region)
		if airports == nil {
			airports = []string{}
		}

		// Get up to 5 sample airports
		sampleCount := 5
		if len(airports) < sampleCount {
			sampleCount = len(airports)
		}
		samples := make([]string, sampleCount)
		copy(samples, airports[:sampleCount])

		result = append(result, RegionInfo{
			Token:          region,
			AirportCount:   len(airports),
			SampleAirports: samples,
		})
	}

	return result
}

// IsRegionToken returns true if the input looks like a region token.
func IsRegionToken(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), regionPrefix)
}

// ExpandAirportTokens expands airport IATA codes + REGION:* tokens into airport codes.
// - uppercases, trims, dedupes
// - validates airports: ^[A-Z]{3}$
// - unknown token => error (handlers return 400)
func ExpandAirportTokens(inputs []string) (airports []string, warnings []string, err error) {
	return ExpandAirportTokensWithOverrides(inputs, nil)
}

// ExpandAirportTokensWithOverrides expands airport IATA codes + REGION:* tokens into airport codes,
// allowing the caller to provide explicit expansions for specific region tokens (e.g. REGION:WORLD_ALL).
//
// Overrides are token -> []IATA (case-insensitive on the token key).
func ExpandAirportTokensWithOverrides(inputs []string, overrides map[string][]string) (airports []string, warnings []string, err error) {
	initRegionCache()

	seen := make(map[string]bool)
	result := make([]string, 0, len(inputs))

	normalizedOverrides := make(map[string][]string, len(overrides))
	for k, v := range overrides {
		normalizedOverrides[strings.ToUpper(strings.TrimSpace(k))] = v
	}

	for _, input := range inputs {
		token := strings.ToUpper(strings.TrimSpace(input))
		if token == "" {
			continue
		}

		if IsRegionToken(token) {
			// Expand region token (with optional override)
			var regionAirports []string
			if override, ok := normalizedOverrides[token]; ok {
				regionAirports = override
			} else {
				if token == RegionWorldAll {
					return nil, nil, fmt.Errorf("%s requires a server-side airport list override and is only supported by endpoints that explicitly enable it", RegionWorldAll)
				}
				regionAirports = GetRegionAirports(token)
				if regionAirports == nil {
					return nil, nil, fmt.Errorf("unknown region token: %s", token)
				}
			}

			if len(regionAirports) == 0 {
				warnings = append(warnings, fmt.Sprintf("region %s contains no airports in configured set", token))
				continue
			}

			for _, code := range regionAirports {
				if !seen[code] {
					seen[code] = true
					result = append(result, code)
				}
			}
		} else {
			// Validate as IATA airport code
			if !airportCodePattern.MatchString(token) {
				return nil, nil, fmt.Errorf("invalid airport code format: %s (expected 3 uppercase letters)", input)
			}

			if !seen[token] {
				seen[token] = true
				result = append(result, token)
			}
		}
	}

	return result, warnings, nil
}

// ValidateNoRegionTokens returns an error if any input contains a REGION:* token.
// Used for single-route endpoints that don't support region expansion.
func ValidateNoRegionTokens(inputs ...string) error {
	for _, input := range inputs {
		token := strings.ToUpper(strings.TrimSpace(input))
		if IsRegionToken(token) {
			return fmt.Errorf("region tokens are not supported on single-route endpoints; use bulk search or price graph sweep endpoints instead: %s", input)
		}
	}
	return nil
}
