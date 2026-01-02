package flights

import (
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/iata"
)

type iataLocationCacheEntry struct {
	city string
	loc  *time.Location
}

var iataLocationCache sync.Map // map[string]iataLocationCacheEntry

func iataLocationCached(iataCode string) (string, *time.Location) {
	if cached, ok := iataLocationCache.Load(iataCode); ok {
		entry := cached.(iataLocationCacheEntry)
		return entry.city, entry.loc
	}

	iataLocation := iata.IATATimeZone(iataCode)
	location, err := time.LoadLocation(iataLocation.Tz)
	if err != nil {
		location = time.UTC
	}

	entry := iataLocationCacheEntry{
		city: iataLocation.City,
		loc:  location,
	}
	iataLocationCache.Store(iataCode, entry)
	return entry.city, entry.loc
}
