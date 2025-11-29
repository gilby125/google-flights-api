package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/flights"
)

type airport struct {
	Iata string
	Tz   string
	City string
	Lat  float64
	Lon  float64
}

type result struct {
	line string
	err  error
}

func getAirports(commitHash string) (map[string]airport, error) {
	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/mwgg/Airports/%s/airports.json", commitHash))
	if err != nil {
		return nil, err
	}

	airports := map[string]airport{}
	err = json.NewDecoder(resp.Body).Decode(&airports)
	if err != nil {
		return nil, err
	}
	return airports, err
}

func main() {
	commitHash := "f259c38566a5acbcb04b64eb5ad01d14bf7fd07c"

	airports, err := getAirports(commitHash)
	if err != nil {
		log.Fatal(err)
	}

	session, err := flights.New()
	if err != nil {
		log.Fatal(err)
	}

	// We have to iterate over the map every time in the same order because the airport
	// database has a bug where it has the same iata code twice with different timezones.
	keys := make([]string, 0)
	for k := range airports {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(chan result)
	checked := map[string]struct{}{}
	var wg sync.WaitGroup

	caseTmpl := `	case "%s":
		return Location{"%s", "%s", %f, %f}
`

	var a airport
	for _, k := range keys {
		a = airports[k]
		if a.Iata == "" || a.Iata == "0" {
			continue
		}
		if _, ok := checked[a.Iata]; ok {
			continue
		}
		checked[a.Iata] = struct{}{}

		wg.Add(1)
		go func(iata, tz, city string, lat, lon float64) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
			defer cancel()

			ok, err := session.IsIATASupported(ctx, iata)
			if err != nil {
				out <- result{err: err}
			}

			if ok {
				res := result{line: fmt.Sprintf(caseTmpl, iata, city, tz, lat, lon), err: nil}
				out <- res
			}
		}(a.Iata, a.Tz, a.City, a.Lat, a.Lon)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	iataFileContent := fmt.Sprintf(
		`// Package iata contains IATA airport codes, which are supported by the Google Flights API, along with time zones and coordinates.
// This package was generated using an airport list (which can be found at this address: [airports.json])
// and the Google Flights API.
//
// Command: go run ./iata/generate/generate.go
//
// Generation date: %s
//
// [airports.json]: https://github.com/mwgg/Airports/blob/%s/airports.json
package iata

// Location contains airport location data including city, timezone, and coordinates.
type Location struct {
	City string
	Tz   string
	Lat  float64
	Lon  float64
}

// IATATimeZone returns airport location data for the given IATA code.
// If the IATA code is not found, it returns a Location with "Not supported IATA Code" values and zero coordinates.
func IATATimeZone(iata string) Location {
	switch iata {
`, time.Now().Format(time.DateOnly), commitHash)

	lines := []string{}

	for res := range out {
		if res.err != nil {
			log.Fatal(res.err)
		}
		lines = append(lines, res.line)
	}

	sort.Strings(lines)

	for _, line := range lines {
		iataFileContent += line
	}

	iataFileContent += `	}
	return Location{"Not supported IATA Code", "Not supported IATA Code", 0, 0}
}
`

	iataFile, err := os.Create("./iata/iata.go")
	if err != nil {
		log.Fatal(err)
	}
	defer iataFile.Close()

	_, err = iataFile.WriteString(iataFileContent)
	if err != nil {
		log.Fatal(err)
	}
}
