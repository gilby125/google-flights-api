package flights

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/text/language"
)

func getCityReqData(city string) string {
	return url.QueryEscape(fmt.Sprintf(`[[["H028ib","[\"%s\",[1,2,3,5,4],null,[1,1,1],1]",null,"generic"]]]`, city))
}

func (s *Session) doRequestLocation(city string, lang language.Tag) (*http.Response, error) {
	requestURL := "https://www.google.com/_/TravelFrontendUi/data/batchexecute?rpcids=H028ib&source-path=%2Ftravel%2Fflights%2Fsearch&f.sid=-8421128425468344897&bl=boq_travel-frontend-ui_20230613.06_p0" +
		"&hl=" + lang.String() +
		"&soc-app=162&soc-platform=1&soc-device=1&_reqid=444052&rt=c"

	jsonBody := []byte(
		`f.req=` + getCityReqData(city) +
			`&at=AAuQa1qJpLKW2Hl-i40OwJyzmo22%3A` + strconv.FormatInt(time.Now().Unix(), 10) + `&`)

	req, err := retryablehttp.NewRequest(http.MethodPost, requestURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header["cookie"] = s.cookies
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36")

	return s.client.Do(req)
}

func abbrCitySchema(city, abbrCity *string) *[][][][]interface{} {
	// [[[[3,"",city,"",abbrCity,null,null,null,null,null,null,3],...]]]
	return &[][][][]interface{}{{{{nil, nil, city, nil, abbrCity}}}}
}

// AbbrCity serializes the city name by requesting it from the Google Flights API. The city name should
// be provided in the language described by [language.Tag].
//
// AbbrCity returns an error if the city name is misspelled or the Google Flights API returns an unexpected response.
func (s *Session) AbbrCity(city string, lang language.Tag) (string, error) {
	if abbrCity, ok := s.Cities.Load(city); ok {
		return abbrCity, nil
	}

	resp, err := s.doRequestLocation(city, lang)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body := bufio.NewReader(resp.Body)
	skipPrefix(body)
	readLine(body) // skip line

	bytesToDecode, err := getInnerBytes(body)
	if err != nil {
		return "", err
	}

	var receivedCity string
	var abbrCity string

	err = json.Unmarshal(bytesToDecode, abbrCitySchema(&receivedCity, &abbrCity))
	if err != nil {
		return "", fmt.Errorf("AbbrCity error during decoding: %v", err)
	}

	if city != receivedCity {
		return "", fmt.Errorf("the requested city name didn't match the found. requested: %s found: %s", city, receivedCity)
	}

	s.Cities.Store(receivedCity, abbrCity)

	return abbrCity, nil
}

func (s *Session) abbrCities(cities []string, lang language.Tag) ([]string, error) {
	abbrCities := []string{}
	for _, c := range cities {
		sc, err := s.AbbrCity(c, lang)
		if err != nil {
			return nil, err
		}
		abbrCities = append(abbrCities, sc)
	}
	return abbrCities, nil
}

func iataCodeSchema(iataCode *string) *[][][][]interface{} {
	// [[[[3,"",city,"",abbrCity,iataCode,null,null,null,null,null,3],...]]]
	return &[][][][]interface{}{{{{nil, nil, nil, nil, nil, iataCode}}}}
}

// IsIATASupported checks whether the provided IATA code is supported by the Google Flights API.
//
// IsIATASupported returns an error if the Google Flights API returns an unexpected response.
func (s *Session) IsIATASupported(iataCode string) (bool, error) {
	resp, err := s.doRequestLocation(iataCode, language.English)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body := bufio.NewReader(resp.Body)
	skipPrefix(body)
	readLine(body) // skip line

	bytesToDecode, err := getInnerBytes(body)
	if err != nil {
		return false, err
	}

	var receivedIataCode string

	err = json.Unmarshal(bytesToDecode, iataCodeSchema(&receivedIataCode))
	if err != nil {
		return false, fmt.Errorf("IsIATASupported error during decoding: %v", err)
	}

	return iataCode == receivedIataCode, nil
}