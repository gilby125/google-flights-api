package mocks

import (
	"context"
	"database/sql"
	"time"

	"github.com/gilby125/google-flights-api/db"
	// "github.com/gilby125/google-flights-api/flights" // Removed unused import
	"github.com/stretchr/testify/mock"
)

// MockTx implements db.Tx
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(ctx, query, args)
	var res sql.Result
	// Handle potential nil result from mock setup
	if r := callArgs.Get(0); r != nil {
		res = r.(sql.Result)
	}
	return res, callArgs.Error(1)
}

// QueryRowContext mocks base method, updated to match db.Tx interface.
func (m *MockTx) QueryRowContext(ctx context.Context, query string, args ...any) db.RowScanner {
	// Prepare arguments for mock call, handling variadic args correctly
	variadicArgs := make([]interface{}, len(args))
	for i, a := range args {
		variadicArgs[i] = a
	}
	allArgs := append([]interface{}{ctx, query}, variadicArgs...)

	callArgs := m.Called(allArgs...)

	// Return the mock object directly if it's not nil.
	// The test setup should provide an object that satisfies db.RowScanner (e.g., *MockSqlRow).
	if r := callArgs.Get(0); r != nil {
		// Assert that the returned object implements db.RowScanner
		scanner, ok := r.(db.RowScanner)
		if !ok {
			// This indicates an issue with the test setup - the mock didn't return a RowScanner.
			// Panic or return a specific error to highlight the test setup problem.
			panic("MockTx.QueryRowContext mock setup returned non-RowScanner type")
		}
		return scanner
	}
	return nil // Return nil if mock setup returns nil
}

// Ensure MockTx implements db.Tx
var _ db.Tx = (*MockTx)(nil)

// MockRows definition removed. It is defined in mocks.go.
// Helper functions ExpectScanRow and SetupRows also removed as they depend on MockRows.

// MockPostgresDB implements db.PostgresDB
type MockPostgresDB struct {
	mock.Mock
}

// --- Implementation of db.PostgresDB Interface ---

func (m *MockPostgresDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Existing methods
func (m *MockPostgresDB) GetSearchByID(id string) (*db.Search, error) {
	args := m.Called(id)
	// Keep existing methods if they are still part of the interface or used elsewhere
	var searchArg *db.Search
	if search, ok := args.Get(0).(*db.Search); ok {
		searchArg = search
	}
	return searchArg, args.Error(1)
}

func (m *MockPostgresDB) SaveSearch(search *db.Search) (string, error) {
	args := m.Called(search)
	return args.String(0), args.Error(1)
}

func (m *MockPostgresDB) Search(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	var rows *sql.Rows
	if r := callArgs.Get(0); r != nil {
		rows = r.(*sql.Rows)
	}
	return rows, callArgs.Error(1)
}

func (m *MockPostgresDB) GetCertificate(domain string) (*db.Certificate, error) {
	callArgs := m.Called(domain)
	var cert *db.Certificate
	if c := callArgs.Get(0); c != nil {
		cert = c.(*db.Certificate)
	}
	return cert, callArgs.Error(1)
}

func (m *MockPostgresDB) StoreCertificate(domain string, cert []byte, key []byte, expires time.Time) error {
	callArgs := m.Called(domain, cert, key, expires)
	return callArgs.Error(0)
}

func (m *MockPostgresDB) InitSchema() error {
	args := m.Called()
	return args.Error(0)
}

// --- New/Updated Methods from Refactored Interface ---

func (m *MockPostgresDB) BeginTx(ctx context.Context) (db.Tx, error) {
	args := m.Called(ctx)
	var tx db.Tx
	if t := args.Get(0); t != nil {
		tx = t.(db.Tx) // Expecting a db.Tx (e.g., *MockTx)
	}
	return tx, args.Error(1)
}

func (m *MockPostgresDB) QueryAirports(ctx context.Context) (db.Rows, error) {
	args := m.Called(ctx)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows) // Expecting a db.Rows (e.g., *MockRows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) QueryAirlines(ctx context.Context) (db.Rows, error) {
	args := m.Called(ctx)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) GetSearchQueryByID(ctx context.Context, id int) (*db.SearchQuery, error) {
	args := m.Called(ctx, id)
	var query *db.SearchQuery
	if q := args.Get(0); q != nil {
		query = q.(*db.SearchQuery)
	}
	return query, args.Error(1)
}

func (m *MockPostgresDB) GetFlightOffersBySearchID(ctx context.Context, searchID int) (db.Rows, error) {
	args := m.Called(ctx, searchID)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) GetFlightSegmentsByOfferID(ctx context.Context, offerID int) (db.Rows, error) {
	args := m.Called(ctx, offerID)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) CountSearches(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) QuerySearchesPaginated(ctx context.Context, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) DeleteJobDetailsByJobID(ctx context.Context, tx db.Tx, jobID int) error {
	args := m.Called(ctx, tx, jobID)
	return args.Error(0)
}

func (m *MockPostgresDB) DeleteScheduledJobByID(ctx context.Context, tx db.Tx, jobID int) (int64, error) {
	args := m.Called(ctx, tx, jobID)
	// Return the configured values directly
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostgresDB) GetJobDetailsByID(ctx context.Context, jobID int) (*db.JobDetails, error) {
	args := m.Called(ctx, jobID)
	var details *db.JobDetails
	if d := args.Get(0); d != nil {
		details = d.(*db.JobDetails)
	}
	return details, args.Error(1)
}

func (m *MockPostgresDB) UpdateJobLastRun(ctx context.Context, jobID int) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockPostgresDB) UpdateJobEnabled(ctx context.Context, jobID int, enabled bool) (int64, error) {
	args := m.Called(ctx, jobID, enabled)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockPostgresDB) GetJobCronExpression(ctx context.Context, jobID int) (string, error) {
	args := m.Called(ctx, jobID)
	return args.String(0), args.Error(1)
}

func (m *MockPostgresDB) ListJobs(ctx context.Context) (db.Rows, error) {
	args := m.Called(ctx)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) CreateScheduledJob(ctx context.Context, tx db.Tx, name, cronExpression string, enabled bool) (int, error) {
	args := m.Called(ctx, tx, name, cronExpression, enabled)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) CreateJobDetails(ctx context.Context, tx db.Tx, details db.JobDetails) error {
	args := m.Called(ctx, tx, details)
	return args.Error(0)
}

func (m *MockPostgresDB) UpdateScheduledJob(ctx context.Context, tx db.Tx, jobID int, name, cronExpression string) error {
	args := m.Called(ctx, tx, jobID, name, cronExpression)
	return args.Error(0)
}

func (m *MockPostgresDB) UpdateJobDetails(ctx context.Context, tx db.Tx, jobID int, details db.JobDetails) error {
	args := m.Called(ctx, tx, jobID, details)
	return args.Error(0)
}

func (m *MockPostgresDB) GetJobByID(ctx context.Context, jobID int) (*db.ScheduledJob, error) {
	args := m.Called(ctx, jobID)
	var job *db.ScheduledJob
	if j := args.Get(0); j != nil {
		job = j.(*db.ScheduledJob)
	}
	return job, args.Error(1)
}

func (m *MockPostgresDB) GetBulkSearchByID(ctx context.Context, searchID int) (*db.BulkSearch, error) {
	args := m.Called(ctx, searchID)
	var search *db.BulkSearch
	if s := args.Get(0); s != nil {
		search = s.(*db.BulkSearch)
	}
	return search, args.Error(1)
}

func (m *MockPostgresDB) QueryBulkSearchResultsPaginated(ctx context.Context, searchID, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, searchID, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) InsertBulkSearchOffer(ctx context.Context, record db.BulkSearchOfferRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockPostgresDB) QueryBulkSearchOffersPaginated(ctx context.Context, searchID, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, searchID, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) CreateBulkSearchRecord(ctx context.Context, jobID sql.NullInt32, totalSearches int, currency, status string) (int, error) {
	args := m.Called(ctx, jobID, totalSearches, currency, status)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) UpdateBulkSearchStatus(ctx context.Context, bulkSearchID int, status string) error {
	args := m.Called(ctx, bulkSearchID, status)
	return args.Error(0)
}

func (m *MockPostgresDB) CompleteBulkSearch(ctx context.Context, summary db.BulkSearchSummary) error {
	args := m.Called(ctx, summary)
	return args.Error(0)
}

func (m *MockPostgresDB) InsertBulkSearchResult(ctx context.Context, result db.BulkSearchResultRecord) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockPostgresDB) ListBulkSearches(ctx context.Context, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) InsertBulkSearchResultsBatch(ctx context.Context, results []db.BulkSearchResultRecord) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockPostgresDB) ListBulkSearchOffers(ctx context.Context, searchID int) ([]db.BulkSearchOffer, error) {
	args := m.Called(ctx, searchID)
	var offers []db.BulkSearchOffer
	if o := args.Get(0); o != nil {
		offers = o.([]db.BulkSearchOffer)
	}
	return offers, args.Error(1)
}

func (m *MockPostgresDB) CreatePriceGraphSweep(ctx context.Context, jobID sql.NullInt32, originCount, destinationCount int, tripLengthMin, tripLengthMax sql.NullInt32, currency string) (int, error) {
	args := m.Called(ctx, jobID, originCount, destinationCount, tripLengthMin, tripLengthMax, currency)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) UpdatePriceGraphSweepStatus(ctx context.Context, sweepID int, status string, startedAt, completedAt sql.NullTime, errorCount int) error {
	args := m.Called(ctx, sweepID, status, startedAt, completedAt, errorCount)
	return args.Error(0)
}

func (m *MockPostgresDB) GetPriceGraphSweepByID(ctx context.Context, sweepID int) (*db.PriceGraphSweep, error) {
	args := m.Called(ctx, sweepID)
	var sweep *db.PriceGraphSweep
	if s := args.Get(0); s != nil {
		sweep = s.(*db.PriceGraphSweep)
	}
	return sweep, args.Error(1)
}

func (m *MockPostgresDB) ListPriceGraphSweeps(ctx context.Context, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) InsertPriceGraphResult(ctx context.Context, record db.PriceGraphResultRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockPostgresDB) ListPriceGraphResults(ctx context.Context, sweepID, limit, offset int) (db.Rows, error) {
	args := m.Called(ctx, sweepID, limit, offset)
	var rows db.Rows
	if r := args.Get(0); r != nil {
		rows = r.(db.Rows)
	}
	return rows, args.Error(1)
}

func (m *MockPostgresDB) SaveContinuousSweepProgress(ctx context.Context, progress db.ContinuousSweepProgress) error {
	args := m.Called(ctx, progress)
	return args.Error(0)
}

func (m *MockPostgresDB) GetContinuousSweepProgress(ctx context.Context) (*db.ContinuousSweepProgress, error) {
	args := m.Called(ctx)
	var progress *db.ContinuousSweepProgress
	if p := args.Get(0); p != nil {
		progress = p.(*db.ContinuousSweepProgress)
	}
	return progress, args.Error(1)
}

func (m *MockPostgresDB) InsertContinuousSweepStats(ctx context.Context, stats db.ContinuousSweepStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *MockPostgresDB) ListContinuousSweepStats(ctx context.Context, limit int) ([]db.ContinuousSweepStats, error) {
	args := m.Called(ctx, limit)
	var statsList []db.ContinuousSweepStats
	if s := args.Get(0); s != nil {
		statsList = s.([]db.ContinuousSweepStats)
	}
	return statsList, args.Error(1)
}

func (m *MockPostgresDB) ListContinuousSweepResults(ctx context.Context, filters db.ContinuousSweepResultsFilter) ([]db.PriceGraphResultRecord, error) {
	args := m.Called(ctx, filters)
	var results []db.PriceGraphResultRecord
	if r := args.Get(0); r != nil {
		results = r.([]db.PriceGraphResultRecord)
	}
	return results, args.Error(1)
}

// --- Deal Detection Methods Mock Implementations ---

func (m *MockPostgresDB) GetRouteBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*db.RouteBaseline, error) {
	args := m.Called(ctx, origin, dest, tripLength, class)
	var baseline *db.RouteBaseline
	if b := args.Get(0); b != nil {
		baseline = b.(*db.RouteBaseline)
	}
	return baseline, args.Error(1)
}

func (m *MockPostgresDB) UpsertRouteBaseline(ctx context.Context, baseline db.RouteBaseline) error {
	args := m.Called(ctx, baseline)
	return args.Error(0)
}

func (m *MockPostgresDB) GetPriceHistoryForRoute(ctx context.Context, origin, dest string, tripLength int, class string, windowDays int) ([]float64, error) {
	args := m.Called(ctx, origin, dest, tripLength, class, windowDays)
	var prices []float64
	if p := args.Get(0); p != nil {
		prices = p.([]float64)
	}
	return prices, args.Error(1)
}

func (m *MockPostgresDB) InsertDetectedDeal(ctx context.Context, deal db.DetectedDeal) (int, error) {
	args := m.Called(ctx, deal)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) UpsertDetectedDeal(ctx context.Context, deal db.DetectedDeal) error {
	args := m.Called(ctx, deal)
	return args.Error(0)
}

func (m *MockPostgresDB) GetDetectedDealByFingerprint(ctx context.Context, fingerprint string) (*db.DetectedDeal, error) {
	args := m.Called(ctx, fingerprint)
	var deal *db.DetectedDeal
	if d := args.Get(0); d != nil {
		deal = d.(*db.DetectedDeal)
	}
	return deal, args.Error(1)
}

func (m *MockPostgresDB) ListActiveDeals(ctx context.Context, filter db.DealFilter) ([]db.DetectedDeal, error) {
	args := m.Called(ctx, filter)
	var deals []db.DetectedDeal
	if d := args.Get(0); d != nil {
		deals = d.([]db.DetectedDeal)
	}
	return deals, args.Error(1)
}

func (m *MockPostgresDB) ExpireOldDeals(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostgresDB) InsertDealAlert(ctx context.Context, alert db.DealAlert) (int, error) {
	args := m.Called(ctx, alert)
	return args.Int(0), args.Error(1)
}

func (m *MockPostgresDB) ListDealAlerts(ctx context.Context, limit, offset int) ([]db.DealAlert, error) {
	args := m.Called(ctx, limit, offset)
	var alerts []db.DealAlert
	if a := args.Get(0); a != nil {
		alerts = a.([]db.DealAlert)
	}
	return alerts, args.Error(1)
}

// Ensure MockPostgresDB implements db.PostgresDB
var _ db.PostgresDB = (*MockPostgresDB)(nil)
