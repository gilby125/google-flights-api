package hotels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

var ds0Re = regexp.MustCompile(`(?s)AF_initDataCallback\(\{key:\s*['"]ds:0['"].*?data:(.*?),\s*sideChannel:\s*\{\}\}\);`)

// SerializeURL generates a Google Hotels search URL based on the provided arguments.
func (s *Session) SerializeURL(ctx context.Context, args Args) (string, error) {
	if err := args.Validate(); err != nil {
		return "", err
	}

	u, _ := url.Parse("https://www.google.com/travel/hotels")
	q := u.Query()

	q.Set("q", args.Location)
	q.Set("checkin", args.CheckInDate.Format("2006-01-02"))
	q.Set("checkout", args.CheckOutDate.Format("2006-01-02"))
	q.Set("adults", fmt.Sprintf("%d", args.Travelers.Adults))
	if args.Travelers.Children > 0 {
		q.Set("children", fmt.Sprintf("%d", args.Travelers.Children))
	}
	q.Set("curr", args.Currency.String())
	q.Set("hl", args.Lang.String())

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// GetOffers scrapes the Google Hotels search page and returns a list of hotels.
func (s *Session) GetOffers(ctx context.Context, args Args) ([]Hotel, error) {
	urlStr, err := s.SerializeURL(ctx, args)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header["cookie"] = s.cookies

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseHotelsFromHTML(string(bodyBytes), args.Currency.String())
}

func parseHotelsFromHTML(htmlBody string, currencyCode string) ([]Hotel, error) {
	// Extract the ds:0 JSON blob.
	match := ds0Re.FindStringSubmatch(htmlBody)
	if len(match) < 2 {
		return nil, fmt.Errorf("could not find hotel data in response")
	}

	var data []any
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return nil, fmt.Errorf("failed to parse hotel data JSON: %w", err)
	}

	var findHotelsList func(interface{}, int) []interface{}
	findHotelsList = func(v interface{}, depth int) []interface{} {
		switch val := v.(type) {
		case []interface{}:
			if len(val) > 0 {
				if first, ok := val[0].([]interface{}); ok && len(first) > 5 {
					if _, ok := first[0].(string); ok {
						if _, ok := first[2].(string); ok {
							return val
						}
					}
				}
			}
			for _, item := range val {
				if res := findHotelsList(item, depth+1); res != nil {
					return res
				}
			}
		case map[string]interface{}:
			for _, item := range val {
				if res := findHotelsList(item, depth+1); res != nil {
					return res
				}
			}
		}
		return nil
	}

	hotelsData := findHotelsList(data, 0)

	if hotelsData == nil {
		return nil, fmt.Errorf("could not locate hotel list in JSON structure")
	}

	var hotels []Hotel
	for _, h := range hotelsData {
		hotelArr, ok := h.([]interface{})
		if !ok || len(hotelArr) < 17 { // Ensure enough elements
			continue
		}

		name, _ := hotelArr[0].(string)

		priceStr, _ := hotelArr[2].(string)
		price := parsePrice(priceStr)

		rating, _ := hotelArr[5].(float64)

		// Images are in index 3
		var images []string
		if imgArr, ok := hotelArr[3].([]interface{}); ok {
			for _, img := range imgArr {
				if str, ok := img.(string); ok {
					images = append(images, str)
				}
			}
		}

		// Coordinates in index 16: [lat, long]
		var lat, long float64
		if coords, ok := hotelArr[16].([]interface{}); ok && len(coords) >= 2 {
			lat, _ = coords[0].(float64)
			long, _ = coords[1].(float64)
		}

		hotels = append(hotels, Hotel{
			Name:      name,
			Price:     price,
			Currency:  currencyCode,
			Rating:    rating,
			Images:    images,
			Latitude:  lat,
			Longitude: long,
		})
	}

	return hotels, nil
}

func parsePrice(priceStr string) float64 {
	// Remove non-numeric characters (except dot)
	re := regexp.MustCompile(`[^\d.,]`)
	cleanStr := re.ReplaceAllString(priceStr, "")
	if strings.Count(cleanStr, ",") == 1 && strings.Count(cleanStr, ".") == 0 {
		cleanStr = strings.ReplaceAll(cleanStr, ",", ".")
	}
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	val, _ := strconv.ParseFloat(cleanStr, 64)
	return val
}
