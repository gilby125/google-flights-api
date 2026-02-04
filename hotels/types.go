package hotels

import (
	"time"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// Hotel represents a single hotel search result.
type Hotel struct {
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Rating      float64  `json:"rating"`
	Stars       int      `json:"stars,omitempty"`
	Address     string   `json:"address,omitempty"`
	Description string   `json:"description,omitempty"`
	Images      []string `json:"images,omitempty"`
	HotelID     string   `json:"hotel_id,omitempty"`
	Latitude    float64  `json:"latitude,omitempty"`
	Longitude   float64  `json:"longitude,omitempty"`
}

// Args defines the arguments for a hotel search.
type Args struct {
	Location     string // City or region name
	CheckInDate  time.Time
	CheckOutDate time.Time
	Travelers    Travelers
	Currency     currency.Unit
	Lang         language.Tag
}

// Travelers holds the count of adults and children.
type Travelers struct {
	Adults   int
	Children int
}

// Validate checks if the arguments are valid.
func (a Args) Validate() error {
	if a.Location == "" {
		return &ValidationError{Field: "Location", Message: "cannot be empty"}
	}
	if a.CheckInDate.IsZero() {
		return &ValidationError{Field: "CheckInDate", Message: "cannot be zero"}
	}
	if a.CheckOutDate.IsZero() {
		return &ValidationError{Field: "CheckOutDate", Message: "cannot be zero"}
	}
	if !a.CheckOutDate.After(a.CheckInDate) {
		return &ValidationError{Field: "CheckOutDate", Message: "must be after CheckInDate"}
	}
	if a.Travelers.Adults < 1 {
		return &ValidationError{Field: "Travelers.Adults", Message: "must be at least 1"}
	}
	if a.Travelers.Children < 0 {
		return &ValidationError{Field: "Travelers.Children", Message: "must be at least 0"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}
