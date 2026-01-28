package worker

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDBSafePrice(t *testing.T) {
	assert.False(t, isDBSafePrice(0))
	assert.False(t, isDBSafePrice(-1))
	assert.False(t, isDBSafePrice(math.NaN()))
	assert.False(t, isDBSafePrice(math.Inf(1)))
	assert.False(t, isDBSafePrice(math.Inf(-1)))

	assert.True(t, isDBSafePrice(1))
	assert.True(t, isDBSafePrice(maxDBPrice-0.01))
	assert.False(t, isDBSafePrice(maxDBPrice))
	assert.False(t, isDBSafePrice(maxDBPrice+1))
}

func TestCheapFirstFinalizeStatus(t *testing.T) {
	assert.Equal(t, "completed", cheapFirstFinalizeStatus(0, 0))
	assert.Equal(t, "failed", cheapFirstFinalizeStatus(0, 1))
	assert.Equal(t, "completed", cheapFirstFinalizeStatus(2, 0))
	assert.Equal(t, "completed_with_errors", cheapFirstFinalizeStatus(2, 1))
}

func TestCheapFirstFinalizePriceStats_NoResults(t *testing.T) {
	minN, maxN, avgN := cheapFirstFinalizePriceStats(0, math.MaxFloat64, 0, 0)
	assert.False(t, minN.Valid)
	assert.False(t, maxN.Valid)
	assert.False(t, avgN.Valid)
}

func TestCheapFirstFinalizePriceStats_WithResults(t *testing.T) {
	minN, maxN, avgN := cheapFirstFinalizePriceStats(2, 100, 200, 300)
	assert.True(t, minN.Valid)
	assert.True(t, maxN.Valid)
	assert.True(t, avgN.Valid)
	assert.InDelta(t, 100.0, minN.Float64, 0.0001)
	assert.InDelta(t, 200.0, maxN.Float64, 0.0001)
	assert.InDelta(t, 150.0, avgN.Float64, 0.0001)
}

func TestCheapFirstFinalizePriceStats_SkipsInvalidPrice(t *testing.T) {
	minN, maxN, avgN := cheapFirstFinalizePriceStats(1, math.MaxFloat64, 200, 200)
	assert.False(t, minN.Valid)
	assert.False(t, maxN.Valid)
	assert.False(t, avgN.Valid)

	minN, maxN, avgN = cheapFirstFinalizePriceStats(1, maxDBPrice, maxDBPrice, maxDBPrice)
	assert.False(t, minN.Valid)
	assert.False(t, maxN.Valid)
	assert.False(t, avgN.Valid)
}
