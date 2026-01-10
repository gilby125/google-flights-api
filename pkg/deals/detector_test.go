package deals

import (
	"context"
	"testing"

	"github.com/gilby125/google-flights-api/db"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockBaselineStore struct {
	mock.Mock
}

func (m *mockBaselineStore) GetRouteBaseline(ctx context.Context, origin, dest string, tripLength int, class string) (*db.RouteBaseline, error) {
	args := m.Called(ctx, origin, dest, tripLength, class)
	var baseline *db.RouteBaseline
	if b := args.Get(0); b != nil {
		baseline = b.(*db.RouteBaseline)
	}
	return baseline, args.Error(1)
}

func (m *mockBaselineStore) UpsertRouteBaseline(ctx context.Context, baseline db.RouteBaseline) error {
	args := m.Called(ctx, baseline)
	return args.Error(0)
}

func (m *mockBaselineStore) GetPriceHistoryForRoute(ctx context.Context, origin, dest string, tripLength int, class string, windowDays int) ([]float64, error) {
	args := m.Called(ctx, origin, dest, tripLength, class, windowDays)
	var prices []float64
	if p := args.Get(0); p != nil {
		prices = p.([]float64)
	}
	return prices, args.Error(1)
}

func TestDealDetector_GetBaseline_UsesAllHistoryWhenWindowDaysZero(t *testing.T) {
	t.Parallel()

	mockDB := &mockBaselineStore{}
	cfg := DefaultDealConfig()
	cfg.BaselineMinSamples = 3

	detector := NewDealDetector(mockDB, cfg)

	mockDB.On("GetRouteBaseline", mock.Anything, "SFO", "LAX", 7, "economy").
		Return((*db.RouteBaseline)(nil), nil).
		Once()
	mockDB.On("GetPriceHistoryForRoute", mock.Anything, "SFO", "LAX", 7, "economy", 0).
		Return([]float64{500, 450, 400}, nil).
		Once()
	mockDB.On("UpsertRouteBaseline", mock.Anything, mock.Anything).
		Return(nil).
		Once()

	baseline, err := detector.getBaseline(context.Background(), "SFO", "LAX", 7, "economy")
	require.NoError(t, err)
	require.NotNil(t, baseline)
	require.Equal(t, 3, baseline.SampleCount)

	mockDB.AssertExpectations(t)
}
