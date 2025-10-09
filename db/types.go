package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gilby125/google-flights-api/flights"
)

// RowScanner defines the interface for scanning a single row result.
// This allows mocking database row scanning behavior.
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// Search represents a flight search query and results (kept for potential future use, though might be redundant)
type Search struct {
	ID          string
	SearchID    string
	Origin      string
	Destination string
	Departure   time.Time
	Return      *time.Time
	Passengers  int
	CabinClass  string
	Status      string
	Results     []flights.FullOffer
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Certificate represents a TLS certificate stored in the database
type Certificate struct {
	Domain     string
	CertChain  []byte
	PrivateKey []byte
	Expires    time.Time
}

// --- Struct Definitions for Query Results ---

// SearchQuery represents the data structure for a single search query row
type SearchQuery struct {
	ID            int
	Origin        string
	Destination   string
	DepartureDate time.Time
	ReturnDate    sql.NullTime
	Status        string
	CreatedAt     time.Time
}

// FlightOffer represents the data structure for a single flight offer row
type FlightOffer struct {
	ID               int
	Price            float64
	Currency         string
	AirlineCodes     sql.NullString
	OutboundDuration sql.NullInt64
	OutboundStops    sql.NullInt64
	ReturnDuration   sql.NullInt64
	ReturnStops      sql.NullInt64
	CreatedAt        time.Time
}

// FlightSegment represents the data structure for a single flight segment row
type FlightSegment struct {
	ID               int
	FlightOfferID    int // Foreign key
	AirlineCode      string
	FlightNumber     string
	DepartureAirport string
	ArrivalAirport   string
	DepartureTime    time.Time
	ArrivalTime      time.Time
	Duration         int
	Airplane         string
	Legroom          string
	IsReturn         bool
}

// ScheduledJob represents the data structure for a scheduled job row
type ScheduledJob struct {
	ID             int
	Name           string
	CronExpression string
	Enabled        bool
	LastRun        sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// JobDetails represents the data structure for job details row
type JobDetails struct {
	JobID              int // Foreign key or primary key if separate table
	Origin             string
	Destination        string
	DepartureDateStart time.Time
	DepartureDateEnd   time.Time
	ReturnDateStart    sql.NullTime
	ReturnDateEnd      sql.NullTime
	TripLength         sql.NullInt32
	DynamicDates       bool          // Use dates relative to execution time
	DaysFromExecution  sql.NullInt32 // Start searching X days from now
	SearchWindowDays   sql.NullInt32 // Search within X days window
	Adults             int
	Children           int
	InfantsLap         int
	InfantsSeat        int
	TripType           string
	Class              string
	Stops              string
	Currency           string // Added Currency
}

// BulkSearch represents the metadata for a bulk search
type BulkSearch struct {
	ID            int
	JobID         sql.NullInt32
	Status        string
	TotalSearches int
	Completed     int
	TotalOffers   int
	ErrorCount    int
	Currency      string
	CreatedAt     time.Time
	CompletedAt   sql.NullTime
	UpdatedAt     time.Time
	MinPrice      sql.NullFloat64
	MaxPrice      sql.NullFloat64
	AveragePrice  sql.NullFloat64
}

// BulkSearchResult represents a single result row from a bulk search
type BulkSearchResult struct {
	Origin               string
	Destination          string
	DepartureDate        time.Time
	ReturnDate           sql.NullTime
	Price                float64
	Currency             string
	AirlineCode          sql.NullString
	Duration             sql.NullInt32
	SrcAirportCode       sql.NullString
	DstAirportCode       sql.NullString
	SrcCity              sql.NullString
	DstCity              sql.NullString
	FlightDuration       sql.NullInt32
	ReturnFlightDuration sql.NullInt32
	OutboundFlights      json.RawMessage
	ReturnFlights        json.RawMessage
	OfferJSON            json.RawMessage
}

// BulkSearchSummary represents aggregated results for a bulk search run
type BulkSearchSummary struct {
	ID           int
	Status       string
	Completed    int
	TotalOffers  int
	ErrorCount   int
	MinPrice     sql.NullFloat64
	MaxPrice     sql.NullFloat64
	AveragePrice sql.NullFloat64
}

// BulkSearchResultRecord represents the data needed to insert a bulk search result
type BulkSearchResultRecord struct {
	BulkSearchID         int
	Origin               string
	Destination          string
	DepartureDate        time.Time
	ReturnDate           sql.NullTime
	Price                float64
	Currency             string
	AirlineCode          sql.NullString
	Duration             sql.NullInt32
	SrcAirportCode       sql.NullString
	DstAirportCode       sql.NullString
	SrcCity              sql.NullString
	DstCity              sql.NullString
	FlightDuration       sql.NullInt32
	ReturnFlightDuration sql.NullInt32
	OutboundFlightsJSON  []byte
	ReturnFlightsJSON    []byte
	OfferJSON            []byte
}

// BulkSearchOffer represents all offers captured during a bulk run
type BulkSearchOffer struct {
	ID                   int
	BulkSearchID         int
	Origin               string
	Destination          string
	DepartureDate        time.Time
	ReturnDate           sql.NullTime
	Price                float64
	Currency             string
	AirlineCodes         []string
	SrcAirportCode       sql.NullString
	DstAirportCode       sql.NullString
	SrcCity              sql.NullString
	DstCity              sql.NullString
	FlightDuration       sql.NullInt32
	ReturnFlightDuration sql.NullInt32
	OutboundFlights      json.RawMessage
	ReturnFlights        json.RawMessage
	OfferJSON            json.RawMessage
	CreatedAt            time.Time
}

// BulkSearchOfferRecord represents the data needed to insert an offer for a bulk run
type BulkSearchOfferRecord struct {
	BulkSearchID         int
	Origin               string
	Destination          string
	DepartureDate        time.Time
	ReturnDate           sql.NullTime
	Price                float64
	Currency             string
	AirlineCodes         []string
	SrcAirportCode       sql.NullString
	DstAirportCode       sql.NullString
	SrcCity              sql.NullString
	DstCity              sql.NullString
	FlightDuration       sql.NullInt32
	ReturnFlightDuration sql.NullInt32
	OutboundFlightsJSON  []byte
	ReturnFlightsJSON    []byte
	OfferJSON            []byte
}

// Airport represents an airport row
type Airport struct {
	Code      string
	Name      string
	City      string
	Country   string
	Latitude  sql.NullFloat64
	Longitude sql.NullFloat64
}

// Airline represents an airline row
type Airline struct {
	Code    string
	Name    string
	Country string
}

// --- End Struct Definitions ---
