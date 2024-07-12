package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/krisukox/google-flights-api/flights"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

var flightSession *flights.Session

type SearchRequest struct {
	SrcCities   []string `json:"srcCities"`
	DstCities   []string `json:"dstCities"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	TripLength  int      `json:"tripLength"`
	Airlines    []string `json:"airlines"`
	TravelClass string   `json:"travelClass"`
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	flightSession, err = flights.New()
	if err != nil {
		log.Fatalf("Failed to initialize flight session: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/search", searchHandler).Methods("POST")

	workDir, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(workDir, "static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(staticDir)))

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	workDir, _ := os.Getwd()
	filePath := filepath.Join(workDir, "static", "index.html")
	http.ServeFile(w, r, filePath)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received search request")

	var searchReq SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&searchReq); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded search request: %+v", searchReq)

	startDate, err := time.Parse("2006-01-02", searchReq.StartDate)
	if err != nil {
		log.Printf("Invalid start date format: %v", err)
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", searchReq.EndDate)
	if err != nil {
		log.Printf("Invalid end date format: %v", err)
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	dstCities := strings.Split(searchReq.DstCities[0], ",")
	for i, city := range dstCities {
		dstCities[i] = strings.TrimSpace(city)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	resultsChan := make(chan []map[string]interface{}, 1)

	go func() {
		results := getCheapOffersConcurrent(
			ctx,
			flightSession,
			startDate,
			endDate,
			searchReq.TripLength,
			searchReq.SrcCities,
			dstCities,
			language.English,
			searchReq.Airlines,
			getFlightClass(searchReq.TravelClass),
		)
		resultsChan <- results
	}()

	select {
	case results := <-resultsChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	case <-ctx.Done():
		log.Println("Search operation timed out")
		http.Error(w, "Search operation timed out", http.StatusRequestTimeout)
	}
}

func getFlightClass(class string) flights.Class {
	switch class {
	case "Business":
		return flights.Business
	case "First":
		return flights.First
	default:
		return flights.Economy
	}
}

func getCheapOffersConcurrent(
	ctx context.Context,
	session *flights.Session,
	rangeStartDate, rangeEndDate time.Time,
	tripLength int,
	srcCities, dstCities []string,
	lang language.Tag,
	desiredAirlines []string,
	class flights.Class,
) []map[string]interface{} {
	options := flights.Options{
		Travelers: flights.Travelers{Adults: 1},
		Currency:  currency.USD,
		Stops:     flights.AnyStops,
		Class:     class,
		TripType:  flights.RoundTrip,
		Lang:      lang,
	}

	var results []map[string]interface{}
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, srcCity := range srcCities {
		for _, dstCity := range dstCities {
			wg.Add(1)
			go func(src, dst string) {
				defer wg.Done()
				priceGraphOffers, err := session.GetPriceGraph(
					ctx,
					flights.PriceGraphArgs{
						RangeStartDate: rangeStartDate,
						RangeEndDate:   rangeEndDate,
						TripLength:     tripLength,
						SrcCities:      []string{src},
						DstCities:      []string{dst},
						Options:        options,
					},
				)
				if err != nil {
					log.Printf("Error getting price graph for %s to %s: %v", src, dst, err)
					return
				}

				var urlCache string
				var urlCacheOnce sync.Once

				for _, offer := range priceGraphOffers {
					offers, _, err := session.GetOffers(
						ctx,
						flights.Args{
							Date:       offer.StartDate,
							ReturnDate: offer.ReturnDate,
							SrcCities:  []string{src},
							DstCities:  []string{dst},
							Options:    options,
						},
					)
					if err != nil {
						log.Printf("Error getting offers for %s to %s: %v", src, dst, err)
						continue
					}

					for _, fullOffer := range offers {
						formattedOffer := map[string]interface{}{
							"srcCity":    src,
							"dstCity":    dst,
							"startDate":  fullOffer.StartDate.Format("2006-01-02"),
							"returnDate": fullOffer.ReturnDate.Format("2006-01-02"),
							"price":      fullOffer.Price,
							"airline":    fullOffer.Flight[0].AirlineName,
						}

						urlCacheOnce.Do(func() {
							url, err := session.SerializeURL(
								ctx,
								flights.Args{
									Date:       fullOffer.StartDate,
									ReturnDate: fullOffer.ReturnDate,
									SrcCities:  []string{src}, DstCities: []string{dst},
									Options: options,
								},
							)
							if err != nil {
								log.Printf("Error serializing URL: %v", err)
							} else {
								urlCache = url
							}
						})

						if urlCache != "" {
							formattedOffer["url"] = urlCache
						}

						mutex.Lock()
						results = append(results, formattedOffer)
						mutex.Unlock()
					}
				}
			}(srcCity, dstCity)
		}
	}

	wg.Wait()
	return results
}
