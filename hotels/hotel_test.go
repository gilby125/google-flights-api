package hotels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHotelsFromHTML(t *testing.T) {
	hotel := func(name, price string, rating float64, lat, long float64) []any {
		arr := make([]any, 17)
		arr[0] = name
		arr[2] = price
		arr[3] = []any{"https://example.com/img.jpg"}
		arr[5] = rating
		arr[16] = []any{lat, long}
		return arr
	}

	hotelsList := []any{
		hotel("Hotel A", "$123", 4.3, 37.0, -122.0),
		hotel("Hotel B", "$234", 4.1, 38.0, -123.0),
		hotel("Hotel C", "$345", 4.8, 39.0, -124.0),
		hotel("Hotel D", "$456", 4.0, 40.0, -125.0),
		hotel("Hotel E", "$567", 3.9, 41.0, -126.0),
		hotel("Hotel F", "$678", 4.2, 42.0, -127.0),
	}

	// Nest it so the recursive finder hits the inner list.
	dataJSON := `[[` + mustJSON(t, hotelsList) + `]]`

	html := `<html><head></head><body><script>AF_initDataCallback({key: 'ds:0', data:` + dataJSON + `, sideChannel: {}});</script></body></html>`

	parsed, err := parseHotelsFromHTML(html, "USD")
	require.NoError(t, err)
	require.Len(t, parsed, 6)
	require.Equal(t, "Hotel A", parsed[0].Name)
	require.Equal(t, 123.0, parsed[0].Price)
	require.Equal(t, "USD", parsed[0].Currency)
	require.Equal(t, 4.3, parsed[0].Rating)
	require.Equal(t, 37.0, parsed[0].Latitude)
	require.Equal(t, -122.0, parsed[0].Longitude)
	require.NotEmpty(t, parsed[0].Images)
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}
