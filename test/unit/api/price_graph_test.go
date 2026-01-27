package api_test

import (
	"testing"
	"time"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/flights"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPriceGraphArgs_DefaultWindowCentersAndDerivesTripLength(t *testing.T) {
	now := time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC)
	departure := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)

	opts := flights.OptionsDefault()
	opts.TripType = flights.RoundTrip

	args, err := api.BuildPriceGraphArgs(
		now,
		"JFK",
		"LHR",
		departure,
		returnDate,
		opts,
		api.PriceGraphBuildParams{Include: true, WindowDays: 30},
	)
	require.NoError(t, err)

	assert.Equal(t, time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC), args.RangeStartDate)
	assert.Equal(t, time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), args.RangeEndDate)
	assert.Equal(t, 7, args.TripLength)
	assert.Equal(t, []string{"JFK"}, args.SrcAirports)
	assert.Equal(t, []string{"LHR"}, args.DstAirports)
}

func TestBuildPriceGraphArgs_ClampsStartToTodayWhenDepartureNearNow(t *testing.T) {
	now := time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC)
	departure := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	returnDate := time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC)

	opts := flights.OptionsDefault()
	opts.TripType = flights.RoundTrip

	args, err := api.BuildPriceGraphArgs(
		now,
		"MKE",
		"FLL",
		departure,
		returnDate,
		opts,
		api.PriceGraphBuildParams{Include: true, WindowDays: 30},
	)
	require.NoError(t, err)

	assert.Equal(t, time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC), args.RangeStartDate)
	assert.Equal(t, time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC), args.RangeEndDate)
	assert.Equal(t, 7, args.TripLength)
}

func TestBuildPriceGraphArgs_UsesExplicitRangeAndTripLength(t *testing.T) {
	now := time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC)
	departure := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	returnDate := time.Time{}

	opts := flights.OptionsDefault()
	opts.TripType = flights.RoundTrip

	args, err := api.BuildPriceGraphArgs(
		now,
		"SFO",
		"JFK",
		departure,
		returnDate,
		opts,
		api.PriceGraphBuildParams{
			Include:           true,
			DepartureDateFrom: "2026-01-10",
			DepartureDateTo:   "2026-03-01",
			TripLengthDays:    5,
		},
	)
	require.NoError(t, err)

	assert.Equal(t, time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC), args.RangeStartDate)
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), args.RangeEndDate)
	assert.Equal(t, 5, args.TripLength)
}

func TestBuildPriceGraphArgs_RequiresBothRangeDatesWhenSpecified(t *testing.T) {
	now := time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC)
	departure := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)

	opts := flights.OptionsDefault()
	opts.TripType = flights.RoundTrip

	_, err := api.BuildPriceGraphArgs(
		now,
		"SFO",
		"JFK",
		departure,
		time.Time{},
		opts,
		api.PriceGraphBuildParams{
			Include:           true,
			DepartureDateFrom: "2026-02-01",
		},
	)
	require.Error(t, err)
}
