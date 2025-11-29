package geo

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Known airport coordinates for testing
var (
	// JFK - New York John F. Kennedy International Airport
	JFK = Coordinates{Lat: 40.6413, Lon: -73.7781}
	// LAX - Los Angeles International Airport
	LAX = Coordinates{Lat: 33.9425, Lon: -118.4081}
	// LHR - London Heathrow Airport
	LHR = Coordinates{Lat: 51.4700, Lon: -0.4543}
	// SYD - Sydney Kingsford Smith Airport
	SYD = Coordinates{Lat: -33.9399, Lon: 151.1753}
	// NRT - Tokyo Narita International Airport
	NRT = Coordinates{Lat: 35.7720, Lon: 140.3929}
	// DXB - Dubai International Airport
	DXB = Coordinates{Lat: 25.2532, Lon: 55.3657}
)

func TestHaversine_KnownDistances(t *testing.T) {
	tests := []struct {
		name     string
		from     Coordinates
		to       Coordinates
		expected float64 // expected distance in miles
		tolerance float64 // acceptable error margin
	}{
		{
			name:      "JFK to LAX",
			from:      JFK,
			to:        LAX,
			expected:  2475, // approximately 2,475 miles
			tolerance: 25,   // within 25 miles
		},
		{
			name:      "LHR to JFK",
			from:      LHR,
			to:        JFK,
			expected:  3459, // approximately 3,459 miles
			tolerance: 25,
		},
		{
			name:      "LHR to SYD",
			from:      LHR,
			to:        SYD,
			expected:  10573, // approximately 10,573 miles
			tolerance: 50,
		},
		{
			name:      "JFK to NRT",
			from:      JFK,
			to:        NRT,
			expected:  6760, // approximately 6,760 miles
			tolerance: 50,
		},
		{
			name:      "DXB to LHR",
			from:      DXB,
			to:        LHR,
			expected:  3420, // approximately 3,420 miles
			tolerance: 25,
		},
		{
			name:      "Same location (JFK to JFK)",
			from:      JFK,
			to:        JFK,
			expected:  0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := Haversine(tt.from.Lat, tt.from.Lon, tt.to.Lat, tt.to.Lon)
			diff := math.Abs(distance - tt.expected)
			assert.LessOrEqual(t, diff, tt.tolerance,
				"Distance %f should be within %f of %f", distance, tt.tolerance, tt.expected)
		})
	}
}

func TestHaversine_Symmetry(t *testing.T) {
	// Distance from A to B should equal distance from B to A
	distAB := Haversine(JFK.Lat, JFK.Lon, LAX.Lat, LAX.Lon)
	distBA := Haversine(LAX.Lat, LAX.Lon, JFK.Lat, JFK.Lon)

	assert.InDelta(t, distAB, distBA, 0.001, "Distance should be symmetric")
}

func TestHaversineKm(t *testing.T) {
	// JFK to LAX should be approximately 3,983 km
	distance := HaversineKm(JFK.Lat, JFK.Lon, LAX.Lat, LAX.Lon)
	assert.InDelta(t, 3983, distance, 50, "JFK to LAX should be ~3,983 km")
}

func TestCostPerMile(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		distance float64
		expected float64
	}{
		{
			name:     "Normal calculation",
			price:    250.0,
			distance: 2500.0,
			expected: 0.10, // $0.10 per mile
		},
		{
			name:     "Zero distance",
			price:    250.0,
			distance: 0,
			expected: 0, // avoid division by zero
		},
		{
			name:     "Negative distance",
			price:    250.0,
			distance: -100,
			expected: 0, // avoid negative values
		},
		{
			name:     "Zero price",
			price:    0,
			distance: 2500.0,
			expected: 0, // free flight
		},
		{
			name:     "High price short distance",
			price:    500.0,
			distance: 100.0,
			expected: 5.0, // $5.00 per mile
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CostPerMile(tt.price, tt.distance)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestCostPerMileCents(t *testing.T) {
	// $0.10 per mile = 10 cents per mile
	cents := CostPerMileCents(250.0, 2500.0)
	assert.InDelta(t, 10.0, cents, 0.001)
}

func TestDistanceBetween(t *testing.T) {
	distance := DistanceBetween(JFK, LAX)
	directHaversine := Haversine(JFK.Lat, JFK.Lon, LAX.Lat, LAX.Lon)

	assert.Equal(t, directHaversine, distance, "DistanceBetween should match Haversine")
}

func TestCoordinates_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		coords   Coordinates
		expected bool
	}{
		{"Valid JFK", JFK, true},
		{"Valid LAX", LAX, true},
		{"Valid Sydney (negative lat)", SYD, true},
		{"Valid origin", Coordinates{0, 0}, true},
		{"Invalid latitude too high", Coordinates{91, 0}, false},
		{"Invalid latitude too low", Coordinates{-91, 0}, false},
		{"Invalid longitude too high", Coordinates{0, 181}, false},
		{"Invalid longitude too low", Coordinates{0, -181}, false},
		{"Edge case max lat", Coordinates{90, 0}, true},
		{"Edge case min lat", Coordinates{-90, 0}, true},
		{"Edge case max lon", Coordinates{0, 180}, true},
		{"Edge case min lon", Coordinates{0, -180}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.coords.IsValid())
		})
	}
}

func TestCoordinates_IsZero(t *testing.T) {
	assert.True(t, Coordinates{0, 0}.IsZero())
	assert.False(t, JFK.IsZero())
	assert.False(t, Coordinates{0, 1}.IsZero())
	assert.False(t, Coordinates{1, 0}.IsZero())
}

func BenchmarkHaversine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Haversine(JFK.Lat, JFK.Lon, LAX.Lat, LAX.Lon)
	}
}

func BenchmarkCostPerMile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CostPerMile(250.0, 2500.0)
	}
}
