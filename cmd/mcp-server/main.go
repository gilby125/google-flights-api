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

func main() {
	// Initialize flights session
	session, err := flights.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing flights session: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"google-flights-mcp",
		"1.0.0",
		server.WithLogging(),
	)

	// Register search_flights tool
	tool := mcp.NewTool("search_flights",
		mcp.WithDescription("Search for flights using Google Flights"),
		mcp.WithString("origin",
			mcp.Description("Origin airport code (e.g., SFO, LHR)"),
		),
		mcp.WithString("destination",
			mcp.Description("Destination airport code (e.g., JFK, CDG)"),
		),
		mcp.WithString("date",
			mcp.Description("Departure date (YYYY-MM-DD)"),
		),
		mcp.WithString("return_date",
			mcp.Description("Return date (YYYY-MM-DD) for round trips. Optional for one-way."),
		),
		mcp.WithString("segments",
			mcp.Description("JSON array of segments for multi-city trips. Each segment should have 'origin', 'destination', 'date'. Example: '[{\"origin\":\"SFO\",\"destination\":\"JFK\",\"date\":\"2026-06-01\"}]'"),
		),
		mcp.WithNumber("adults",
			mcp.Description("Number of adults (default 1)"),
		),
		mcp.WithString("currency",
			mcp.Description("Currency code (e.g., USD, EUR). Default USD."),
		),
		mcp.WithString("trip_type",
			mcp.Description("Trip type: 'round_trip', 'one_way', or 'multi_city'. Default is round_trip if return_date is provided, else one_way."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argsMap, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		origin, _ := argsMap["origin"].(string)
		destination, _ := argsMap["destination"].(string)
		dateStr, _ := argsMap["date"].(string)
		returnDateStr, _ := argsMap["return_date"].(string)
		segmentsStr, _ := argsMap["segments"].(string)
		
		adultsVal, _ := argsMap["adults"].(float64)
		adults := int(adultsVal)
		if adults == 0 {
			adults = 1
		}

		currencyStr, _ := argsMap["currency"].(string)
		if currencyStr == "" {
			currencyStr = "USD"
		}

		tripTypeStr, _ := argsMap["trip_type"].(string)

		// Parse dates
		var date time.Time
		var err error
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid date format: %v", err)), nil
			}
		}

		var returnDate time.Time
		if returnDateStr != "" {
			returnDate, err = time.Parse("2006-01-02", returnDateStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid return_date format: %v", err)), nil
			}
		}

		// Determine TripType
		var tripType flights.TripType
		if tripTypeStr != "" {
			switch strings.ToLower(tripTypeStr) {
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

		// Parse Currency
		currUnit, err := currency.ParseISO(currencyStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid currency code: %v", err)), nil
		}

		// Build Args
		args := flights.Args{
			Date:        date,
			ReturnDate:  returnDate,
			SrcAirports: []string{origin},
			DstAirports: []string{destination},
			Options: flights.Options{
				Travelers: flights.Travelers{Adults: adults},
				Currency:  currUnit,
				Stops:     flights.AnyStops,
				Class:     flights.Economy,
				TripType:  tripType,
				Lang:      language.English,
			},
		}

		if tripType == flights.MultiCity {
			if segmentsStr == "" {
				return mcp.NewToolResultError("segments argument is required for multi_city trip_type"), nil
			}
			var segments []struct {
				Origin      string `json:"origin"`
				Destination string `json:"destination"`
				Date        string `json:"date"`
			}
			if err := json.Unmarshal([]byte(segmentsStr), &segments); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid segments JSON: %v", err)), nil
			}

			for _, seg := range segments {
				t, err := time.Parse("2006-01-02", seg.Date)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Invalid date in segment: %v", err)), nil
				}
				args.Segments = append(args.Segments, flights.Segment{
					Date:        t,
					SrcAirports: []string{seg.Origin},
					DstAirports: []string{seg.Destination},
				})
			}
			// Clear top-level src/dst to avoid confusion, or leave them as empty (they are slices)
			args.SrcAirports = nil
			args.DstAirports = nil
		} else {
			if origin == "" || destination == "" || dateStr == "" {
				return mcp.NewToolResultError("origin, destination, and date are required for one_way and round_trip"), nil
			}
		}

		// Call GetOffers
		offers, priceRange, err := session.GetOffers(ctx, args)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error searching flights: %v", err)), nil
		}

		// Format results
		type FormattedOffer struct {
			Price          float64       `json:"price"`
			Currency       string        `json:"currency"`
			StartDate      string        `json:"start_date"`
			ReturnDate     string        `json:"return_date"`
			FlightDuration string        `json:"flight_duration"`
			Airlines       []string      `json:"airlines"`
			Flights        []FlightInfo  `json:"flights"`
		}

		var formattedOffers []FormattedOffer
		for _, offer := range offers {
			var airlineNames []string
			var flightsInfo []FlightInfo

			for _, f := range offer.Flight {
				airlineNames = append(airlineNames, f.AirlineName)
				flightsInfo = append(flightsInfo, FlightInfo{
					DepAirport: f.DepAirportCode,
					ArrAirport: f.ArrAirportCode,
					DepTime:    f.DepTime.Format(time.RFC3339),
					ArrTime:    f.ArrTime.Format(time.RFC3339),
					Duration:   f.Duration.String(),
					Airline:    f.AirlineName,
					FlightNum:  f.FlightNumber,
				})
			}

			formattedOffers = append(formattedOffers, FormattedOffer{
				Price:          offer.Price,
				Currency:       currencyStr,
				StartDate:      offer.StartDate.Format("2006-01-02"),
				ReturnDate:     offer.ReturnDate.Format("2006-01-02"),
				FlightDuration: offer.FlightDuration.String(),
				Airlines:       uniqueStrings(airlineNames),
				Flights:        flightsInfo,
			})
		}

		response := map[string]interface{}{
			"offers":      formattedOffers,
			"price_range": priceRange,
		}

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}

type FlightInfo struct {
	DepAirport string `json:"dep_airport"`
	ArrAirport string `json:"arr_airport"`
	DepTime    string `json:"dep_time"`
	ArrTime    string `json:"arr_time"`
	Duration   string `json:"duration"`
	Airline    string `json:"airline"`
	FlightNum  string `json:"flight_num"`
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
