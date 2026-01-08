package deals

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/gilby125/google-flights-api/config"
	"github.com/gilby125/google-flights-api/db"
)

// DealDetector identifies flight deals from price data
type DealDetector struct {
	db     db.PostgresDB
	config config.DealConfig
}

// NewDealDetector creates a new deal detector instance
func NewDealDetector(database db.PostgresDB, cfg config.DealConfig) *DealDetector {
	return &DealDetector{
		db:     database,
		config: cfg,
	}
}

// DefaultDealConfig returns default deal detection configuration
func DefaultDealConfig() config.DealConfig {
	return config.DealConfig{
		GoodDealThreshold:    0.20,
		GreatDealThreshold:   0.35,
		AmazingDealThreshold: 0.50,
		ErrorFareThreshold:   0.70,
		CostPerMileEconomy:   0.05,
		CostPerMileBusiness:  0.08,
		CostPerMileFirst:     0.12,
		BaselineWindowDays:   90,
		BaselineMinSamples:   10,
		DealTTLHours:         24,
		AutoPublish:          true,
	}
}

// DetectDeal checks if a price result qualifies as a deal
func (d *DealDetector) DetectDeal(ctx context.Context, result db.PriceGraphResultRecord) (*db.DetectedDeal, error) {
	// Get baseline for this route
	baseline, err := d.getBaseline(ctx, result.Origin, result.Destination,
		int(result.TripLength.Int32), result.Class)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline: %w", err)
	}

	// Skip if insufficient baseline data
	if baseline == nil || baseline.SampleCount < d.config.BaselineMinSamples {
		return nil, nil
	}

	// Calculate discount percentage
	baselinePrice := baseline.MedianPrice.Float64
	if baselinePrice <= 0 {
		baselinePrice = baseline.MeanPrice.Float64
	}
	if baselinePrice <= 0 {
		return nil, nil
	}

	discountPercent := (baselinePrice - result.Price) / baselinePrice

	// Check if this qualifies as a deal
	if discountPercent < d.config.GoodDealThreshold && !d.isCostPerMileDeal(result) {
		return nil, nil // Not a deal
	}

	// Classify the deal
	classification := d.classifyDeal(discountPercent)

	// Calculate composite score (0-100)
	score := d.calculateScore(discountPercent, result.CostPerMile.Float64, result.Class)

	// Generate fingerprint for deduplication
	fingerprint := d.generateFingerprint(result)

	// Set expiration
	expiresAt := time.Now().Add(time.Duration(d.config.DealTTLHours) * time.Hour)

	deal := &db.DetectedDeal{
		Origin:             result.Origin,
		Destination:        result.Destination,
		DepartureDate:      result.DepartureDate,
		ReturnDate:         result.ReturnDate,
		TripLength:         result.TripLength,
		Price:              result.Price,
		Currency:           result.Currency,
		BaselineMean:       baseline.MeanPrice,
		BaselineMedian:     baseline.MedianPrice,
		DiscountPercent:    sql.NullFloat64{Float64: discountPercent * 100, Valid: true},
		DealScore:          sql.NullInt32{Int32: int32(score), Valid: true},
		DealClassification: sql.NullString{String: classification, Valid: true},
		DistanceMiles:      result.DistanceMiles,
		CostPerMile:        result.CostPerMile,
		CabinClass:         result.Class,
		SourceType:         db.DealSourceSweep,
		SearchURL:          result.SearchURL,
		DealFingerprint:    fingerprint,
		FirstSeenAt:        time.Now(),
		LastSeenAt:         time.Now(),
		TimesSeen:          1,
		Status:             db.DealStatusActive,
		ExpiresAt:          sql.NullTime{Time: expiresAt, Valid: true},
	}

	return deal, nil
}

// classifyDeal determines the deal classification based on discount percentage
func (d *DealDetector) classifyDeal(discountPercent float64) string {
	switch {
	case discountPercent >= d.config.ErrorFareThreshold:
		return db.DealClassErrorFare
	case discountPercent >= d.config.AmazingDealThreshold:
		return db.DealClassAmazing
	case discountPercent >= d.config.GreatDealThreshold:
		return db.DealClassGreat
	case discountPercent >= d.config.GoodDealThreshold:
		return db.DealClassGood
	default:
		return ""
	}
}

// calculateScore computes a composite 0-100 score for the deal
func (d *DealDetector) calculateScore(discountPercent, costPerMile float64, cabinClass string) int {
	score := 0.0

	// Discount component (0-60 points)
	// 20% = 20 points, 50% = 50 points, cap at 60
	discountPoints := math.Min(discountPercent*100, 60)
	score += discountPoints

	// Cost-per-mile component (0-40 points)
	threshold := d.getCostPerMileThreshold(cabinClass)
	if costPerMile > 0 && threshold > 0 {
		// Lower CPM = more points
		if costPerMile <= threshold*0.5 {
			score += 40 // Exceptional CPM
		} else if costPerMile <= threshold*0.75 {
			score += 30
		} else if costPerMile <= threshold {
			score += 20
		} else if costPerMile <= threshold*1.5 {
			score += 10
		}
	}

	return int(math.Min(score, 100))
}

// isCostPerMileDeal checks if the deal qualifies based on cost-per-mile alone
func (d *DealDetector) isCostPerMileDeal(result db.PriceGraphResultRecord) bool {
	if !result.CostPerMile.Valid || result.CostPerMile.Float64 <= 0 {
		return false
	}

	threshold := d.getCostPerMileThreshold(result.Class)
	return result.CostPerMile.Float64 <= threshold
}

// getCostPerMileThreshold returns the CPM threshold for a cabin class
func (d *DealDetector) getCostPerMileThreshold(cabinClass string) float64 {
	switch cabinClass {
	case "business":
		return d.config.CostPerMileBusiness
	case "first":
		return d.config.CostPerMileFirst
	case "premium_economy":
		return d.config.CostPerMileEconomy * 1.5
	default:
		return d.config.CostPerMileEconomy
	}
}

// generateFingerprint creates a unique identifier for deduplication
func (d *DealDetector) generateFingerprint(result db.PriceGraphResultRecord) string {
	// Fingerprint based on: route + price bucket + date range
	priceBucket := int(result.Price/50) * 50 // Round to nearest $50

	data := fmt.Sprintf("%s-%s-%d-%s-%s",
		result.Origin,
		result.Destination,
		priceBucket,
		result.DepartureDate.Format("2006-01"),
		result.Class,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getBaseline retrieves the price baseline for a route
func (d *DealDetector) getBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*db.RouteBaseline, error) {
	// First try to get exact match
	baseline, err := d.db.GetRouteBaseline(ctx, origin, dest, tripLength, class)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if baseline != nil {
		return baseline, nil
	}

	// Try to calculate baseline from recent prices
	return d.calculateBaseline(ctx, origin, dest, tripLength, class)
}

// calculateBaseline computes baseline statistics from price history
func (d *DealDetector) calculateBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*db.RouteBaseline, error) {
	prices, err := d.db.GetPriceHistoryForRoute(ctx, origin, dest, tripLength, class, d.config.BaselineWindowDays)
	if err != nil {
		return nil, err
	}

	if len(prices) < d.config.BaselineMinSamples {
		return nil, nil
	}

	// Calculate statistics
	sort.Float64s(prices)

	baseline := &db.RouteBaseline{
		Origin:      origin,
		Destination: dest,
		TripLength:  tripLength,
		Class:       class,
		SampleCount: len(prices),
		MeanPrice:   sql.NullFloat64{Float64: mean(prices), Valid: true},
		MedianPrice: sql.NullFloat64{Float64: median(prices), Valid: true},
		StddevPrice: sql.NullFloat64{Float64: stddev(prices), Valid: true},
		MinPrice:    sql.NullFloat64{Float64: prices[0], Valid: true},
		MaxPrice:    sql.NullFloat64{Float64: prices[len(prices)-1], Valid: true},
		P10Price:    sql.NullFloat64{Float64: percentile(prices, 10), Valid: true},
		P25Price:    sql.NullFloat64{Float64: percentile(prices, 25), Valid: true},
		P75Price:    sql.NullFloat64{Float64: percentile(prices, 75), Valid: true},
		P90Price:    sql.NullFloat64{Float64: percentile(prices, 90), Valid: true},
	}

	// Upsert the baseline
	if err := d.db.UpsertRouteBaseline(ctx, *baseline); err != nil {
		log.Printf("Failed to upsert baseline for %s-%s: %v", origin, dest, err)
	}

	return baseline, nil
}

// Helper functions for statistics

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func median(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2
	}
	return values[n/2]
}

func percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}
	idx := int(float64(len(values)-1) * float64(p) / 100)
	return values[idx]
}

func stddev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	sum := 0.0
	for _, v := range values {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(values)-1))
}
