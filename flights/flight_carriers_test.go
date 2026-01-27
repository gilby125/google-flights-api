package flights

import (
	"context"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

func TestGetRawData_CarriersInjected(t *testing.T) {
	session := &Session{}

	args := Args{
		Date:        mustParseDate(t, "2026-02-14"),
		ReturnDate:  mustParseDate(t, "2026-02-21"),
		SrcAirports: []string{"LAX"},
		DstAirports: []string{"JFK"},
		Options: Options{
			Travelers: Travelers{Adults: 1},
			Currency:  currency.USD,
			Stops:     AnyStops,
			Class:     Economy,
			TripType:  RoundTrip,
			Lang:      language.English,
			Carriers:  []string{"ua", "DL", "UA", "STAR_ALLIANCE"},
		},
	}

	raw, err := session.getRawData(context.Background(), args)
	if err != nil {
		t.Fatalf("getRawData: %v", err)
	}

	// Ensure the carrier filter array is present in both legs.
	if !strings.Contains(raw, `,0,[\"UA\",\"DL\",\"STAR_ALLIANCE\"],[],\"2026-02-14\"`) {
		t.Fatalf("expected outbound carrier filter in rawData, got: %s", raw)
	}
	if !strings.Contains(raw, `,0,[\"UA\",\"DL\",\"STAR_ALLIANCE\"],[],\"2026-02-21\"`) {
		t.Fatalf("expected return carrier filter in rawData, got: %s", raw)
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}
