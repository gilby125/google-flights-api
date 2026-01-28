package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gin-gonic/gin"
)

func GetRouteDetails(neo4jDB db.Neo4jDatabase, defaultExcludeAirlines []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if neo4jDB == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "neo4j is not configured"})
			return
		}

		origin := strings.ToUpper(strings.TrimSpace(c.Query("origin")))
		dest := strings.ToUpper(strings.TrimSpace(c.Query("dest")))
		if origin == "" || dest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "origin and dest query parameters are required"})
			return
		}

		dateFrom := strings.TrimSpace(c.Query("dateFrom"))
		dateTo := strings.TrimSpace(c.Query("dateTo"))
		tripType := strings.ToLower(strings.TrimSpace(c.Query("tripType")))
		if tripType != "" && tripType != "one_way" && tripType != "round_trip" && tripType != "unknown" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tripType must be one of: one_way, round_trip, unknown"})
			return
		}

		maxAgeDays := 0
		if v := strings.TrimSpace(c.Query("maxAgeDays")); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 && parsed <= 3650 {
				maxAgeDays = parsed
			}
		}

		limitSamples := 20
		if v := strings.TrimSpace(c.Query("limitSamples")); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 && parsed <= 200 {
				limitSamples = parsed
			}
		}

		parseCodes := func(raw string) []string {
			out := []string{}
			seen := make(map[string]struct{})
			for _, part := range strings.Split(raw, ",") {
				code := strings.ToUpper(strings.TrimSpace(part))
				if code == "" {
					continue
				}
				if len(code) > 8 {
					continue
				}
				if _, ok := seen[code]; ok {
					continue
				}
				seen[code] = struct{}{}
				out = append(out, code)
			}
			return out
		}

		airlines := parseCodes(strings.TrimSpace(c.Query("airlines")))
		rawExclude, hasExclude := c.GetQuery("excludeAirlines")
		excludeAirlines := parseCodes(strings.TrimSpace(rawExclude))
		if !hasExclude && len(excludeAirlines) == 0 && len(defaultExcludeAirlines) > 0 {
			excludeAirlines = append(excludeAirlines, defaultExcludeAirlines...)
		}

		query := `
			MATCH (a:Airport {code: $origin}), (b:Airport {code: $dest})
			CALL {
				WITH a, b
				OPTIONAL MATCH (a)-[r:PRICE_POINT]->(b)
				WITH r,
				     (r IS NOT NULL
				      AND r.price IS NOT NULL AND toFloat(r.price) > 0
				      AND r.date IS NOT NULL
				      AND ($dateFrom = '' OR r.date >= date($dateFrom))
				      AND ($dateTo = '' OR r.date <= date($dateTo))
				      AND (size($airlines) = 0 OR r.airline IN $airlines)
				      AND (size($excludeAirlines) = 0 OR r.airline IS NULL OR NOT r.airline IN $excludeAirlines)
				      AND ($maxAgeDays = 0 OR coalesce(r.last_seen_at, r.first_seen_at) IS NULL OR coalesce(r.last_seen_at, r.first_seen_at) >= datetime() - duration({days: $maxAgeDays}))
				      AND ($tripType = '' OR coalesce(r.trip_type, 'unknown') = $tripType)
				     ) AS ok
				WITH
					CASE WHEN ok THEN toFloat(r.price) ELSE null END AS p,
					CASE WHEN ok THEN r.airline ELSE null END AS airline,
					CASE WHEN ok THEN coalesce(r.first_seen_at, r.last_seen_at) ELSE null END AS firstSeen,
					CASE WHEN ok THEN coalesce(r.last_seen_at, r.first_seen_at) ELSE null END AS lastSeen
				RETURN
					min(p) AS minPrice,
					max(p) AS maxPrice,
					avg(p) AS avgPrice,
					count(p) AS pricePointCount,
					[x IN collect(DISTINCT airline) WHERE x IS NOT NULL AND x <> ''] AS airlines,
					toString(min(firstSeen)) AS firstSeenAt,
					toString(max(lastSeen)) AS lastSeenAt
			}
			CALL {
				WITH a, b
				OPTIONAL MATCH (a)-[r:PRICE_POINT]->(b)
				WITH r,
				     (r IS NOT NULL
				      AND r.price IS NOT NULL AND toFloat(r.price) > 0
				      AND r.date IS NOT NULL
				      AND ($dateFrom = '' OR r.date >= date($dateFrom))
				      AND ($dateTo = '' OR r.date <= date($dateTo))
				      AND (size($airlines) = 0 OR r.airline IN $airlines)
				      AND (size($excludeAirlines) = 0 OR r.airline IS NULL OR NOT r.airline IN $excludeAirlines)
				      AND ($maxAgeDays = 0 OR coalesce(r.last_seen_at, r.first_seen_at) IS NULL OR coalesce(r.last_seen_at, r.first_seen_at) >= datetime() - duration({days: $maxAgeDays}))
				      AND ($tripType = '' OR coalesce(r.trip_type, 'unknown') = $tripType)
				     ) AS ok
				WITH r, ok
				ORDER BY CASE WHEN ok THEN 0 ELSE 1 END, toFloat(coalesce(r.price, 1e15)) ASC, r.date ASC
				WITH collect(CASE WHEN ok THEN r ELSE null END) AS rs
				RETURN head([x IN rs WHERE x IS NOT NULL]) AS minR
			}
			CALL {
				WITH a, b
				OPTIONAL MATCH (a)-[r:PRICE_POINT]->(b)
				WITH r,
				     (r IS NOT NULL
				      AND r.price IS NOT NULL AND toFloat(r.price) > 0
				      AND r.date IS NOT NULL
				      AND ($dateFrom = '' OR r.date >= date($dateFrom))
				      AND ($dateTo = '' OR r.date <= date($dateTo))
				      AND (size($airlines) = 0 OR r.airline IN $airlines)
				      AND (size($excludeAirlines) = 0 OR r.airline IS NULL OR NOT r.airline IN $excludeAirlines)
				      AND ($maxAgeDays = 0 OR coalesce(r.last_seen_at, r.first_seen_at) IS NULL OR coalesce(r.last_seen_at, r.first_seen_at) >= datetime() - duration({days: $maxAgeDays}))
				      AND ($tripType = '' OR coalesce(r.trip_type, 'unknown') = $tripType)
				     ) AS ok
				WITH r, ok
				ORDER BY CASE WHEN ok THEN 0 ELSE 1 END, toFloat(coalesce(r.price, -1)) DESC, r.date DESC
				WITH collect(CASE WHEN ok THEN r ELSE null END) AS rs
				RETURN head([x IN rs WHERE x IS NOT NULL]) AS maxR
			}
			CALL {
				WITH a, b
				OPTIONAL MATCH (a)-[r:PRICE_POINT]->(b)
				WITH r,
				     (r IS NOT NULL
				      AND r.price IS NOT NULL AND toFloat(r.price) > 0
				      AND r.date IS NOT NULL
				      AND ($dateFrom = '' OR r.date >= date($dateFrom))
				      AND ($dateTo = '' OR r.date <= date($dateTo))
				      AND (size($airlines) = 0 OR r.airline IN $airlines)
				      AND (size($excludeAirlines) = 0 OR r.airline IS NULL OR NOT r.airline IN $excludeAirlines)
				      AND ($maxAgeDays = 0 OR coalesce(r.last_seen_at, r.first_seen_at) IS NULL OR coalesce(r.last_seen_at, r.first_seen_at) >= datetime() - duration({days: $maxAgeDays}))
				      AND ($tripType = '' OR coalesce(r.trip_type, 'unknown') = $tripType)
				     ) AS ok
				WITH r, ok
				ORDER BY coalesce(r.last_seen_at, r.first_seen_at) DESC, r.date DESC
				WITH collect(CASE
					WHEN ok THEN {
						date: toString(r.date),
						price: toFloat(r.price),
						airline: coalesce(r.airline, ''),
						seen_at: toString(coalesce(r.last_seen_at, r.first_seen_at)),
						trip_type: coalesce(r.trip_type, 'unknown'),
						return_date: toString(r.return_date)
					}
					ELSE null
				END) AS raw
				RETURN [x IN raw WHERE x IS NOT NULL][0..$limitSamples] AS samples
			}
			RETURN
				$origin AS origin,
				$dest AS destination,
				coalesce(minPrice, 0.0) AS minPrice,
				coalesce(maxPrice, 0.0) AS maxPrice,
				coalesce(avgPrice, 0.0) AS avgPrice,
				toInteger(coalesce(pricePointCount, 0)) AS pricePointCount,
				coalesce(airlines, []) AS airlines,
				coalesce(firstSeenAt, '') AS firstSeenAt,
				coalesce(lastSeenAt, '') AS lastSeenAt,
				coalesce(toString(minR.date), '') AS minPriceDate,
				coalesce(minR.airline, '') AS minPriceAirline,
				coalesce(toString(coalesce(minR.last_seen_at, minR.first_seen_at)), '') AS minPriceSeenAt,
				coalesce(coalesce(minR.trip_type, 'unknown'), 'unknown') AS minPriceTripType,
				coalesce(toString(minR.return_date), '') AS minPriceReturnDate,
				coalesce(toString(maxR.date), '') AS maxPriceDate,
				coalesce(maxR.airline, '') AS maxPriceAirline,
				coalesce(toString(coalesce(maxR.last_seen_at, maxR.first_seen_at)), '') AS maxPriceSeenAt,
				coalesce(coalesce(maxR.trip_type, 'unknown'), 'unknown') AS maxPriceTripType,
				coalesce(toString(maxR.return_date), '') AS maxPriceReturnDate,
				coalesce(samples, []) AS samples
		`

		result, err := neo4jDB.ExecuteReadQuery(c.Request.Context(), query, map[string]interface{}{
			"origin":          origin,
			"dest":            dest,
			"dateFrom":        dateFrom,
			"dateTo":          dateTo,
			"airlines":        airlines,
			"excludeAirlines": excludeAirlines,
			"maxAgeDays":      maxAgeDays,
			"tripType":        tripType,
			"limitSamples":    limitSamples,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer result.Close()

		if !result.Next() {
			c.JSON(http.StatusNotFound, gin.H{"error": "no price data found for this route"})
			return
		}

		rec := result.Record()
		if rec == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no price data found for this route"})
			return
		}

		stats := &db.RouteStats{}
		if v, ok := rec.Get("origin"); ok {
			stats.Origin, _ = v.(string)
		}
		if v, ok := rec.Get("destination"); ok {
			stats.Destination, _ = v.(string)
		}
		if v, ok := rec.Get("minPrice"); ok {
			switch vv := v.(type) {
			case float64:
				stats.MinPrice = vv
			case int64:
				stats.MinPrice = float64(vv)
			}
		}
		if v, ok := rec.Get("maxPrice"); ok {
			switch vv := v.(type) {
			case float64:
				stats.MaxPrice = vv
			case int64:
				stats.MaxPrice = float64(vv)
			}
		}
		if v, ok := rec.Get("avgPrice"); ok {
			switch vv := v.(type) {
			case float64:
				stats.AvgPrice = vv
			case int64:
				stats.AvgPrice = float64(vv)
			}
		}
		if v, ok := rec.Get("pricePointCount"); ok {
			if i, ok := v.(int64); ok {
				stats.PricePoints = int(i)
			}
		}
		if v, ok := rec.Get("airlines"); ok {
			if arr, ok := v.([]interface{}); ok {
				for _, it := range arr {
					if s, ok := it.(string); ok && s != "" {
						stats.Airlines = append(stats.Airlines, s)
					}
				}
			}
		}
		if v, ok := rec.Get("firstSeenAt"); ok {
			stats.FirstSeenAt, _ = v.(string)
		}
		if v, ok := rec.Get("lastSeenAt"); ok {
			stats.LastSeenAt, _ = v.(string)
		}
		if v, ok := rec.Get("minPriceDate"); ok {
			stats.MinPriceDate, _ = v.(string)
		}
		if v, ok := rec.Get("minPriceAirline"); ok {
			stats.MinPriceAirline, _ = v.(string)
		}
		if v, ok := rec.Get("minPriceSeenAt"); ok {
			stats.MinPriceSeenAt, _ = v.(string)
		}
		if v, ok := rec.Get("minPriceTripType"); ok {
			stats.MinPriceTripType, _ = v.(string)
		}
		if v, ok := rec.Get("minPriceReturnDate"); ok {
			stats.MinPriceReturnDate, _ = v.(string)
		}
		if v, ok := rec.Get("maxPriceDate"); ok {
			stats.MaxPriceDate, _ = v.(string)
		}
		if v, ok := rec.Get("maxPriceAirline"); ok {
			stats.MaxPriceAirline, _ = v.(string)
		}
		if v, ok := rec.Get("maxPriceSeenAt"); ok {
			stats.MaxPriceSeenAt, _ = v.(string)
		}
		if v, ok := rec.Get("maxPriceTripType"); ok {
			stats.MaxPriceTripType, _ = v.(string)
		}
		if v, ok := rec.Get("maxPriceReturnDate"); ok {
			stats.MaxPriceReturnDate, _ = v.(string)
		}
		if v, ok := rec.Get("samples"); ok {
			if arr, ok := v.([]interface{}); ok {
				for _, it := range arr {
					m, ok := it.(map[string]interface{})
					if !ok {
						continue
					}
					sample := db.RoutePricePoint{}
					if vv, ok := m["date"].(string); ok {
						sample.Date = vv
					}
					if vv, ok := m["price"].(float64); ok {
						sample.Price = vv
					}
					if vv, ok := m["airline"].(string); ok {
						sample.Airline = vv
					}
					if vv, ok := m["seen_at"].(string); ok {
						sample.SeenAt = vv
					}
					if vv, ok := m["trip_type"].(string); ok {
						sample.TripType = vv
					}
					if vv, ok := m["return_date"].(string); ok {
						sample.ReturnDate = vv
					}
					stats.Samples = append(stats.Samples, sample)
				}
			}
		}

		if err := result.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if stats.PricePoints == 0 {
			fallbackQuery := `
				MATCH (a:Airport {code: $origin})-[r:ROUTE]->(b:Airport {code: $dest})
				WHERE r.avgPrice IS NOT NULL AND toFloat(r.avgPrice) > 0
				  AND (size($airlines) = 0 OR r.airline IN $airlines)
				  AND (size($excludeAirlines) = 0 OR r.airline IS NULL OR NOT r.airline IN $excludeAirlines)
				RETURN
					min(toFloat(r.avgPrice)) AS minPrice,
					max(toFloat(r.avgPrice)) AS maxPrice,
					avg(toFloat(r.avgPrice)) AS avgPrice,
					toInteger(count(r)) AS routeEdgeCount,
					[x IN collect(DISTINCT r.airline) WHERE x IS NOT NULL AND x <> ''] AS airlines
			`
			fallbackResult, fbErr := neo4jDB.ExecuteReadQuery(c.Request.Context(), fallbackQuery, map[string]interface{}{
				"origin":          origin,
				"dest":            dest,
				"airlines":        airlines,
				"excludeAirlines": excludeAirlines,
			})
			if fbErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fbErr.Error()})
				return
			}
			defer fallbackResult.Close()

			if fallbackResult.Next() && fallbackResult.Record() != nil {
				fb := fallbackResult.Record()
				routeEdgeCount := int64(0)
				if v, ok := fb.Get("routeEdgeCount"); ok {
					if i, ok := v.(int64); ok {
						routeEdgeCount = i
					}
				}
				if routeEdgeCount <= 0 {
					c.JSON(http.StatusNotFound, gin.H{"error": "no route data found for this origin/destination"})
					return
				}

				if v, ok := fb.Get("minPrice"); ok {
					switch vv := v.(type) {
					case float64:
						stats.MinPrice = vv
					case int64:
						stats.MinPrice = float64(vv)
					}
				}
				if v, ok := fb.Get("maxPrice"); ok {
					switch vv := v.(type) {
					case float64:
						stats.MaxPrice = vv
					case int64:
						stats.MaxPrice = float64(vv)
					}
				}
				if v, ok := fb.Get("avgPrice"); ok {
					switch vv := v.(type) {
					case float64:
						stats.AvgPrice = vv
					case int64:
						stats.AvgPrice = float64(vv)
					}
				}
				if v, ok := fb.Get("routeEdgeCount"); ok {
					if i, ok := v.(int64); ok {
						stats.RouteEdges = int(i)
					}
				}
				if v, ok := fb.Get("airlines"); ok {
					stats.Airlines = nil
					if arr, ok := v.([]interface{}); ok {
						for _, it := range arr {
							if s, ok := it.(string); ok && s != "" {
								stats.Airlines = append(stats.Airlines, s)
							}
						}
					}
				}
				stats.Source = "route"
				stats.Note = "No date-specific PRICE_POINT samples matched; showing ROUTE avgPrice aggregates instead."
				c.JSON(http.StatusOK, stats)
				return
			}

			if err := fallbackResult.Err(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusNotFound, gin.H{"error": "no route data found for this origin/destination"})
			return
		}

		stats.Source = "price_point"
		c.JSON(http.StatusOK, stats)
	}
}
