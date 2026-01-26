package api

import (
	"fmt"
	"regexp"
	"strings"
)

var airportCodePattern = regexp.MustCompile(`^[A-Z0-9]{3}$`)
var spaceSeparatedCodesPattern = regexp.MustCompile(`^[A-Z0-9]{3}(?:\s+[A-Z0-9]{3})+$`)

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
		code := normalizeAirportToken(part)
		if !airportCodePattern.MatchString(code) {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
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
		return nil, nil, fmt.Errorf("no valid origin airport codes found")
	}
	if len(destinations) == 0 {
		return nil, nil, fmt.Errorf("no valid destination airport codes found")
	}

	return origins, destinations, nil
}
