package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gilby125/google-flights-api/flights"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

func main() {
	// Create a new session
	session, err := flights.New()
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	// Define search parameters
	now := time.Now()
	departureDate := now.AddDate(0, 1, 0) // 1 month from now
	returnDate := now.AddDate(0, 1, 7)    // 1 month + 7 days from now

	// Define search queries
	queries := []struct {
		origin      string
		destination string
	}{
		{"LAX", "JFK"},
		{"SFO", "MIA"},
		// Add more queries as needed
	}

	// Perform searches
	for i, query := range queries {
		offers, priceRange, err := session.GetOffers(
			context.Background(),
			flights.Args{
				Date:        departureDate,
				ReturnDate:  returnDate,
				SrcAirports: []string{query.origin},
				DstAirports: []string{query.destination},
				Options: flights.Options{
					Travelers: flights.Travelers{Adults: 1},
					Currency:  currency.USD,
					Stops:     flights.AnyStops,
					Class:     flights.Economy,
					TripType:  flights.RoundTrip,
					Lang:      language.English,
				},
			},
		)
		if err != nil {
			log.Printf("Error for query %d (%s to %s): %v",
				i+1, query.origin, query.destination, err)
			continue
		}

		fmt.Printf("Results for query %d (Origin: %s, Destination: %s):\n",
			i+1, query.origin, query.destination)

		if priceRange != nil {
			fmt.Printf("  Price range: %.2f - %.2f\n", priceRange.Low, priceRange.High)
		}

		for j, offer := range offers {
			if j >= 5 { // Limit to first 5 results
				break
			}
			fmt.Printf("  Option %d: Price: %.2f, Duration: %s\n",
				j+1, offer.Price, offer.FlightDuration)
		}
		fmt.Println()
	}
}
