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
	JobType        string // "bulk_search" or "price_graph_sweep"
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

// PriceGraphSweep captures metadata about a price graph sweep run
type PriceGraphSweep struct {
	ID               int
	JobID            sql.NullInt32
	Status           string
	OriginCount      int
	DestinationCount int
	TripLengthMin    sql.NullInt32
	TripLengthMax    sql.NullInt32
	Currency         string
	ErrorCount       int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	StartedAt        sql.NullTime
	CompletedAt      sql.NullTime
}

// PriceGraphResult represents the cheapest fare returned by the price graph API
type PriceGraphResult struct {
	ID            int
	SweepID       int
	Origin        string
	Destination   string
	DepartureDate time.Time
	ReturnDate    sql.NullTime
	TripLength    sql.NullInt32
	Price         float64
	Currency      string
	DistanceMiles sql.NullFloat64
	CostPerMile   sql.NullFloat64
	Adults        int
	Children      int
	InfantsLap    int
	InfantsSeat   int
	TripType      string
	Class         string
	Stops         string
	SearchURL     sql.NullString
	QueriedAt     time.Time
	CreatedAt     time.Time
}

// PriceGraphResultRecord is used when inserting price graph results
type PriceGraphResultRecord struct {
	SweepID       int
	Origin        string
	Destination   string
	DepartureDate time.Time
	ReturnDate    sql.NullTime
	TripLength    sql.NullInt32
	Price         float64
	Currency      string
	DistanceMiles sql.NullFloat64
	CostPerMile   sql.NullFloat64
	Adults        int
	Children      int
	InfantsLap    int
	InfantsSeat   int
	TripType      string
	Class         string
	Stops         string
	SearchURL     sql.NullString
	QueriedAt     time.Time
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
	DistanceMiles        sql.NullFloat64
	CostPerMile          sql.NullFloat64
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
	DistanceMiles        sql.NullFloat64
	CostPerMile          sql.NullFloat64
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

// PriceGraphSweepJobDetails represents sweep-specific job configuration
type PriceGraphSweepJobDetails struct {
	ID                  int
	JobID               int
	TripLengths         []int
	DepartureWindowDays int
	DynamicDates        bool
	TripType            string
	Class               string
	Stops               string
	Adults              int
	Currency            string
	RateLimitMillis     int
	InternationalOnly   bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ContinuousSweepProgress tracks the current state of the continuous sweep
type ContinuousSweepProgress struct {
	ID                  int
	SweepNumber         int
	RouteIndex          int
	TotalRoutes         int
	CurrentOrigin       sql.NullString
	CurrentDestination  sql.NullString
	QueriesCompleted    int
	ErrorsCount         int
	LastError           sql.NullString
	SweepStartedAt      sql.NullTime
	LastUpdated         time.Time
	TripLengths         []int
	PacingMode          string // "adaptive" or "fixed"
	TargetDurationHours int
	MinDelayMs          int
	IsRunning           bool
	IsPaused            bool
	InternationalOnly   bool
}

// ContinuousSweepStats represents historical stats for completed sweeps
type ContinuousSweepStats struct {
	ID                   int
	SweepNumber          int
	StartedAt            time.Time
	CompletedAt          sql.NullTime
	TotalRoutes          int
	SuccessfulQueries    int
	FailedQueries        int
	TotalDurationSeconds sql.NullInt32
	AvgDelayMs           sql.NullInt32
	MinPriceFound        sql.NullFloat64
	MaxPriceFound        sql.NullFloat64
	CreatedAt            time.Time
}

// SweepStatusResponse is the API response for sweep status
type SweepStatusResponse struct {
	IsRunning           bool      `json:"is_running"`
	IsPaused            bool      `json:"is_paused"`
	SweepNumber         int       `json:"sweep_number"`
	RouteIndex          int       `json:"route_index"`
	TotalRoutes         int       `json:"total_routes"`
	ProgressPercent     float64   `json:"progress_percent"`
	CurrentOrigin       string    `json:"current_origin"`
	CurrentDestination  string    `json:"current_destination"`
	Class               string    `json:"class"`
	Stops               string    `json:"stops"`
	TripLengths         []int     `json:"trip_lengths"`
	QueriesCompleted    int       `json:"queries_completed"`
	ErrorsCount         int       `json:"errors_count"`
	LastError           string    `json:"last_error,omitempty"`
	SweepStartedAt      time.Time `json:"sweep_started_at,omitempty"`
	EstimatedCompletion time.Time `json:"estimated_completion,omitempty"`
	PacingMode          string    `json:"pacing_mode"`
	CurrentDelayMs      int       `json:"current_delay_ms"`
	MinDelayMs          int       `json:"min_delay_ms"`
	TargetDurationHours int       `json:"target_duration_hours"`
	QueriesPerHour      float64   `json:"queries_per_hour"`
}

// ContinuousSweepResultsFilter defines filters for querying continuous sweep results
type ContinuousSweepResultsFilter struct {
	Origin      string
	Destination string
	FromDate    time.Time
	ToDate      time.Time
	Limit       int
	Offset      int
}

// --- Deal Detection Types ---

// DealClassification constants
const (
	DealClassGood      = "good"       // 20-35% below baseline
	DealClassGreat     = "great"      // 35-50% below baseline
	DealClassAmazing   = "amazing"    // 50%+ below baseline
	DealClassErrorFare = "error_fare" // Suspiciously cheap (70%+)
)

// DealSourceType constants
const (
	DealSourceSweep   = "sweep"   // From continuous price sweeps
	DealSourceSocial  = "social"  // From social-pulse webhook
	DealSourceWebhook = "webhook" // From external webhook
	DealSourceManual  = "manual"  // Manually entered
)

// DealStatus constants
const (
	DealStatusActive    = "active"
	DealStatusExpired   = "expired"
	DealStatusPublished = "published"
	DealStatusVerified  = "verified"
)

// RouteBaseline stores historical price statistics for a route
type RouteBaseline struct {
	ID          int
	Origin      string
	Destination string
	TripLength  int
	Class       string
	SampleCount int
	MeanPrice   sql.NullFloat64
	MedianPrice sql.NullFloat64
	StddevPrice sql.NullFloat64
	MinPrice    sql.NullFloat64
	MaxPrice    sql.NullFloat64
	P10Price    sql.NullFloat64
	P25Price    sql.NullFloat64
	P75Price    sql.NullFloat64
	P90Price    sql.NullFloat64
	WindowStart sql.NullTime
	WindowEnd   sql.NullTime
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

// DetectedDeal represents a flight deal identified by the system
type DetectedDeal struct {
	ID                 int
	Origin             string
	Destination        string
	DepartureDate      time.Time
	ReturnDate         sql.NullTime
	TripLength         sql.NullInt32
	Price              float64
	Currency           string
	BaselineMean       sql.NullFloat64
	BaselineMedian     sql.NullFloat64
	DiscountPercent    sql.NullFloat64
	DealScore          sql.NullInt32
	DealClassification sql.NullString
	DistanceMiles      sql.NullFloat64
	CostPerMile        sql.NullFloat64
	CabinClass         string
	SourceType         string
	SourceID           sql.NullString
	SearchURL          sql.NullString
	DealFingerprint    string
	FirstSeenAt        time.Time
	LastSeenAt         time.Time
	TimesSeen          int
	Status             string
	Verified           bool
	VerifiedPrice      sql.NullFloat64
	VerifiedAt         sql.NullTime
	ExpiresAt          sql.NullTime
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// DealAlert represents a published deal ready for notification
type DealAlert struct {
	ID                   int
	DetectedDealID       int
	Origin               string
	Destination          string
	Price                float64
	Currency             string
	DiscountPercent      sql.NullFloat64
	DealClassification   sql.NullString
	DealScore            sql.NullInt32
	PublishedAt          time.Time
	PublishMethod        string
	NotificationSent     bool
	NotificationSentAt   sql.NullTime
	NotificationChannels []string
	CreatedAt            time.Time
}

// DealSource represents an external deal source (for webhook integration)
type DealSource struct {
	ID             int
	Name           string
	SourceType     string
	WebhookURL     sql.NullString
	APIKey         sql.NullString
	Enabled        bool
	LastReceivedAt sql.NullTime
	DealsReceived  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// DealFilter is used to filter deals in queries
type DealFilter struct {
	Origin         string
	Destination    string
	MinScore       int
	MaxPrice       *float64
	MinDiscount    *float64
	Classification string
	Status         string
	SourceType     string
	IncludeExpired bool
	Limit          int
	Offset         int
}

// --- End Struct Definitions ---
