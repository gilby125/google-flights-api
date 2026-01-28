package mocks

import (
	"context"

	"errors"  // Added import
	"reflect" // Added for Scan simulation

	"github.com/gilby125/google-flights-api/db" // Added db import for Neo4jResult interface
	"github.com/gilby125/google-flights-api/queue"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j" // Added neo4j import

	"github.com/stretchr/testify/mock"
)

// --- MockSqlRow Helper ---

// MockSqlRow simulates *sql.Row for testing QueryRowContext(...).Scan(...)
type MockSqlRow struct {
	Values []interface{}
	Err    error
}

// NewMockSqlRow creates a MockSqlRow for testing.
// Pass the values that Scan should populate, followed by the error (if any).
func NewMockSqlRow(values ...interface{}) *MockSqlRow {
	// Check if the last argument is an error
	var err error
	if len(values) > 0 {
		if e, ok := values[len(values)-1].(error); ok {
			err = e
			values = values[:len(values)-1] // Remove error from values
		}
	}
	return &MockSqlRow{Values: values, Err: err}
}

// Scan simulates the Scan method of *sql.Row
func (r *MockSqlRow) Scan(dest ...interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	if len(dest) != len(r.Values) {
		return errors.New("mock Scan: argument count mismatch")
	}
	for i, d := range dest {
		dv := reflect.ValueOf(d)
		if dv.Kind() != reflect.Ptr {
			return errors.New("mock Scan: destination argument must be a pointer")
		}
		sv := reflect.ValueOf(r.Values[i])
		// Handle potential type mismatches or assign directly
		// This is a simplified assignment; real Scan does more complex conversions
		if dv.Elem().Type() == sv.Type() {
			dv.Elem().Set(sv)
		} else {
			// Attempt basic assignment, might panic if types are incompatible
			// A more robust mock would handle type conversions like sql.Scan does
			dv.Elem().Set(sv)
			// return fmt.Errorf("mock Scan: type mismatch for argument %d", i)
		}
	}
	return nil
}

// --- End MockSqlRow Helper ---

// --- MockQueryRowScanner Helper ---

// MockQueryRowScanner simulates a RowScanner for testing QueryRowContext(...).Scan(...)
type MockQueryRowScanner struct {
	mock.Mock
}

// Scan mocks the Scan method of db.RowScanner
func (m *MockQueryRowScanner) Scan(dest ...interface{}) error {
	// Prepare arguments for the mock call
	callArgs := make([]interface{}, len(dest))
	for i, d := range dest {
		callArgs[i] = d
	}
	args := m.Called(callArgs...) // Pass dest as variadic arguments

	// Simulate assignment if needed (optional, depends on test needs)
	// Example: If the mock is configured to return values for Scan
	// retVals := args.Get(1).([]interface{}) // Assuming second return is values
	// if len(dest) == len(retVals) {
	//     for i, d := range dest {
	//         // Simplified assignment logic
	//         reflect.ValueOf(d).Elem().Set(reflect.ValueOf(retVals[i]))
	//     }
	// }

	return args.Error(0) // Return the configured error
}

// Ensure MockQueryRowScanner implements db.RowScanner
var _ db.RowScanner = (*MockQueryRowScanner)(nil)

// --- End MockQueryRowScanner Helper ---

type Worker struct {
	mock.Mock
}

func (m *Worker) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Worker) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Worker) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *Worker) Process(jobID string, payload []byte) error {
	args := m.Called(jobID, payload)
	return args.Error(0)
}

type Scheduler struct {
	mock.Mock
}

func (m *Scheduler) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Scheduler) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Scheduler) AddJob(job interface{}) error {
	args := m.Called(job)
	return args.Error(0)
}

// --- Neo4j Mocks ---

// MockNeo4jResult implements db.Neo4jResult
type MockNeo4jResult struct {
	mock.Mock
	Records []*neo4j.Record
	Current int
	Error   error
}

func (m *MockNeo4jResult) Next() bool { // Removed ctx parameter
	args := m.Called() // Removed ctx from Called()
	// Simulate iteration based on mock setup or internal state
	if m.Error != nil {
		return false
	}
	hasNext := m.Current < len(m.Records)
	if hasNext {
		m.Current++
	}
	return args.Bool(0) // Allow overriding via mock setup if needed, otherwise use internal state
}

func (m *MockNeo4jResult) Record() *neo4j.Record {
	args := m.Called()
	// Return the current record based on internal state or mock setup
	if m.Current > 0 && m.Current <= len(m.Records) {
		return args.Get(0).(*neo4j.Record) // Allow overriding
	}
	return nil // Or return the actual record: m.Records[m.Current-1]
}

func (m *MockNeo4jResult) Err() error {
	args := m.Called()
	if m.Error != nil {
		return m.Error
	}
	return args.Error(0) // Allow overriding
}

func (m *MockNeo4jResult) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockNeo4jResult implements db.Neo4jResult
var _ db.Neo4jResult = (*MockNeo4jResult)(nil)

// MockNeo4jDB implements db.Neo4jDatabase
type MockNeo4jDB struct {
	mock.Mock
}

func (m *MockNeo4jDB) CreateAirport(code string, name string, city string, country string, lat float64, lon float64) error {
	args := m.Called(code, name, city, country, lat, lon)
	return args.Error(0)
}

func (m *MockNeo4jDB) CreateRoute(origin string, dest string, airline string, flightNum string, price float64, duration int) error {
	args := m.Called(origin, dest, airline, flightNum, price, duration)
	return args.Error(0)
}

func (m *MockNeo4jDB) AddPricePoint(origin string, dest string, departDate string, returnDate string, price float64, airline string, tripType string) error {
	args := m.Called(origin, dest, departDate, returnDate, price, airline, tripType)
	return args.Error(0)
}

func (m *MockNeo4jDB) CreateAirline(code, name, country string) error {
	args := m.Called(code, name, country)
	return args.Error(0)
}

func (m *MockNeo4jDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNeo4jDB) ExecuteReadQuery(ctx context.Context, query string, params map[string]interface{}) (db.Neo4jResult, error) {
	args := m.Called(ctx, query, params)
	var result db.Neo4jResult
	if r := args.Get(0); r != nil {
		result = r.(db.Neo4jResult) // Expecting *MockNeo4jResult
	}
	return result, args.Error(1)
}

func (m *MockNeo4jDB) InitSchema() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNeo4jDB) FindCheapestPath(ctx context.Context, origin, dest string, maxHops int, maxPrice float64) ([]db.PathResult, error) {
	args := m.Called(ctx, origin, dest, maxHops, maxPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.PathResult), args.Error(1)
}

func (m *MockNeo4jDB) FindConnections(ctx context.Context, origin string, maxHops int, maxPrice float64) ([]db.Connection, error) {
	args := m.Called(ctx, origin, maxHops, maxPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.Connection), args.Error(1)
}

func (m *MockNeo4jDB) GetRouteStats(ctx context.Context, origin, dest string) (*db.RouteStats, error) {
	args := m.Called(ctx, origin, dest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.RouteStats), args.Error(1)
}

// Ensure MockNeo4jDB implements db.Neo4jDatabase
var _ db.Neo4jDatabase = (*MockNeo4jDB)(nil)

// --- Queue Mock ---

type Queue struct {
	mock.Mock
}

// Enqueue matches the Queue interface: Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error)
func (m *Queue) Enqueue(ctx context.Context, jobType string, payload interface{}) (string, error) {
	args := m.Called(ctx, jobType, payload)
	return args.String(0), args.Error(1)
}

func (m *Queue) Dequeue(ctx context.Context, queueName string) (*queue.Job, error) {
	args := m.Called(ctx, queueName)
	jobArg := args.Get(0)
	if jobArg == nil {
		return nil, args.Error(1)
	}
	return jobArg.(*queue.Job), args.Error(1)
}

func (m *Queue) Ack(ctx context.Context, queueName string, jobID string) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

// MockRows simulates sql.Rows for testing database interactions
type MockRows struct {
	mock.Mock
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	// Prepare arguments for the mock call, converting dest to []interface{}
	scanArgs := make([]interface{}, len(dest))
	for i, d := range dest {
		scanArgs[i] = d
	}
	args := m.Called(scanArgs...) // Pass dest as variadic arguments
	return args.Error(0)
}

func (m *MockRows) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Queue) Nack(ctx context.Context, queueName string, jobID string) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

// GetJobStatus matches the Queue interface: GetJobStatus(ctx context.Context, jobID string) (string, error)
func (m *Queue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	args := m.Called(ctx, jobID)
	return args.String(0), args.Error(1)
}

// GetQueueStats matches the Queue interface: GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error)
func (m *Queue) GetQueueStats(ctx context.Context, queueName string) (map[string]int64, error) {
	args := m.Called(ctx, queueName)
	// Safely cast the first argument to map[string]int64
	stats, _ := args.Get(0).(map[string]int64)
	return stats, args.Error(1)
}

func (m *Queue) CancelJob(ctx context.Context, queueName, jobID string) error {
	args := m.Called(ctx, queueName, jobID)
	return args.Error(0)
}

func (m *Queue) CancelProcessing(ctx context.Context, queueName string) (int64, error) {
	args := m.Called(ctx, queueName)
	var canceled int64
	if v := args.Get(0); v != nil {
		canceled, _ = v.(int64)
	}
	return canceled, args.Error(1)
}

func (m *Queue) IsJobCanceled(ctx context.Context, jobID string) (bool, error) {
	args := m.Called(ctx, jobID)
	return args.Bool(0), args.Error(1)
}

func (m *Queue) ClearQueue(ctx context.Context, queueName string) (int64, error) {
	args := m.Called(ctx, queueName)
	var cleared int64
	if v := args.Get(0); v != nil {
		cleared, _ = v.(int64)
	}
	return cleared, args.Error(1)
}

func (m *Queue) GetJob(ctx context.Context, jobID string) (*queue.Job, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*queue.Job), args.Error(1)
}

func (m *Queue) ListJobs(ctx context.Context, queueName, state string, limit, offset int) ([]*queue.Job, error) {
	args := m.Called(ctx, queueName, state, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*queue.Job), args.Error(1)
}

func (m *Queue) GetBacklog(ctx context.Context, queueName string, limit int) ([]*queue.Job, error) {
	args := m.Called(ctx, queueName, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*queue.Job), args.Error(1)
}

func (m *Queue) GetEnqueueMetrics(ctx context.Context, queueName string, minutes int) (map[string]int64, error) {
	args := m.Called(ctx, queueName, minutes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *Queue) ClearFailed(ctx context.Context, queueName string) (int64, error) {
	args := m.Called(ctx, queueName)
	var cleared int64
	if v := args.Get(0); v != nil {
		cleared, _ = v.(int64)
	}
	return cleared, args.Error(1)
}

func (m *Queue) ClearProcessing(ctx context.Context, queueName string) (int64, error) {
	args := m.Called(ctx, queueName)
	var cleared int64
	if v := args.Get(0); v != nil {
		cleared, _ = v.(int64)
	}
	return cleared, args.Error(1)
}

func (m *Queue) RetryFailed(ctx context.Context, queueName string, limit int) (int64, error) {
	args := m.Called(ctx, queueName, limit)
	var retried int64
	if v := args.Get(0); v != nil {
		retried, _ = v.(int64)
	}
	return retried, args.Error(1)
}
