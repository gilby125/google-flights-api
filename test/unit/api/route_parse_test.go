package api_test

import (
	"testing"

	"github.com/gilby125/google-flights-api/api"
	"github.com/stretchr/testify/assert"
)

func TestParseRouteInputs_Single(t *testing.T) {
	origins, destinations, err := api.ParseRouteInputs("JFK", "LAX")
	assert.NoError(t, err)
	assert.Equal(t, []string{"JFK"}, origins)
	assert.Equal(t, []string{"LAX"}, destinations)
}

func TestParseRouteInputs_Single_WithLabels(t *testing.T) {
	origins, destinations, err := api.ParseRouteInputs(
		"JFK - John F. Kennedy International Airport, New York",
		"lax - Los Angeles International Airport, Los Angeles",
	)
	assert.NoError(t, err)
	assert.Equal(t, []string{"JFK"}, origins)
	assert.Equal(t, []string{"LAX"}, destinations)
}

func TestParseRouteInputs_MultiCrossProduct(t *testing.T) {
	origins, destinations, err := api.ParseRouteInputs("MKE,MSN", "FLL,MIA")
	assert.NoError(t, err)
	assert.Equal(t, []string{"MKE", "MSN"}, origins)
	assert.Equal(t, []string{"FLL", "MIA"}, destinations)
}

func TestParseRouteInputs_CombinedExpression(t *testing.T) {
	origins, destinations, err := api.ParseRouteInputs("MKE,MSN>FLL,MIA", "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"MKE", "MSN"}, origins)
	assert.Equal(t, []string{"FLL", "MIA"}, destinations)
}

func TestParseRouteInputs_SpaceSeparatedCodes(t *testing.T) {
	origins, destinations, err := api.ParseRouteInputs("mke msn", "fll")
	assert.NoError(t, err)
	assert.Equal(t, []string{"MKE", "MSN"}, origins)
	assert.Equal(t, []string{"FLL"}, destinations)
}

func TestParseRouteInputs_MissingDestination(t *testing.T) {
	_, _, err := api.ParseRouteInputs("JFK", "")
	assert.Error(t, err)
}
