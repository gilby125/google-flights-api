package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/gilby125/google-flights-api/flights"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"

	"github.com/gilby125/google-flights-api/db"
)

type Worker struct {
	postgresDB *db.PostgresDB
	neo4jDB    *db.Neo4jDB
}

type FlightSearchPayload struct {
	Origin        string
	Destination   string
	DepartureDate time.Time
	ReturnDate    time.Time
	Adults        int
	Children      int
	InfantsLap    int
	InfantsSeat   int
	TripType      string
	Class         string
	Stops         string
	Currency      string
}

type BulkSearchPayload struct {
	Origins           []string
	Destinations      []string
	DepartureDateFrom time.Time
	DepartureDateTo   time.Time
	ReturnDateFrom    time.Time
	ReturnDateTo      time.Time
	TripLength        int
	Adults            int
	Children          int
	InfantsLap        int
	InfantsSeat       int
	TripType          string
	Class             string
	Stops             string
	Currency          string
}

// storeFlightOffers stores flight offers in the database
func (w *Worker) storeFlightOffers(ctx context.Context, payload FlightSearchPayload, offers []flights.FullOffer, priceRange *flights.PriceRange) error {
	// Begin a transaction
	tx, err := w.postgresDB.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the search query
	var queryID int
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO search_queries 
		(origin, destination, departure_date, return_date, adults, children, infants_lap, infants_seat, trip_type, class, stops, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
		RETURNING id`,
		payload.Origin,
		payload.Destination,
		payload.DepartureDate,
		payload.ReturnDate,
		payload.Adults,
		payload.Children,
		payload.InfantsLap,
		payload.InfantsSeat,
		payload.TripType,
		payload.Class,
		payload.Stops,
		"completed",
	).Scan(&queryID)

	if err != nil {
		return fmt.Errorf("failed to insert search query: %w", err)
	}

	// Store each offer
	for _, offer := range offers {
		// Insert the flight offer
		var offerID int
		totalDuration := int(offer.FlightDuration.Minutes())
		err = tx.QueryRowContext(
			ctx,
			`INSERT INTO flight_offers 
			(search_query_id, price, currency, departure_date, return_date, total_duration) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id`,
			queryID,
			offer.Price,
			payload.Currency,
			offer.StartDate,
			offer.ReturnDate,
			totalDuration,
		).Scan(&offerID)

		if err != nil {
			return fmt.Errorf("failed to insert flight offer: %w", err)
		}

		// Store flight segments
		for _, flight := range offer.Flight {
			// Ensure airline exists
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO airlines (code, name, country) 
				VALUES ($1, $2, $3) 
				ON CONFLICT (code) DO UPDATE SET name = $2`,
				flight.FlightNumber[:2], // Airline code is typically the first two characters of flight number
				flight.AirlineName,
				"", // We don't have country info from the API
			)
			if err != nil {
				return fmt.Errorf("failed to insert airline: %w", err)
			}

			// Ensure airports exist
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO airports (code, name, city, country, latitude, longitude) 
				VALUES ($1, $2, $3, $4, $5, $6) 
				ON CONFLICT (code) DO UPDATE SET name = $2, city = $3`,
				flight.DepAirportCode,
				flight.DepAirportName,
				flight.DepCity,
				"",  // We don't have country info from the API
				0.0, // We don't have coordinates from the API
				0.0,
			)
			if err != nil {
				return fmt.Errorf("failed to insert departure airport: %w", err)
			}

			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO airports (code, name, city, country, latitude, longitude) 
				VALUES ($1, $2, $3, $4, $5, $6) 
				ON CONFLICT (code) DO UPDATE SET name = $2, city = $3`,
				flight.ArrAirportCode,
				flight.ArrAirportName,
				flight.ArrCity,
				"",  // We don't have country info from the API
				0.0, // We don't have coordinates from the API
				0.0,
			)
			if err != nil {
				return fmt.Errorf("failed to insert arrival airport: %w", err)
			}

			// Insert flight segment
			duration := int(flight.Duration.Minutes())
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO flight_segments 
				(flight_offer_id, airline_code, flight_number, departure_airport, arrival_airport, 
				departure_time, arrival_time, duration, airplane, legroom, is_return) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
				offerID,
				flight.FlightNumber[:2], // Airline code
				flight.FlightNumber,
				flight.DepAirportCode,
				flight.ArrAirportCode,
				flight.DepTime,
				flight.ArrTime,
				duration,
				flight.Airplane,
				flight.Legroom,
				false, // Not a return flight
			)
			if err != nil {
				return fmt.Errorf("failed to insert flight segment: %w", err)
			}
		}

		// Store in Neo4j for graph analysis
		if err := w.storeFlightInNeo4j(ctx, offer); err != nil {
			return fmt.Errorf("failed to store flight in Neo4j: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// storeFlightInNeo4j stores flight data in Neo4j for graph analysis
func (w *Worker) storeFlightInNeo4j(ctx context.Context, offer flights.FullOffer) error {
	// Create airports in Neo4j
	for _, flight := range offer.Flight {
		// Create departure airport
		if err := w.neo4jDB.CreateAirport(
			flight.DepAirportCode,
			flight.DepAirportName,
			flight.DepCity,
			"",  // We don't have country info from the API
			0.0, // We don't have coordinates from the API
			0.0,
		); err != nil {
			return fmt.Errorf("failed to create departure airport in Neo4j: %w", err)
		}

		// Create arrival airport
		if err := w.neo4jDB.CreateAirport(
			flight.ArrAirportCode,
			flight.ArrAirportName,
			flight.ArrCity,
			"",  // We don't have country info from the API
			0.0, // We don't have coordinates from the API
			0.0,
		); err != nil {
			return fmt.Errorf("failed to create arrival airport in Neo4j: %w", err)
		}

		// Create airline
		if err := w.neo4jDB.CreateAirline(
			flight.FlightNumber[:2], // Airline code
			flight.AirlineName,
			"", // We don't have country info from the API
		); err != nil {
			return fmt.Errorf("failed to create airline in Neo4j: %w", err)
		}

		// Create route
		if err := w.neo4jDB.CreateRoute(
			flight.DepAirportCode,
			flight.ArrAirportCode,
			flight.FlightNumber[:2], // Airline code
			flight.FlightNumber,
			offer.Price,
			int(flight.Duration.Minutes()),
		); err != nil {
			return fmt.Errorf("failed to create route in Neo4j: %w", err)
		}

		// Add price point
		if err := w.neo4jDB.AddPricePoint(
			flight.DepAirportCode,
			flight.ArrAirportCode,
			offer.StartDate.Format("2006-01-02"),
			offer.Price,
			flight.FlightNumber[:2], // Airline code
		); err != nil {
			return fmt.Errorf("failed to add price point in Neo4j: %w", err)
		}
	}

	return nil
}

// processPriceGraphSearch uses the price graph API to find the best dates for a trip
func (w *Worker) processPriceGraphSearch(ctx context.Context, session *flights.Session, origin, destination string, payload BulkSearchPayload) error {
	// Map the payload to flights API arguments
	var tripType flights.TripType
	switch payload.TripType {
	case "one_way":
		tripType = flights.OneWay
	case "round_trip":
		tripType = flights.RoundTrip
	default:
		return fmt.Errorf("invalid trip type: %s", payload.TripType)
	}

	var class flights.Class
	switch payload.Class {
	case "economy":
		class = flights.Economy
	case "premium_economy":
		class = flights.PremiumEconomy
	case "business":
		class = flights.Business
	case "first":
		class = flights.First
	default:
		return fmt.Errorf("invalid class: %s", payload.Class)
	}

	var stops flights.Stops
	switch payload.Stops {
	case "nonstop":
		stops = flights.Nonstop
	case "one_stop":
		stops = flights.Stop1
	case "two_stops":
		stops = flights.Stop2
	case "any":
		stops = flights.AnyStops
	default:
		return fmt.Errorf("invalid stops: %s", payload.Stops)
	}

	// Parse currency
	cur, err := currency.ParseISO(payload.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %s", payload.Currency)
	}

	// Get price graph data
	offers, err := session.GetPriceGraph(
		ctx,
		flights.PriceGraphArgs{
			RangeStartDate: payload.DepartureDateFrom,
			RangeEndDate:   payload.DepartureDateTo,
			TripLength:     payload.TripLength,
			SrcAirports:    []string{origin},
			DstAirports:    []string{destination},
			Options: flights.Options{
				Travelers: flights.Travelers{
					Adults:       payload.Adults,
					Children:     payload.Children,
					InfantOnLap:  payload.InfantsLap,
					InfantInSeat: payload.InfantsSeat,
				},
				Currency: cur,
				Stops:    stops,
				Class:    class,
				TripType: tripType,
				Lang:     language.English,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get price graph data: %w", err)
	}

	// Convert Offer to FullOffer
	fullOffers := make([]flights.FullOffer, 0, len(offers))
	for _, offer := range offers {
		fullOffer := flights.FullOffer{
			Offer:          offer,
			SrcAirportCode: origin,
			DstAirportCode: destination,
			// We don't have flight details from price graph, so we'll leave other fields empty
			Flight:         []flights.Flight{},
			FlightDuration: 0,
		}
		fullOffers = append(fullOffers, fullOffer)
	}

	// Process the results
	if err := w.storeFlightOffers(ctx, FlightSearchPayload{
		Origin:        origin,
		Destination:   destination,
		DepartureDate: payload.DepartureDateFrom,
		ReturnDate:    payload.DepartureDateTo,
		Adults:        payload.Adults,
		Children:      payload.Children,
		InfantsLap:    payload.InfantsLap,
		InfantsSeat:   payload.InfantsSeat,
		TripType:      payload.TripType,
		Class:         payload.Class,
		Stops:         payload.Stops,
		Currency:      payload.Currency,
	}, fullOffers, nil); err != nil {
		return fmt.Errorf("failed to store flight offers: %w", err)
	}

	return nil
}
