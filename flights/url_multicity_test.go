package flights

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/flights/internal/urlpb"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
)

func TestSerializeURL_MultiCityTripType(t *testing.T) {
	session, err := New()
	require.NoError(t, err)

	date1, err := time.Parse("2006-01-02", "2026-06-01")
	require.NoError(t, err)
	date2, err := time.Parse("2006-01-02", "2026-06-05")
	require.NoError(t, err)

	serialized, err := session.SerializeURL(
		context.Background(),
		Args{
			Segments: []Segment{
				{
					Date:        date1,
					SrcAirports: []string{"SFO"},
					DstAirports: []string{"JFK"},
				},
				{
					Date:        date2,
					SrcAirports: []string{"JFK"},
					DstAirports: []string{"LHR"},
				},
			},
			Options: Options{
				Travelers: Travelers{Adults: 1},
				Currency:  currency.USD,
				Stops:     AnyStops,
				Class:     Economy,
				TripType:  MultiCity,
				Lang:      language.English,
			},
		},
	)
	require.NoError(t, err)

	parsed, err := url.Parse(serialized)
	require.NoError(t, err)

	tfs := parsed.Query().Get("tfs")
	require.NotEmpty(t, tfs)

	raw, err := base64.RawURLEncoding.DecodeString(tfs)
	require.NoError(t, err)

	var u urlpb.Url
	require.NoError(t, proto.Unmarshal(raw, &u))
	require.Equal(t, urlpb.Url_MULTI_CITY, u.TripType)
	require.Len(t, u.Flight, 2)
}
