package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gilby125/google-flights-api/flights"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

const dateLayout = "2006-01-02"

type flightInfo struct {
	DepAirport string `json:"dep_airport"`
	ArrAirport string `json:"arr_airport"`
	DepTime    string `json:"dep_time"`
	ArrTime    string `json:"arr_time"`
	Duration   string `json:"duration"`
	Airline    string `json:"airline"`
	FlightNum  string `json:"flight_num"`
}

type formattedOffer struct {
	Price          float64      `json:"price"`
	Currency       string       `json:"currency"`
	StartDate      string       `json:"start_date"`
	ReturnDate     string       `json:"return_date"`
	FlightDuration string       `json:"flight_duration"`
	Airlines       []string     `json:"airlines"`
	Flights        []flightInfo `json:"flights"`
}

type formattedGraphOffer struct {
	StartDate  string  `json:"start_date"`
	ReturnDate string  `json:"return_date"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency"`
}

type segmentInput struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
}

func main() {
	session, err := flights.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing flights session: %v\n", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"google-flights-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	searchFlightsTool := mcp.NewTool("search_flights",
		mcp.WithDescription("Search for flights using Google Flights (one-way, round-trip, or multi-city)"),
		mcp.WithString("origin", mcp.Description("Origin airport code (e.g., SFO, LHR)")),
		mcp.WithString("destination", mcp.Description("Destination airport code (e.g., JFK, CDG)")),
		mcp.WithString("date", mcp.Description("Departure date (YYYY-MM-DD)")),
		mcp.WithString("return_date", mcp.Description("Return date (YYYY-MM-DD) for round trips")),
		mcp.WithString("segments", mcp.Description("JSON array of segments for multi-city. Example: '[{\"origin\":\"SFO\",\"destination\":\"JFK\",\"date\":\"2026-06-01\"}]'")),
		mcp.WithNumber("adults", mcp.Description("Number of adults (default 1)")),
		mcp.WithString("currency", mcp.Description("Currency code (e.g., USD, EUR). Default USD.")),
		mcp.WithString("carriers", mcp.Description("Comma-separated IATA airline codes and/or alliance tokens (best-effort). Example: 'UA,DL' or 'STAR_ALLIANCE'.")),
		mcp.WithString("trip_type", mcp.Description("Trip type: 'round_trip', 'one_way', or 'multi_city'. Default: round_trip if return_date is provided, else one_way.")),
	)

	s.AddTool(searchFlightsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argsMap, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		origin, _ := argsMap["origin"].(string)
		destination, _ := argsMap["destination"].(string)
		dateStr, _ := argsMap["date"].(string)
		returnDateStr, _ := argsMap["return_date"].(string)
		segmentsStr, _ := argsMap["segments"].(string)
		carriersStr, _ := argsMap["carriers"].(string)

		adultsVal, _ := argsMap["adults"].(float64)
		adults := int(adultsVal)
		if adults <= 0 {
			adults = 1
		}

		currencyStr, _ := argsMap["currency"].(string)
		if currencyStr == "" {
			currencyStr = "USD"
		}

		tripTypeStr, _ := argsMap["trip_type"].(string)

		var carriers []string
		if carriersStr != "" {
			for _, c := range strings.Split(carriersStr, ",") {
				token := strings.TrimSpace(c)
				if token == "" {
					continue
				}
				carriers = append(carriers, token)
			}
		}

		var date time.Time
		if dateStr != "" {
			parsed, err := time.Parse(dateLayout, dateStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid date format: %v", err)), nil
			}
			date = parsed
		}

		var returnDate time.Time
		if returnDateStr != "" {
			parsed, err := time.Parse(dateLayout, returnDateStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid return_date format: %v", err)), nil
			}
			returnDate = parsed
		}

		var tripType flights.TripType
		if tripTypeStr != "" {
			switch strings.ToLower(strings.TrimSpace(tripTypeStr)) {
			case "one_way":
				tripType = flights.OneWay
			case "round_trip":
				tripType = flights.RoundTrip
			case "multi_city":
				tripType = flights.MultiCity
			default:
				return mcp.NewToolResultError(fmt.Sprintf("Invalid trip_type: %s", tripTypeStr)), nil
			}
		} else {
			if !returnDate.IsZero() {
				tripType = flights.RoundTrip
			} else {
				tripType = flights.OneWay
			}
		}

		currUnit, err := currency.ParseISO(currencyStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid currency code: %v", err)), nil
		}

		options := flights.OptionsDefault()
		options.Travelers = flights.Travelers{Adults: adults}
		options.Currency = currUnit
		options.Stops = flights.AnyStops
		options.Class = flights.Economy
		options.TripType = tripType
		options.Lang = language.English
		options.Carriers = carriers

		searchArgs := flights.Args{
			Date:        date,
			ReturnDate:  returnDate,
			SrcAirports: []string{origin},
			DstAirports: []string{destination},
			Options:     options,
		}

		if tripType == flights.MultiCity {
			if segmentsStr == "" {
				return mcp.NewToolResultError("segments is required for trip_type=multi_city"), nil
			}
			var segments []segmentInput
			if err := json.Unmarshal([]byte(segmentsStr), &segments); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid segments JSON: %v", err)), nil
			}
			if len(segments) < 1 {
				return mcp.NewToolResultError("segments must have at least 1 segment"), nil
			}

			searchArgs.SrcAirports = nil
			searchArgs.DstAirports = nil
			searchArgs.Date = time.Time{}
			searchArgs.ReturnDate = time.Time{}

			for i, seg := range segments {
				segDate, err := time.Parse(dateLayout, seg.Date)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Invalid segment %d date: %v", i, err)), nil
				}
				searchArgs.Segments = append(searchArgs.Segments, flights.Segment{
					Date:        segDate,
					SrcAirports: []string{strings.TrimSpace(seg.Origin)},
					DstAirports: []string{strings.TrimSpace(seg.Destination)},
				})
			}
		} else {
			if origin == "" || destination == "" || dateStr == "" {
				return mcp.NewToolResultError("origin, destination, and date are required for one_way and round_trip"), nil
			}
		}

		offers, priceRange, err := session.GetOffers(ctx, searchArgs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error searching flights: %v", err)), nil
		}

		searchURL, err := session.SerializeURL(ctx, searchArgs)
		if err != nil {
			searchURL = ""
		}

		formattedOffers := make([]formattedOffer, 0, len(offers))
		for _, offer := range offers {
			airlineNames := make([]string, 0, len(offer.Flight))
			flightsInfo := make([]flightInfo, 0, len(offer.Flight))

			for _, f := range offer.Flight {
				airlineNames = append(airlineNames, f.AirlineName)
				flightsInfo = append(flightsInfo, flightInfo{
					DepAirport: f.DepAirportCode,
					ArrAirport: f.ArrAirportCode,
					DepTime:    f.DepTime.Format(time.RFC3339),
					ArrTime:    f.ArrTime.Format(time.RFC3339),
					Duration:   f.Duration.String(),
					Airline:    f.AirlineName,
					FlightNum:  f.FlightNumber,
				})
			}

			formattedOffers = append(formattedOffers, formattedOffer{
				Price:          offer.Price,
				Currency:       currencyStr,
				StartDate:      offer.StartDate.Format(dateLayout),
				ReturnDate:     offer.ReturnDate.Format(dateLayout),
				FlightDuration: offer.FlightDuration.String(),
				Airlines:       uniqueStrings(airlineNames),
				Flights:        flightsInfo,
			})
		}

		resp := map[string]any{
			"offers":      formattedOffers,
			"price_range": priceRange,
			"search_url":  searchURL,
		}

		jsonBytes, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	getPriceGraphTool := mcp.NewTool("get_price_graph",
		mcp.WithDescription("Get price graph data (calendar graph) for a date range (round-trip only)"),
		mcp.WithString("origin", mcp.Description("Origin airport code (e.g., SFO)"), mcp.Required()),
		mcp.WithString("destination", mcp.Description("Destination airport code (e.g., CDG)"), mcp.Required()),
		mcp.WithString("range_start_date", mcp.Description("Start date of the range (YYYY-MM-DD)"), mcp.Required()),
		mcp.WithString("range_end_date", mcp.Description("End date of the range (YYYY-MM-DD)"), mcp.Required()),
		mcp.WithNumber("trip_length", mcp.Description("Trip length in days (default 7)")),
		mcp.WithString("currency", mcp.Description("Currency code (default USD)")),
		mcp.WithString("carriers", mcp.Description("Comma-separated IATA carrier codes/alliance tokens to include (best-effort)")),
	)

	s.AddTool(getPriceGraphTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argsMap, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		origin, _ := argsMap["origin"].(string)
		destination, _ := argsMap["destination"].(string)
		rangeStartDateStr, _ := argsMap["range_start_date"].(string)
		rangeEndDateStr, _ := argsMap["range_end_date"].(string)
		carriersStr, _ := argsMap["carriers"].(string)

		tripLengthVal, _ := argsMap["trip_length"].(float64)
		tripLength := int(tripLengthVal)
		if tripLength <= 0 {
			tripLength = 7
		}

		currencyStr, _ := argsMap["currency"].(string)
		if currencyStr == "" {
			currencyStr = "USD"
		}

		var carriers []string
		if carriersStr != "" {
			for _, c := range strings.Split(carriersStr, ",") {
				token := strings.TrimSpace(c)
				if token == "" {
					continue
				}
				carriers = append(carriers, token)
			}
		}

		rangeStartDate, err := time.Parse(dateLayout, rangeStartDateStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid range_start_date: %v", err)), nil
		}
		rangeEndDate, err := time.Parse(dateLayout, rangeEndDateStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid range_end_date: %v", err)), nil
		}

		currUnit, err := currency.ParseISO(currencyStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid currency code: %v", err)), nil
		}

		options := flights.OptionsDefault()
		options.Currency = currUnit
		options.Lang = language.English
		options.Carriers = carriers

		pgArgs := flights.PriceGraphArgs{
			RangeStartDate: rangeStartDate,
			RangeEndDate:   rangeEndDate,
			TripLength:     tripLength,
			SrcAirports:    []string{origin},
			DstAirports:    []string{destination},
			Options:        options,
		}

		offers, _, err := session.GetPriceGraph(ctx, pgArgs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting price graph: %v", err)), nil
		}

		formattedOffers := make([]formattedGraphOffer, 0, len(offers))
		minPrice := float64(0)
		var bestOffer *flights.Offer

		for i := range offers {
			o := offers[i]
			if o.Price > 0 && (minPrice == 0 || o.Price < minPrice) {
				minPrice = o.Price
				bestOffer = &o
			}
			formattedOffers = append(formattedOffers, formattedGraphOffer{
				StartDate:  o.StartDate.Format(dateLayout),
				ReturnDate: o.ReturnDate.Format(dateLayout),
				Price:      o.Price,
				Currency:   currencyStr,
			})
		}

		bestAirline := ""
		if bestOffer != nil {
			sampleArgs := flights.Args{
				Date:        bestOffer.StartDate,
				ReturnDate:  bestOffer.ReturnDate,
				SrcAirports: []string{origin},
				DstAirports: []string{destination},
				Options:     options,
			}
			sampleOffers, _, err := session.GetOffers(ctx, sampleArgs)
			if err == nil && len(sampleOffers) > 0 && len(sampleOffers[0].Flight) > 0 {
				bestAirline = sampleOffers[0].Flight[0].AirlineName
			}
		}

		resp := map[string]any{
			"offers":             formattedOffers,
			"best_price":         minPrice,
			"best_price_airline": bestAirline,
		}

		jsonBytes, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, entry := range input {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		out = append(out, entry)
	}
	return out
}
