package db

import (
	"database/sql"
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
	ID           string
	SearchID     string
	Origin       string
	Destination  string
	Departure    time.Time
	Return       *time.Time
	Passengers   int
	CabinClass   string
	Status       string
	Results      []flights.FullOffer
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Certificate represents a TLS certificate stored in the database
type Certificate struct {
	Domain      string
	CertChain   []byte
	PrivateKey  []byte
	Expires     time.Time
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
	ID            int
	SearchQueryID int // Foreign key
	Price         float64
	Currency      string
	DepartureDate time.Time
	ReturnDate    sql.NullTime
	TotalDuration int
	CreatedAt     time.Time
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
	Status        string
	TotalSearches int
	Completed     int
	CreatedAt     time.Time
	CompletedAt   sql.NullTime
	MinPrice      sql.NullFloat64
	MaxPrice      sql.NullFloat64
	AveragePrice  sql.NullFloat64
}

// BulkSearchResult represents a single result row from a bulk search
type BulkSearchResult struct {
	Origin        string
	Destination   string
	DepartureDate time.Time
	ReturnDate    sql.NullTime
	Price         float64
	Currency      string
	AirlineCode   string
	Duration      int
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
