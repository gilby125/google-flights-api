package api_test

import (
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanDirectSearchDates_PriceGraphOnlyDefaultsRangeAndTripLength(t *testing.T) {
	now := time.Date(2026, 1, 28, 15, 4, 0, 0, time.UTC)
	req := api.DirectSearchRequest{
		TripType:          "round_trip",
		IncludePriceGraph: true,
	}

	plan, err := api.PlanDirectSearchDates(now, req)
	require.NoError(t, err)

	assert.True(t, plan.PriceGraphOnly)
	assert.Equal(t, time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC), plan.DepartureDate)
	assert.Equal(t, time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC), plan.ReturnDate)
	assert.Equal(t, "2026-01-28", plan.PriceGraph.DepartureDateFrom)
	assert.Equal(t, "2026-07-08", plan.PriceGraph.DepartureDateTo) // +161 days
	assert.Equal(t, 7, plan.PriceGraph.TripLengthDays)
}

func TestPlanDirectSearchDates_ErrorsWhenNoDepartureAndNoPriceGraph(t *testing.T) {
	now := time.Date(2026, 1, 28, 15, 4, 0, 0, time.UTC)
	req := api.DirectSearchRequest{IncludePriceGraph: false}

	_, err := api.PlanDirectSearchDates(now, req)
	require.Error(t, err)
}

func TestPlanDirectSearchDates_ParsesFixedDates(t *testing.T) {
	now := time.Date(2026, 1, 28, 15, 4, 0, 0, time.UTC)
	req := api.DirectSearchRequest{
		DepartureDate:     "2026-03-01",
		ReturnDate:        "2026-03-08",
		TripType:          "round_trip",
		IncludePriceGraph: true,
	}

	plan, err := api.PlanDirectSearchDates(now, req)
	require.NoError(t, err)

	assert.False(t, plan.PriceGraphOnly)
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), plan.DepartureDate)
	assert.Equal(t, time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC), plan.ReturnDate)
	assert.True(t, plan.PriceGraph.Include)
}
