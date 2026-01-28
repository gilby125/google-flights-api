package api

import (
	"fmt"
	"strings"
	"time"
)

type DirectSearchDatePlan struct {
	DepartureDate  time.Time
	ReturnDate     time.Time
	PriceGraphOnly bool
	PriceGraph     PriceGraphBuildParams
}

func PlanDirectSearchDates(now time.Time, req DirectSearchRequest) (DirectSearchDatePlan, error) {
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	priceGraphParams := PriceGraphBuildParams{
		Include:           req.IncludePriceGraph,
		WindowDays:        req.PriceGraphWindowDays,
		DepartureDateFrom: strings.TrimSpace(req.PriceGraphDepartureDateFrom),
		DepartureDateTo:   strings.TrimSpace(req.PriceGraphDepartureDateTo),
		TripLengthDays:    req.PriceGraphTripLengthDays,
	}

	hasDeparture := strings.TrimSpace(req.DepartureDate) != ""
	hasReturn := strings.TrimSpace(req.ReturnDate) != ""

	if !hasDeparture {
		if hasReturn {
			return DirectSearchDatePlan{}, fmt.Errorf("return_date requires departure_date")
		}
		if !req.IncludePriceGraph {
			return DirectSearchDatePlan{}, fmt.Errorf("departure_date is required unless include_price_graph is true")
		}

		tripLength := req.PriceGraphTripLengthDays
		if tripLength <= 0 {
			if strings.TrimSpace(req.TripType) == "one_way" {
				tripLength = 1
			} else {
				tripLength = 7
			}
		}

		departure := nowDay
		returnDate := departure.AddDate(0, 0, tripLength)

		if priceGraphParams.DepartureDateFrom == "" && priceGraphParams.DepartureDateTo == "" {
			priceGraphParams.DepartureDateFrom = departure.Format(dateLayout)
			priceGraphParams.DepartureDateTo = departure.AddDate(0, 0, maxPriceGraphRangeDays).Format(dateLayout)
		}
		priceGraphParams.TripLengthDays = tripLength

		return DirectSearchDatePlan{
			DepartureDate:  departure,
			ReturnDate:     returnDate,
			PriceGraphOnly: true,
			PriceGraph:     priceGraphParams,
		}, nil
	}

	departureDate, err := time.Parse(dateLayout, strings.TrimSpace(req.DepartureDate))
	if err != nil {
		return DirectSearchDatePlan{}, fmt.Errorf("invalid departure date format (expected YYYY-MM-DD): %w", err)
	}

	var returnDate time.Time
	if hasReturn {
		returnDate, err = time.Parse(dateLayout, strings.TrimSpace(req.ReturnDate))
		if err != nil {
			return DirectSearchDatePlan{}, fmt.Errorf("invalid return date format (expected YYYY-MM-DD): %w", err)
		}
	} else if strings.TrimSpace(req.TripType) == "round_trip" {
		returnDate = departureDate.AddDate(0, 0, 7)
	} else {
		returnDate = departureDate.AddDate(0, 0, 1)
	}

	return DirectSearchDatePlan{
		DepartureDate:  departureDate,
		ReturnDate:     returnDate,
		PriceGraphOnly: false,
		PriceGraph:     priceGraphParams,
	}, nil
}
