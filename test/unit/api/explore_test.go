package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gilby125/google-flights-api/api"
	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type staticNeo4jResult struct {
	records []*neo4j.Record
	idx     int
	err     error
}

func (r *staticNeo4jResult) Next() bool {
	if r.err != nil {
		return false
	}
	if r.idx >= len(r.records) {
		return false
	}
	r.idx++
	return true
}

func (r *staticNeo4jResult) Record() *neo4j.Record {
	if r.idx == 0 || r.idx > len(r.records) {
		return nil
	}
	return r.records[r.idx-1]
}

func (r *staticNeo4jResult) Err() error { return r.err }

func (r *staticNeo4jResult) Close() error { return nil }

var _ db.Neo4jResult = (*staticNeo4jResult)(nil)

func TestGetTopAirports(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/airports/top", api.GetTopAirports())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/airports/top", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got []db.TopAirport
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Greater(t, len(got), 10)
	assert.Equal(t, "ATL", got[0].Code)
}

func TestGetExplore_OriginRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/graph/explore", api.GetExplore(new(mocks.MockNeo4jDB)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/explore", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetExplore_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockNeo4j := new(mocks.MockNeo4jDB)
	router.GET("/api/v1/graph/explore", api.GetExplore(mockNeo4j))

	rec := &neo4j.Record{
		Keys: []string{
			"originCode", "originLat", "originLon",
			"destCode", "destName", "destCity", "destCountry", "destLat", "destLon",
			"cheapestPrice", "hops",
		},
		Values: []any{
			"ORD", 41.9742, -87.9073,
			"LHR", "Heathrow", "London", "GB", 51.47, -0.4543,
			450.0, int64(1),
		},
	}

	result := &staticNeo4jResult{records: []*neo4j.Record{rec}}

	mockNeo4j.
		On("ExecuteReadQuery", mock.Anything, mock.Anything, mock.MatchedBy(func(params map[string]interface{}) bool {
			return params["origin"] == "ORD"
		})).
		Return(result, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/explore?origin=ord&maxHops=1&maxPrice=1000&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ORD", body["origin"])
	assert.Equal(t, float64(1), body["maxHops"])
	assert.Equal(t, float64(1000), body["maxPrice"])
	assert.Equal(t, float64(1), body["count"])

	edges, ok := body["edges"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, edges, 1)

	e0 := edges[0].(map[string]interface{})
	assert.Equal(t, "ORD", e0["origin_code"])
	assert.Equal(t, "LHR", e0["dest_code"])
	assert.Equal(t, float64(450), e0["cheapest_price"])

	mockNeo4j.AssertExpectations(t)
}
