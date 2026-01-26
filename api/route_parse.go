package api

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gilby125/google-flights-api/pkg/macros"
)

var airportCodePattern = regexp.MustCompile(`^[A-Z0-9]{3}$`)
var spaceSeparatedCodesPattern = regexp.MustCompile(`^[A-Z0-9]{3}(?:\s+[A-Z0-9]{3})+$`)

var regionAliasOnce sync.Once
var regionAliasToToken map[string]string

func normalizeAirportToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}

	// Common UI format: "JFK - John F. Kennedy International Airport, New York"
	if strings.Contains(trimmed, " - ") {
		trimmed = strings.SplitN(trimmed, " - ", 2)[0]
	}

	trimmed = strings.ToUpper(strings.TrimSpace(trimmed))
	trimmed = strings.Trim(trimmed, ",;")
	return trimmed
}

func normalizeRegionAlias(value string) string {
	upper := strings.ToUpper(strings.TrimSpace(value))
	upper = strings.TrimPrefix(upper, "REGION:")

	var b strings.Builder
	b.Grow(len(upper))

	lastUnderscore := false
	for _, r := range upper {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ':
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}

	out := strings.Trim(b.String(), "_")
	out = strings.ReplaceAll(out, "_", "")
	// NOTE: we remove underscores to allow matching "NORTHAMERICA" etc.
	return out
}

func initRegionAliasMap() {
	regionAliasToToken = make(map[string]string, len(macros.AllRegions())*2)
	for _, token := range macros.AllRegions() {
		alias := strings.TrimPrefix(strings.ToUpper(token), "REGION:")
		key := normalizeRegionAlias(alias)
		if key != "" {
			regionAliasToToken[key] = token
		}
	}
	// Common shorthand
	regionAliasToToken["MIDEAST"] = macros.RegionMiddleEast
	regionAliasToToken["MIDDLEEAST"] = macros.RegionMiddleEast
}

func canonicalizeRegionToken(value string) (string, bool) {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return "", false
	}

	if macros.IsRegionToken(trimmed) {
		return trimmed, true
	}

	regionAliasOnce.Do(initRegionAliasMap)
	key := normalizeRegionAlias(trimmed)
	if key == "" {
		return "", false
	}
	token, ok := regionAliasToToken[key]
	return token, ok
}

func splitAirportList(raw string) []string {
	input := strings.TrimSpace(raw)
	if input == "" {
		return nil
	}

	upper := strings.ToUpper(input)

	var parts []string
	switch {
	case strings.ContainsAny(upper, ",;|/\\"):
		parts = strings.FieldsFunc(upper, func(r rune) bool {
			switch r {
			case ',', ';', '|', '/', '\\':
				return true
			default:
				return false
			}
		})
	case spaceSeparatedCodesPattern.MatchString(strings.TrimSpace(upper)):
		parts = strings.Fields(upper)
	default:
		parts = []string{upper}
	}

	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		token := normalizeAirportToken(part)
		if token == "" {
			continue
		}

		switch {
		case airportCodePattern.MatchString(token):
			// ok
		case macros.IsRegionToken(token):
			// ok
		default:
			if regionToken, ok := canonicalizeRegionToken(token); ok {
				token = regionToken
			} else {
				continue
			}
		}

		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}

	return out
}

// ParseRouteInputs supports:
// - Single route: origin="JFK", destination="LAX"
// - Multi route cross product: origin="MKE,MSN", destination="FLL,MIA"
// - Combined expression: origin="MKE,MSN>FLL,MIA", destination=""
func ParseRouteInputs(originRaw, destinationRaw string) ([]string, []string, error) {
	originRaw = strings.TrimSpace(originRaw)
	destinationRaw = strings.TrimSpace(destinationRaw)

	if originRaw == "" {
		return nil, nil, fmt.Errorf("origin is required")
	}

	// Allow encoding "origins>destinations" into the origin field.
	if destinationRaw == "" && (strings.Contains(originRaw, ">") || strings.Contains(originRaw, "→")) {
		parts := strings.FieldsFunc(originRaw, func(r rune) bool {
			return r == '>' || r == '→'
		})
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid multi-route expression; expected \"ORIGINS>DESTINATIONS\"")
		}
		originRaw = strings.TrimSpace(parts[0])
		destinationRaw = strings.TrimSpace(parts[1])
	}

	if destinationRaw == "" {
		return nil, nil, fmt.Errorf("destination is required")
	}

	origins := splitAirportList(originRaw)
	destinations := splitAirportList(destinationRaw)

	if len(origins) == 0 {
		return nil, nil, fmt.Errorf("no valid origin airport/region tokens found")
	}
	if len(destinations) == 0 {
		return nil, nil, fmt.Errorf("no valid destination airport/region tokens found")
	}

	return origins, destinations, nil
}
