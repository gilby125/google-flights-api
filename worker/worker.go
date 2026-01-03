package worker

import (
	"bytes"
	"context"

	// "crypto/tls" // Unused
	// "database/sql" // Unused
	// "encoding/json" // Unused
	"encoding/pem"
	"fmt"

	// "log" // Unused
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"strings"

	"github.com/cloudflare/cloudflare-go"

	// "net/http" // Unused
	// "net/url" // Unused
	"time"

	// "github.com/gilby125/google-flights-api/config" // Unused
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/google/uuid"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// Helper function to parse stops string into flights.Stops type
func parseWorkerStops(s string) flights.Stops {
	switch strings.ToLower(s) {
	case "nonstop":
		return flights.Nonstop
	case "one_stop":
		return flights.Stop1
	case "two_stops":
		return flights.Stop2
	case "any":
		return flights.AnyStops
	default:
		return flights.AnyStops // Default to AnyStops
	}
}

// Helper function to parse class string into flights.Class type
func parseWorkerClass(s string) flights.Class {
	switch strings.ToLower(s) {
	case "economy":
		return flights.Economy
	case "premium_economy":
		return flights.PremiumEconomy
	case "business":
		return flights.Business
	case "first":
		return flights.First
	default:
		return flights.Economy // Default to Economy
	}
}

// CertificateJob represents a TLS certificate management task
type CertificateJob struct {
	Domain           string
	ChallengeType    string
	CloudflareToken  string
	CloudflareZoneID string
	ForceRenewal     bool
}

// CertWorker handles certificate issuance/renewal with Cloudflare DNS challenge
type CertWorker struct {
	ID         int
	JobChannel chan CertificateJob
	WorkerPool chan chan CertificateJob
	quit       chan bool
	postgresDB db.PostgresDB // Changed type from pointer to interface
	cfClient   *cloudflare.API
}

// Main Worker struct remains for flight search operations
type Worker struct {
	postgresDB db.PostgresDB // Changed type from pointer to interface
	neo4jDB    db.Neo4jDatabase
}

// NewWorker creates a new worker instance
func NewWorker(postgresDB db.PostgresDB, neo4jDB db.Neo4jDatabase) *Worker {
	return &Worker{
		postgresDB: postgresDB,
		neo4jDB:    neo4jDB,
	}
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
	Class         string // Changed to string for JSON unmarshal
	Stops         string // Changed to string for JSON unmarshal
	Currency      string
}

type BulkSearchPayload struct {
	Origin            string
	Destination       string
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
	Class             string // Changed to string for JSON unmarshal
	Stops             string // Changed to string for JSON unmarshal
	Currency          string
	BulkSearchID      int `json:"bulk_search_id,omitempty"`
	JobID             int `json:"job_id,omitempty"`
}

// PriceGraphSweepPayload defines the data needed to execute a price graph sweep
type PriceGraphSweepPayload struct {
	SweepID           int       `json:"sweep_id,omitempty"`
	JobID             int       `json:"job_id,omitempty"`
	Origins           []string  `json:"origins"`
	Destinations      []string  `json:"destinations"`
	DepartureDateFrom time.Time `json:"departure_date_from"`
	DepartureDateTo   time.Time `json:"departure_date_to"`
	TripLengths       []int     `json:"trip_lengths,omitempty"`
	TripType          string    `json:"trip_type"`
	Class             string    `json:"class"`
	Stops             string    `json:"stops"`
	Adults            int       `json:"adults"`
	Children          int       `json:"children"`
	InfantsLap        int       `json:"infants_lap"`
	InfantsSeat       int       `json:"infants_seat"`
	Currency          string    `json:"currency"`
	RateLimitMillis   int       `json:"rate_limit_millis,omitempty"`
}

// StoreFlightOffers stores flight offers in the database (Exported for testing)
func (w *Worker) StoreFlightOffers(ctx context.Context, payload FlightSearchPayload, offers []flights.FullOffer, priceRange *flights.PriceRange) error {
	// Begin a transaction using the interface method
	tx, err := w.postgresDB.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the search query
	var queryID int
	searchID := uuid.New().String()
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO search_queries
		(origin, destination, departure_date, return_date, adults, children, infants_lap, infants_seat, trip_type, class, stops, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) -- Removed search_id ($1)
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

	// Insert corresponding search_results record with UUID search_id
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO search_results
		(search_id, search_query_id, origin, destination, departure_date, return_date, adults, children, infants_lap, infants_seat, trip_type, class, stops, currency, search_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())`,
		searchID,
		queryID,
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
		payload.Currency,
	)

	if err != nil {
		return fmt.Errorf("failed to insert search results: %w", err)
	}

	// Store each offer
	for _, offer := range offers {
		// Insert the flight offer
		var offerID int
		totalDuration := int(offer.FlightDuration.Minutes())
		// The searchID is generated above (line 142) and used directly below.
		// No need to fetch it back from search_queries.
		// Extract airline codes from the offer (using the first flight segment for now)
		airlineCodes := []string{}
		if len(offer.Flight) > 0 && len(offer.Flight[0].FlightNumber) >= 2 {
			airlineCodes = append(airlineCodes, offer.Flight[0].FlightNumber[:2])
		}

		// Convert airlineCodes slice to PostgreSQL array string format
		// If empty, use empty array
		var airlineCodesStr string
		if len(airlineCodes) > 0 {
			airlineCodesStr = "{" + strings.Join(airlineCodes, ",") + "}"
		} else {
			airlineCodesStr = "{}"
		}

		err = tx.QueryRowContext(
			ctx,
			`INSERT INTO flight_offers
			(search_id, price, currency, airline_codes, outbound_duration, outbound_stops, return_duration, return_stops)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`,
			searchID,
			offer.Price,
			payload.Currency,
			airlineCodesStr,
			totalDuration, // outbound_duration (in minutes)
			0,             // outbound_stops (placeholder - needs proper calculation)
			totalDuration, // return_duration (placeholder - same as outbound for now)
			0,             // return_stops (placeholder)
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
		if err := w.StoreFlightInNeo4j(ctx, offer); err != nil { // Call exported method
			return fmt.Errorf("failed to store flight in Neo4j: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// StoreFlightInNeo4j stores flight data in Neo4j for graph analysis (Exported for testing)
func (w *Worker) StoreFlightInNeo4j(ctx context.Context, offer flights.FullOffer) error {
	if w.neo4jDB == nil {
		return nil
	}

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

// Start begins listening for certificate jobs
func (cw *CertWorker) Start() {
	go func() {
		for {
			// Register worker to the pool
			cw.WorkerPool <- cw.JobChannel

			select {
			case job := <-cw.JobChannel:
				// Process certificate job
				if err := cw.ProcessJob(context.Background(), job); err != nil {
					fmt.Printf("CertWorker %d failed job: %v\n", cw.ID, err)
				}
			case <-cw.quit:
				return
			}
		}
	}()
}

// Stop signals the worker to stop processing jobs
func (cw *CertWorker) Stop() {
	go func() {
		cw.quit <- true
	}()
}

// ProcessJob handles certificate issuance/renewal workflow
func (cw *CertWorker) ProcessJob(ctx context.Context, job CertificateJob) error {
	// Validate Cloudflare credentials
	if _, err := cw.cfClient.UserDetails(ctx); err != nil {
		return fmt.Errorf("cloudflare auth failed: %w", err)
	}

	// Check existing certificates
	existingCert, err := cw.postgresDB.GetCertificate(job.Domain)
	if err == nil && !job.ForceRenewal && existingCert.Expires.After(time.Now().Add(72*time.Hour)) {
		return nil // Certificate still valid
	}

	// Create new certificate
	cert, privKey, err := generateCertificate(cw.cfClient, job.Domain)
	if err != nil {
		return fmt.Errorf("certificate generation failed: %w", err)
	}

	// Store in database
	if err := cw.postgresDB.StoreCertificate(job.Domain, cert, privKey, time.Now().Add(90*24*time.Hour)); err != nil {
		return fmt.Errorf("certificate generation failed: %w", err)
	}

	return nil
}

// generateCertificate creates a new TLS certificate using Let's Encrypt
func generateCertificate(cfClient *cloudflare.API, domain string) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: domain},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(90 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	certBuf := &bytes.Buffer{}
	pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privKeyBuf := &bytes.Buffer{}
	pem.Encode(privKeyBuf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certBuf.Bytes(), privKeyBuf.Bytes(), nil
}

// processPriceGraphSearch uses the price graph API to find the best dates for a trip
func (w *Worker) processPriceGraphSearch(ctx context.Context, session *flights.Session, origin, destination string, payload BulkSearchPayload) error {
	offers, _, err := session.GetPriceGraph(
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
				Currency: currency.USD,                    // Assuming USD for now, might need adjustment if payload.Currency is different
				Stops:    parseWorkerStops(payload.Stops), // Use helper function
				Class:    parseWorkerClass(payload.Class), // Use helper function
				TripType: flights.RoundTrip,               //Hardcoded for now
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
	if err := w.StoreFlightOffers(ctx, FlightSearchPayload{ // Use exported method
		Origin:        origin,
		Destination:   destination,
		DepartureDate: payload.DepartureDateFrom,
		ReturnDate:    payload.ReturnDateTo,
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
