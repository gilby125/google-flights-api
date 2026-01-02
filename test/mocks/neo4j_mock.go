package mocks

import (
	"context"

	"github.com/gilby125/google-flights-api/db" // Adjust import path if necessary
	"github.com/stretchr/testify/mock"
)

// MockNeo4jDatabase is a mock implementation of the db.Neo4jDatabase interface
type MockNeo4jDatabase struct {
	mock.Mock
}

// CreateAirport mocks the CreateAirport method
func (m *MockNeo4jDatabase) CreateAirport(code, name, city, country string, latitude, longitude float64) error {
	args := m.Called(code, name, city, country, latitude, longitude)
	return args.Error(0)
}

// CreateRoute mocks the CreateRoute method
func (m *MockNeo4jDatabase) CreateRoute(originCode, destCode, airlineCode, flightNumber string, avgPrice float64, avgDuration int) error {
	args := m.Called(originCode, destCode, airlineCode, flightNumber, avgPrice, avgDuration)
	return args.Error(0)
}

// AddPricePoint mocks the AddPricePoint method
func (m *MockNeo4jDatabase) AddPricePoint(originCode, destCode string, date string, price float64, airlineCode string) error {
	args := m.Called(originCode, destCode, date, price, airlineCode)
	return args.Error(0)
}

// CreateAirline mocks the CreateAirline method
func (m *MockNeo4jDatabase) CreateAirline(code, name, country string) error {
	args := m.Called(code, name, country)
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockNeo4jDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ExecuteReadQuery mocks the ExecuteReadQuery method
func (m *MockNeo4jDatabase) ExecuteReadQuery(ctx context.Context, query string, params map[string]interface{}) (db.Neo4jResult, error) {
	args := m.Called(ctx, query, params)
	// Need to handle the case where the first return value (Neo4jResult) is nil
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(db.Neo4jResult), args.Error(1)
}

// InitSchema mocks the InitSchema method
func (m *MockNeo4jDatabase) InitSchema() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockNeo4jDatabase implements the interface
var _ db.Neo4jDatabase = (*MockNeo4jDatabase)(nil)
