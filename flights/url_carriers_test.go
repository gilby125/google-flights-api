package flights

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/flights/internal/urlpb"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
)

func TestSerializeURL_Carriers(t *testing.T) {
	session := &Session{}

	date := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC)

	u, err := session.SerializeURL(
		context.Background(),
		Args{
			Date:        date,
			ReturnDate:  returnDate,
			SrcAirports: []string{"LAX"},
			DstAirports: []string{"JFK"},
			Options: Options{
				Travelers: Travelers{Adults: 1},
				Currency:  currency.USD,
				Stops:     AnyStops,
				Class:     Economy,
				TripType:  RoundTrip,
				Lang:      language.English,
				Carriers:  []string{"UA", "STAR_ALLIANCE"},
			},
		},
	)
	if err != nil {
		t.Fatalf("SerializeURL: %v", err)
	}

	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	tfs := parsed.Query().Get("tfs")
	if tfs == "" {
		t.Fatalf("missing tfs param: %s", u)
	}
	raw, err := base64.RawURLEncoding.DecodeString(tfs)
	if err != nil {
		t.Fatalf("decode tfs: %v", err)
	}

	var msg urlpb.Url
	if err := proto.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal tfs: %v", err)
	}
	if len(msg.Flight) != 2 {
		t.Fatalf("expected 2 flights in tfs, got %d", len(msg.Flight))
	}
	for i, f := range msg.Flight {
		if len(f.Carriers) == 0 {
			t.Fatalf("expected carriers in flight[%d]", i)
		}
	}
}
