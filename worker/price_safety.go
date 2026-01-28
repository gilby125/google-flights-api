package worker

import "math"

const maxDBPrice = 9_999_999_999.99 // DECIMAL(12,2) must round to an absolute value < 10^10.

func isDBSafePrice(price float64) bool {
	if price <= 0 {
		return false
	}
	if math.IsNaN(price) || math.IsInf(price, 0) {
		return false
	}
	return price < maxDBPrice
}
