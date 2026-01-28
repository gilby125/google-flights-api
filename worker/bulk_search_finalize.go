package worker

import (
	"database/sql"
	"math"
)

func cheapFirstFinalizeStatus(withResults int, errorCount int) string {
	if errorCount > 0 && withResults == 0 {
		return "failed"
	}
	if errorCount > 0 && withResults > 0 {
		return "completed_with_errors"
	}
	return "completed"
}

func cheapFirstFinalizePriceStats(withResults int, minPrice, maxPrice, sumPrice float64) (sql.NullFloat64, sql.NullFloat64, sql.NullFloat64) {
	var minPriceNull, maxPriceNull, avgPriceNull sql.NullFloat64
	if withResults <= 0 {
		return minPriceNull, maxPriceNull, avgPriceNull
	}
	if minPrice >= math.MaxFloat64 {
		return minPriceNull, maxPriceNull, avgPriceNull
	}
	if !isDBSafePrice(minPrice) || !isDBSafePrice(maxPrice) {
		return minPriceNull, maxPriceNull, avgPriceNull
	}
	minPriceNull = sql.NullFloat64{Float64: minPrice, Valid: true}
	maxPriceNull = sql.NullFloat64{Float64: maxPrice, Valid: true}
	avgPriceNull = sql.NullFloat64{Float64: sumPrice / float64(withResults), Valid: true}
	return minPriceNull, maxPriceNull, avgPriceNull
}
