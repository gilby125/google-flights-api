// Package geo provides geographic distance calculations.
package geo

import "math"

const (
	// EarthRadiusMiles is the mean radius of Earth in miles.
	EarthRadiusMiles = 3958.8
	// EarthRadiusKm is the mean radius of Earth in kilometers.
	EarthRadiusKm = 6371.0
)

// Haversine calculates the great-circle distance between two points
// on Earth given their latitude and longitude in decimal degrees.
// Returns the distance in miles.
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	return HaversineWithRadius(lat1, lon1, lat2, lon2, EarthRadiusMiles)
}

// HaversineKm calculates the great-circle distance in kilometers.
func HaversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	return HaversineWithRadius(lat1, lon1, lat2, lon2, EarthRadiusKm)
}

// HaversineWithRadius calculates the great-circle distance using a custom radius.
func HaversineWithRadius(lat1, lon1, lat2, lon2, radius float64) float64 {
	// Convert degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)
	deltaLat := degreesToRadians(lat2 - lat1)
	deltaLon := degreesToRadians(lon2 - lon1)

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radius * c
}

// CostPerMile calculates the cost per mile for a flight.
// Returns 0 if distance is zero or negative to avoid division by zero.
func CostPerMile(price, distanceMiles float64) float64 {
	if distanceMiles <= 0 {
		return 0
	}
	return price / distanceMiles
}

// CostPerMileCents returns cost per mile in cents (price * 100 / distance).
// This is useful for displaying values like "$0.05/mile" as "5 cents/mile".
func CostPerMileCents(price, distanceMiles float64) float64 {
	return CostPerMile(price, distanceMiles) * 100
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Coordinates represents a geographic point.
type Coordinates struct {
	Lat float64
	Lon float64
}

// DistanceBetween calculates the distance in miles between two coordinate points.
func DistanceBetween(from, to Coordinates) float64 {
	return Haversine(from.Lat, from.Lon, to.Lat, to.Lon)
}

// IsValid returns true if the coordinates are within valid ranges.
// Latitude must be between -90 and 90, longitude between -180 and 180.
func (c Coordinates) IsValid() bool {
	return c.Lat >= -90 && c.Lat <= 90 && c.Lon >= -180 && c.Lon <= 180
}

// IsZero returns true if both coordinates are zero (likely unset).
func (c Coordinates) IsZero() bool {
	return c.Lat == 0 && c.Lon == 0
}
