package db

// TopAirport represents an airport with its country code for international filtering
type TopAirport struct {
	Code    string
	Country string
}

// Top100Airports contains IATA codes and country codes for major global airports.
// Despite the name, the list currently has 95 airports (not 100).
// Route calculations:
//   - All routes (n*(n-1)): 95*94 = 8,930 routes
//   - International routes (different countries): ~8,096 routes
//   - With 2 trip lengths (7,14 days): ~16,192 queries per sweep
// Source: ACI World Airport Traffic Rankings 2023
var Top100Airports = []TopAirport{
	// North America - United States (25)
	{"ATL", "US"}, // Atlanta
	{"DFW", "US"}, // Dallas/Fort Worth
	{"DEN", "US"}, // Denver
	{"LAX", "US"}, // Los Angeles
	{"ORD", "US"}, // Chicago O'Hare
	{"JFK", "US"}, // New York JFK
	{"SFO", "US"}, // San Francisco
	{"LAS", "US"}, // Las Vegas
	{"MIA", "US"}, // Miami
	{"CLT", "US"}, // Charlotte
	{"SEA", "US"}, // Seattle
	{"PHX", "US"}, // Phoenix
	{"EWR", "US"}, // Newark
	{"MCO", "US"}, // Orlando
	{"MSP", "US"}, // Minneapolis
	{"BOS", "US"}, // Boston
	{"DTW", "US"}, // Detroit
	{"IAH", "US"}, // Houston
	{"BWI", "US"}, // Baltimore
	{"FLL", "US"}, // Fort Lauderdale
	{"TPA", "US"}, // Tampa
	{"SAN", "US"}, // San Diego
	{"PDX", "US"}, // Portland
	{"STL", "US"}, // St. Louis
	{"HNL", "US"}, // Honolulu

	// North America - Canada & Mexico (3)
	{"YYZ", "CA"}, // Toronto
	{"MEX", "MX"}, // Mexico City
	{"CUN", "MX"}, // Cancun

	// Europe - United Kingdom (4)
	{"LHR", "GB"}, // London Heathrow
	{"LGW", "GB"}, // London Gatwick
	{"MAN", "GB"}, // Manchester
	{"STN", "GB"}, // London Stansted

	// Europe - France (2)
	{"CDG", "FR"}, // Paris Charles de Gaulle
	{"ORY", "FR"}, // Paris Orly

	// Europe - Germany (4)
	{"FRA", "DE"}, // Frankfurt
	{"MUC", "DE"}, // Munich
	{"DUS", "DE"}, // Dusseldorf
	{"HAM", "DE"}, // Hamburg

	// Europe - Other Western Europe (12)
	{"AMS", "NL"}, // Amsterdam
	{"MAD", "ES"}, // Madrid
	{"BCN", "ES"}, // Barcelona
	{"PMI", "ES"}, // Palma de Mallorca
	{"FCO", "IT"}, // Rome Fiumicino
	{"MXP", "IT"}, // Milan Malpensa
	{"ZRH", "CH"}, // Zurich
	{"VIE", "AT"}, // Vienna
	{"LIS", "PT"}, // Lisbon
	{"OSL", "NO"}, // Oslo
	{"CPH", "DK"}, // Copenhagen
	{"ARN", "SE"}, // Stockholm

	// Europe - Other (9)
	{"DUB", "IE"}, // Dublin
	{"ATH", "GR"}, // Athens
	{"BRU", "BE"}, // Brussels
	{"WAW", "PL"}, // Warsaw
	{"IST", "TR"}, // Istanbul
	{"AYT", "TR"}, // Antalya
	{"SAW", "TR"}, // Istanbul Sabiha
	{"DME", "RU"}, // Moscow Domodedovo
	{"SVO", "RU"}, // Moscow Sheremetyevo

	// Asia - Japan & Korea (3)
	{"HND", "JP"}, // Tokyo Haneda
	{"NRT", "JP"}, // Tokyo Narita
	{"ICN", "KR"}, // Seoul Incheon

	// Asia - China (14)
	{"PVG", "CN"}, // Shanghai Pudong
	{"PEK", "CN"}, // Beijing Capital
	{"CAN", "CN"}, // Guangzhou
	{"SHA", "CN"}, // Shanghai Hongqiao
	{"SZX", "CN"}, // Shenzhen
	{"CTU", "CN"}, // Chengdu
	{"KMG", "CN"}, // Kunming
	{"XIY", "CN"}, // Xi'an
	{"HGH", "CN"}, // Hangzhou
	{"CKG", "CN"}, // Chongqing
	{"NKG", "CN"}, // Nanjing
	{"TAO", "CN"}, // Qingdao
	{"WUH", "CN"}, // Wuhan
	{"CSX", "CN"}, // Changsha

	// Asia - Greater China (2)
	{"HKG", "HK"}, // Hong Kong
	{"TPE", "TW"}, // Taipei

	// Asia - Southeast Asia (5)
	{"SIN", "SG"}, // Singapore
	{"BKK", "TH"}, // Bangkok
	{"KUL", "MY"}, // Kuala Lumpur
	{"MNL", "PH"}, // Manila
	{"CGK", "ID"}, // Jakarta

	// Asia - South Asia (2)
	{"DEL", "IN"}, // Delhi
	{"BOM", "IN"}, // Mumbai

	// Asia - Middle East (3)
	{"DXB", "AE"}, // Dubai
	{"DOH", "QA"}, // Doha
	{"AUH", "AE"}, // Abu Dhabi

	// Oceania (2)
	{"SYD", "AU"}, // Sydney
	{"MEL", "AU"}, // Melbourne

	// South America (2)
	{"GRU", "BR"}, // Sao Paulo
	{"BOG", "CO"}, // Bogota

	// Africa (3)
	{"JNB", "ZA"}, // Johannesburg
	{"CMN", "MA"}, // Casablanca
	{"CAI", "EG"}, // Cairo
}

// Route represents an origin-destination pair
type Route struct {
	Origin      string
	Destination string
}

// GenerateInternationalRoutes returns all origin-destination pairs where countries differ
func GenerateInternationalRoutes(airports []TopAirport) []Route {
	var routes []Route
	for _, origin := range airports {
		for _, dest := range airports {
			if origin.Code != dest.Code && origin.Country != dest.Country {
				routes = append(routes, Route{
					Origin:      origin.Code,
					Destination: dest.Code,
				})
			}
		}
	}
	return routes
}

// GenerateAllRoutes returns all origin-destination pairs (including domestic)
func GenerateAllRoutes(airports []TopAirport) []Route {
	var routes []Route
	for _, origin := range airports {
		for _, dest := range airports {
			if origin.Code != dest.Code {
				routes = append(routes, Route{
					Origin:      origin.Code,
					Destination: dest.Code,
				})
			}
		}
	}
	return routes
}

// GetAirportCountry returns the country code for an airport, or empty string if not found
func GetAirportCountry(code string) string {
	for _, airport := range Top100Airports {
		if airport.Code == code {
			return airport.Country
		}
	}
	return ""
}

// CountRoutes returns statistics about route generation
func CountRoutes() (total int, international int, domestic int) {
	countryAirports := make(map[string]int)
	for _, airport := range Top100Airports {
		countryAirports[airport.Country]++
	}

	n := len(Top100Airports)
	total = n * (n - 1) // All permutations

	// Calculate domestic routes
	for _, count := range countryAirports {
		domestic += count * (count - 1)
	}

	international = total - domestic
	return
}
