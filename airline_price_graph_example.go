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

func getBestPrices(
	rangeStartDate, rangeEndDate time.Time,
	tripLengths []int,
	srcCity, dstCity string,
) {
	session, err := flights.New()
	if err != nil {
		log.Fatal(err)
	}

	options := flights.Options{
		Travelers: flights.Travelers{Adults: 1},
		Currency:  currency.USD,
		Stops:     flights.AnyStops,
		Class:     flights.Economy,
		TripType:  flights.RoundTrip,
		Lang:      language.English,
	}

	bestOffer := flights.Offer{}

	for _, tripLength := range tripLengths {
		offers, err := session.GetPriceGraph(
			context.Background(),
			flights.PriceGraphArgs{
				RangeStartDate: rangeStartDate,
				RangeEndDate:   rangeEndDate,
				TripLength:     tripLength,
				SrcCities:      []string{srcCity},
				DstCities:      []string{dstCity},
				Options:        options,
			},
		)
		if err != nil {
			log.Printf("Error getting price graph for trip length %d: %v\n", tripLength, err)
			continue
		}

		for _, o := range offers {
			if bestOffer.Price == 0 || o.Price < bestOffer.Price {
				bestOffer = o
			}
		}
	}

	if bestOffer.Price == 0 {
		fmt.Println("No offers found.")
		return
	}

	fmt.Printf("Best price for flights from %s to %s:\n", srcCity, dstCity)
	fmt.Printf("Departure: %s\n", bestOffer.StartDate.Format("2006-01-02"))
	fmt.Printf("Return: %s\n", bestOffer.ReturnDate.Format("2006-01-02"))
	fmt.Printf("Price: $%.2f\n", bestOffer.Price)
	fmt.Printf("Trip Length: %d days\n", int(bestOffer.ReturnDate.Sub(bestOffer.StartDate).Hours()/24))

	url, err := session.SerializeURL(
		context.Background(),
		flights.Args{
			Date:       bestOffer.StartDate,
			ReturnDate: bestOffer.ReturnDate,
			SrcCities:  []string{srcCity},
			DstCities:  []string{dstCity},
			Options:    options,
		},
	)
	if err != nil {
		log.Printf("Error generating URL: %v\n", err)
	} else {
		fmt.Printf("URL: %s\n", url)
	}
}

func main() {
	getBestPrices(
		time.Now().AddDate(0, 1, 0), // Start date: 1 month from now
		time.Now().AddDate(0, 2, 0), // End date: 2 months from now
		[]int{3, 7, 5, 10, 9, 14},   // Trip lengths: 7, 10, and 14 days
		"New York",
		"London",
	)
}
