package flights

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const priceGraphDiagnosticsEnv = "PRICE_GRAPH_DIAGNOSTICS"

// ParseErrors tracks non-fatal parsing issues for diagnostics.
//
// NOTE: This is best-effort instrumentation; parsing may drift as Google changes
// response formats.
type ParseErrors struct {
	UnmarshalFailures int `json:"unmarshal_failures"`
	DateParseFailures int `json:"date_parse_failures"`
	ZeroPriceCount    int `json:"zero_price_count"`
	TotalOffersRaw    int `json:"total_offers_raw"`
	EmptySections     int `json:"empty_sections"`
	// Samples contain redacted diagnostics (fingerprints + context), never raw payloads.
	Samples []string `json:"samples,omitempty"`
}

func (s *Session) getPriceGraphRawData(ctx context.Context, args PriceGraphArgs) (string, error) {
	return s.getRawData(ctx, args.Convert())
}

func (s *Session) getPriceGraphReqData(ctx context.Context, args PriceGraphArgs) (string, error) {
	serializedRangeStartDate := args.RangeStartDate.Format("2006-01-02")
	serializedRangeEndDate := args.RangeEndDate.Format("2006-01-02")

	rawData, err := s.getPriceGraphRawData(ctx, args)
	if err != nil {
		return "", err
	}

	prefix := `[null,"[null,`
	suffix := fmt.Sprintf(`],null,null,null,1,null,null,null,null,null,[]],[\"%s\",\"%s\"],null,[%d,%d]]"]`,
		serializedRangeStartDate, serializedRangeEndDate, args.TripLength, args.TripLength)

	reqData := prefix
	reqData += rawData
	reqData += suffix

	return url.QueryEscape(reqData), nil
}

func (s *Session) doRequestPriceGraph(ctx context.Context, args PriceGraphArgs) (*http.Response, error) {
	url := "https://www.google.com/_/FlightsFrontendUi/data/travel.frontend.flights.FlightsFrontendService/GetCalendarGraph?f.sid=-8920707734915550076&bl=boq_travel-frontend-ui_20230627.07_p1&hl=en&soc-app=162&soc-platform=1&soc-device=1&_reqid=261464&rt=c"

	reqDate, err := s.getPriceGraphReqData(ctx, args)
	if err != nil {
		return nil, err
	}

	jsonBody := []byte(
		`f.req=` + reqDate +
			`&at=AAuQa1oq5qIkgkQ2nG9vQZFTgSME%3A` + strconv.FormatInt(time.Now().Unix(), 10) + `&`)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", `*/*`)
	req.Header.Set("accept-language", `en-US,en;q=0.9`)
	req.Header.Set("cache-control", `no-cache`)
	req.Header.Set("content-type", `application/x-www-form-urlencoded;charset=UTF-8`)
	req.Header["cookie"] = s.cookies
	req.Header.Set("pragma", `no-cache`)
	req.Header.Set("user-agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`)
	req.Header.Set("x-goog-ext-259736195-jspb",
		fmt.Sprintf(`["en-US","US","%s",1,null,[-120],null,[[48764689,47907128,48676280,48710756,48627726,48480739,48593234,48707380]],1,[]]`, args.Currency))

	return s.client.Do(req)
}

func priceGraphSchema(startDate, returnDate *string, price *float64) *[]interface{} {
	// [startDate,returnDate,[[null,price],""],1]
	return &[]interface{}{startDate, returnDate, &[]interface{}{&[]interface{}{nil, price}}}
}

func truncateSample(input string, max int) string {
	if max <= 0 {
		return ""
	}
	input = strings.TrimSpace(input)
	if len(input) <= max {
		return input
	}
	return input[:max] + "…"
}

func priceGraphDiagnosticsEnabled() bool {
	raw, ok := os.LookupEnv(priceGraphDiagnosticsEnv)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func fingerprintBytes(input []byte) string {
	sum := sha256.Sum256(input)
	// Short fingerprint for logs; full SHA-256 is overkill.
	return hex.EncodeToString(sum[:8])
}

func sampleFingerprint(prefix string, input []byte) string {
	if len(input) == 0 {
		return fmt.Sprintf("%s sha256=<empty> len=0", prefix)
	}
	return fmt.Sprintf("%s sha256=%s len=%d", prefix, fingerprintBytes(input), len(input))
}

func extractRawPriceValue(raw json.RawMessage) (value any, ok bool) {
	var row []any
	if err := json.Unmarshal(raw, &row); err != nil {
		return nil, false
	}
	if len(row) < 3 {
		return nil, false
	}

	pricing, ok := row[2].([]any)
	if !ok || len(pricing) < 1 {
		return nil, false
	}

	firstCell, ok := pricing[0].([]any)
	if !ok || len(firstCell) < 2 {
		return nil, false
	}

	return firstCell[1], true
}

func mergeParseErrors(dst, src *ParseErrors) {
	if dst == nil || src == nil {
		return
	}
	dst.UnmarshalFailures += src.UnmarshalFailures
	dst.DateParseFailures += src.DateParseFailures
	dst.ZeroPriceCount += src.ZeroPriceCount
	dst.TotalOffersRaw += src.TotalOffersRaw
	dst.EmptySections += src.EmptySections

	if len(dst.Samples) >= 5 {
		return
	}
	for _, sample := range src.Samples {
		if sample == "" {
			continue
		}
		dst.Samples = append(dst.Samples, sample)
		if len(dst.Samples) >= 5 {
			return
		}
	}
}

func getPriceGraphSection(sectionIndex int, bytesToDecode []byte) ([]Offer, *ParseErrors) {
	offers := []Offer{}
	parseErrors := &ParseErrors{Samples: make([]string, 0, 5)}

	var err error

	rawOffers := []json.RawMessage{}

	if err = json.Unmarshal([]byte(bytesToDecode), &[]interface{}{nil, &rawOffers}); err != nil {
		parseErrors.UnmarshalFailures++
		if len(parseErrors.Samples) < 5 {
			parseErrors.Samples = append(parseErrors.Samples, sampleFingerprint("raw_offers_unmarshal", bytesToDecode))
		}
		if priceGraphDiagnosticsEnabled() && len(bytesToDecode) > 0 {
			slog.Info("price graph: failed to unmarshal raw offers array",
				"section", sectionIndex,
				"error", err,
				"bytes_length", len(bytesToDecode),
				"sha256", fingerprintBytes(bytesToDecode),
			)
		}
		return nil, parseErrors
	}

	parseErrors.TotalOffersRaw = len(rawOffers)
	zeroPriceLogged := 0

	for i, o := range rawOffers {
		finalOffer := Offer{}

		startDate := ""
		returnDate := ""

		if err = json.Unmarshal(o, priceGraphSchema(&startDate, &returnDate, &finalOffer.Price)); err != nil {
			parseErrors.UnmarshalFailures++
			if len(parseErrors.Samples) < 5 {
				parseErrors.Samples = append(parseErrors.Samples, sampleFingerprint(fmt.Sprintf("offer_unmarshal[%d]", i), o))
			}
			continue
		}

		if finalOffer.StartDate, err = time.Parse("2006-01-02", startDate); err != nil {
			parseErrors.DateParseFailures++
			if len(parseErrors.Samples) < 5 {
				parseErrors.Samples = append(parseErrors.Samples, sampleFingerprint(fmt.Sprintf("start_date_parse[%d]", i), o))
			}
			continue
		}
		if finalOffer.ReturnDate, err = time.Parse("2006-01-02", returnDate); err != nil {
			parseErrors.DateParseFailures++
			if len(parseErrors.Samples) < 5 {
				parseErrors.Samples = append(parseErrors.Samples, sampleFingerprint(fmt.Sprintf("return_date_parse[%d]", i), o))
			}
			continue
		}

		if finalOffer.Price == 0 {
			parseErrors.ZeroPriceCount++
			if zeroPriceLogged < 10 {
				zeroPriceLogged++
				if priceGraphDiagnosticsEnabled() {
					rawValue, ok := extractRawPriceValue(o)
					rawType := "<unknown>"
					rawNil := false
					rawFloat := false
					if ok {
						if rawValue == nil {
							rawNil = true
							rawType = "null"
						} else {
							rawType = fmt.Sprintf("%T", rawValue)
							if _, isFloat := rawValue.(float64); isFloat {
								rawFloat = true
							}
						}
					}
					slog.Info("price graph: zero-price offer parsed",
						"section", sectionIndex,
						"index", i,
						"start_date", startDate,
						"return_date", returnDate,
						"sha256", fingerprintBytes(o),
						"len", len(o),
						"raw_price_extracted", ok,
						"raw_price_type", rawType,
						"raw_price_is_null", rawNil,
						"raw_price_is_float", rawFloat,
					)
				}
			}
		}

		offers = append(offers, finalOffer)
	}

	if priceGraphDiagnosticsEnabled() && (parseErrors.UnmarshalFailures > 0 || parseErrors.DateParseFailures > 0 || parseErrors.ZeroPriceCount > 0) {
		slog.Info("price graph: section parse summary",
			"section", sectionIndex,
			"raw_offers", parseErrors.TotalOffersRaw,
			"unmarshal_failures", parseErrors.UnmarshalFailures,
			"date_parse_failures", parseErrors.DateParseFailures,
			"zero_price_count", parseErrors.ZeroPriceCount,
			"samples", parseErrors.Samples,
		)
	}

	return offers, parseErrors
}

// GetPriceGraph retrieves offers (date range) from the "Price graph" section of Google Flight search.
// The city names should be provided in the language described by args.Lang. The offers are returned
// in a slice of [Offer].
//
// GetPriceGraph returns an error if any of the requests fail or if any of the city names are misspelled.
//
// Requirements are described by the [PriceGraphArgs.Validate] function.
func (s *Session) GetPriceGraph(ctx context.Context, args PriceGraphArgs) ([]Offer, *ParseErrors, error) {
	if err := args.Validate(); err != nil {
		return nil, nil, err
	}

	offers := []Offer{}
	parseErrors := &ParseErrors{Samples: make([]string, 0, 5)}

	resp, err := s.doRequestPriceGraph(ctx, args)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body := bufio.NewReader(resp.Body)
	skipPrefix(body)

	sectionIndex := 0

	for {
		readLine(body) // skip line
		bytesToDecode, err := getInnerBytes(body)
		if err != nil {
			sortSlice(offers, func(lv, rv Offer) bool {
				return lv.StartDate.Before(rv.StartDate)
			})
			return offers, parseErrors, nil
		}

		sectionIndex++
		// The upstream response sometimes includes empty frames near the end; treat them as benign.
		if len(bytes.TrimSpace(bytesToDecode)) == 0 {
			parseErrors.EmptySections++
			continue
		}
		offers_, sectionErrors := getPriceGraphSection(sectionIndex, bytesToDecode)
		mergeParseErrors(parseErrors, sectionErrors)
		offers = append(offers, offers_...)
	}
}
