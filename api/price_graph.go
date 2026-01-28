package api

import (
	"fmt"
	"sort"
	"time"

	"github.com/gilby125/google-flights-api/flights"
)

const maxPriceGraphRangeDays = 161

type PriceGraphBuildParams struct {
	Include           bool
	WindowDays        int
	DepartureDateFrom string
	DepartureDateTo   string
	TripLengthDays    int
}

func TopPriceGraphPoints(points []map[string]interface{}, n int) []map[string]interface{} {
	if n <= 0 || len(points) == 0 {
		return nil
	}

	type scored struct {
		point map[string]interface{}
		price float64
		ok    bool
	}

	scoredPoints := make([]scored, 0, len(points))
	for _, p := range points {
		if p == nil {
			continue
		}
		priceRaw, ok := p["price"]
		if !ok {
			continue
		}
		var price float64
		switch v := priceRaw.(type) {
		case float64:
			price = v
		case float32:
			price = float64(v)
		case int:
			price = float64(v)
		case int64:
			price = float64(v)
		case int32:
			price = float64(v)
		case uint:
			price = float64(v)
		case uint64:
			price = float64(v)
		case uint32:
			price = float64(v)
		default:
			continue
		}
		if price <= 0 {
			continue
		}
		scoredPoints = append(scoredPoints, scored{point: p, price: price, ok: true})
	}

	if len(scoredPoints) == 0 {
		return nil
	}

	sort.SliceStable(scoredPoints, func(i, j int) bool {
		return scoredPoints[i].price < scoredPoints[j].price
	})

	if n > len(scoredPoints) {
		n = len(scoredPoints)
	}

	out := make([]map[string]interface{}, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, scoredPoints[i].point)
	}
	return out
}

func BuildPriceGraphArgs(now time.Time, origin, destination string, departureDate, returnDate time.Time, opts flights.Options, params PriceGraphBuildParams) (flights.PriceGraphArgs, error) {
	if !params.Include {
		return flights.PriceGraphArgs{}, fmt.Errorf("price graph not enabled")
	}

	tripLength := params.TripLengthDays
	if tripLength <= 0 {
		tripLength = int(returnDate.Sub(departureDate).Hours() / 24)
		if tripLength <= 0 {
			switch opts.TripType {
			case flights.OneWay:
				tripLength = 1
			default:
				tripLength = 7
			}
		}
	}

	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var rangeStart time.Time
	var rangeEnd time.Time

	switch {
	case params.DepartureDateFrom != "" || params.DepartureDateTo != "":
		if params.DepartureDateFrom == "" || params.DepartureDateTo == "" {
			return flights.PriceGraphArgs{}, fmt.Errorf("both price_graph_departure_date_from and price_graph_departure_date_to are required when specifying a price-graph range")
		}

		parsedFrom, err := time.Parse(dateLayout, params.DepartureDateFrom)
		if err != nil {
			return flights.PriceGraphArgs{}, fmt.Errorf("invalid price_graph_departure_date_from (expected YYYY-MM-DD): %w", err)
		}
		parsedTo, err := time.Parse(dateLayout, params.DepartureDateTo)
		if err != nil {
			return flights.PriceGraphArgs{}, fmt.Errorf("invalid price_graph_departure_date_to (expected YYYY-MM-DD): %w", err)
		}

		rangeStart = parsedFrom
		rangeEnd = parsedTo

	default:
		windowDays := params.WindowDays
		if windowDays <= 0 {
			windowDays = 30
		}
		if windowDays > maxPriceGraphRangeDays {
			windowDays = maxPriceGraphRangeDays
		}
		if windowDays < 2 {
			windowDays = 2
		}

		half := windowDays / 2
		rangeStart = departureDate.AddDate(0, 0, -half)
		rangeEnd = rangeStart.AddDate(0, 0, windowDays)

		if rangeStart.Before(nowDay) {
			rangeStart = nowDay
			rangeEnd = rangeStart.AddDate(0, 0, windowDays)
		}
	}

	// Ensure start < end and cap at max range.
	rangeStart = time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 0, 0, 0, 0, time.UTC)
	rangeEnd = time.Date(rangeEnd.Year(), rangeEnd.Month(), rangeEnd.Day(), 0, 0, 0, 0, time.UTC)

	if rangeStart.Before(nowDay) {
		rangeStart = nowDay
	}

	if !rangeEnd.After(rangeStart) {
		rangeEnd = rangeStart.AddDate(0, 0, 1)
	}

	maxEnd := rangeStart.AddDate(0, 0, maxPriceGraphRangeDays)
	if rangeEnd.After(maxEnd) {
		rangeEnd = maxEnd
	}

	return flights.PriceGraphArgs{
		RangeStartDate: rangeStart,
		RangeEndDate:   rangeEnd,
		TripLength:     tripLength,
		SrcAirports:    []string{origin},
		DstAirports:    []string{destination},
		Options:        opts,
	}, nil
}

func SerializePriceGraphResponse(origin, destination, currency string, args flights.PriceGraphArgs, offers []flights.Offer, parseErrors *flights.ParseErrors, err error) map[string]interface{} {
	out := map[string]interface{}{
		"origin":      origin,
		"destination": destination,
		"currency":    currency,
	}

	if !args.RangeStartDate.IsZero() {
		out["range_start_date"] = args.RangeStartDate.Format(dateLayout)
	}
	if !args.RangeEndDate.IsZero() {
		out["range_end_date"] = args.RangeEndDate.Format(dateLayout)
	}
	if args.TripLength > 0 {
		out["trip_length_days"] = args.TripLength
	}

	if err != nil {
		out["error"] = err.Error()
		return out
	}

	if parseErrors != nil {
		out["parse_errors"] = parseErrors
	}

	points := make([]map[string]interface{}, 0, len(offers))
	for _, offer := range offers {
		point := map[string]interface{}{
			"departure_date": offer.StartDate.Format(dateLayout),
			"return_date":    offer.ReturnDate.Format(dateLayout),
			"price":          offer.Price,
		}
		points = append(points, point)
	}
	out["points"] = points
	return out
}
