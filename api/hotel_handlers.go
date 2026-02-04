package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gilby125/google-flights-api/hotels"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// HotelSearchRequest represents a hotel search request
type HotelSearchRequest struct {
	Location     string   `json:"location" binding:"required"`
	CheckInDate  DateOnly `json:"checkin_date" binding:"required"`
	CheckOutDate DateOnly `json:"checkout_date" binding:"required"`
	Adults       int      `json:"adults" binding:"required,min=1"`
	Children     int      `json:"children" binding:"min=0"`
	Currency     string   `json:"currency" binding:"required,len=3"`
}

// DirectHotelSearch returns a handler for direct hotel search (immediate results)
func DirectHotelSearch(session *hotels.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		if session == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Hotel search service unavailable"})
			return
		}

		var req HotelSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert request to hotels.Args
		curr, err := currency.ParseISO(strings.ToUpper(req.Currency))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency code"})
			return
		}

		args := hotels.Args{
			Location:     req.Location,
			CheckInDate:  req.CheckInDate.Time,
			CheckOutDate: req.CheckOutDate.Time,
			Travelers: hotels.Travelers{
				Adults:   req.Adults,
				Children: req.Children,
			},
			Currency: curr,
			Lang:     language.English, // Default to English for now
		}

		if err := args.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("validation failed: %v", err)})
			return
		}

		// Perform the search
		offers, err := session.GetOffers(c.Request.Context(), args)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"location":      req.Location,
			"checkin_date":  req.CheckInDate,
			"checkout_date": req.CheckOutDate,
			"offers":        offers,
			"count":         len(offers),
		})
	}
}
