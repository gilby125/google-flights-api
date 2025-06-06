// This example gets best offers in the provided date range and print the cheapest one.
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

func getCheapesOffer(
	rangeStartDate, rangeEndDate time.Time,
	tripLength int,
	srcCity, dstCity string,
	lang language.Tag,
) {
	session, err := flights.New()
	if err != nil {
		log.Fatal(err)
	}

	options := flights.Options{
		Travelers: flights.Travelers{Adults: 1},
		Currency:  currency.PLN,
		Stops:     flights.AnyStops,
		Class:     flights.Economy,
		TripType:  flights.RoundTrip,
		Lang:      lang,
	}

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
		log.Fatal(err)
	}

	var bestOffer flights.Offer
	for _, o := range offers {
		if o.Price != 0 && (bestOffer.Price == 0 || o.Price < bestOffer.Price) {
			bestOffer = o
		}
	}

	fmt.Printf("%s %s\n", bestOffer.StartDate, bestOffer.ReturnDate)
	fmt.Printf("price %d\n", int(bestOffer.Price))
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
		log.Fatal(err)
	}
	fmt.Println(url)
}

func main() {
	getCheapesOffer(
		time.Now().AddDate(0, 0, 60),
		time.Now().AddDate(0, 0, 90),
		2,
		"Warsaw",
		"Athens",
		language.English,
	)
}
