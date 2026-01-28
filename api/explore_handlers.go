package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gin-gonic/gin"
)

type ExploreEdge struct {
	OriginCode string  `json:"origin_code"`
	OriginLat  float64 `json:"origin_lat"`
	OriginLon  float64 `json:"origin_lon"`

	DestCode    string  `json:"dest_code"`
	DestName    string  `json:"dest_name,omitempty"`
	DestCity    string  `json:"dest_city,omitempty"`
	DestCountry string  `json:"dest_country,omitempty"`
	DestLat     float64 `json:"dest_lat"`
	DestLon     float64 `json:"dest_lon"`

	CheapestPrice float64 `json:"cheapest_price"`
	Hops          int     `json:"hops"`
}

type ExploreResponse struct {
	Origin   string   `json:"origin"`
	Origins  []string `json:"origins,omitempty"`
	MaxHops  int      `json:"maxHops"`
	MaxPrice float64  `json:"maxPrice"`
	DateFrom string   `json:"dateFrom,omitempty"`
	DateTo   string   `json:"dateTo,omitempty"`
	Limit    int      `json:"limit"`
	Source   string   `json:"source"`

	Count int           `json:"count"`
	Edges []ExploreEdge `json:"edges"`
}

// GetTopAirports returns a small curated list of airports for UI pickers.
// GET /api/v1/airports/top
func GetTopAirports() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, db.Top100Airports)
	}
}

// GetExplore aggregates route data for map/globe UIs.
// GET /api/v1/graph/explore?origin=ORD&maxHops=2&maxPrice=500&dateFrom=2026-01-01&dateTo=2026-12-31&airlines=AA,DL&limit=500
func GetExplore(neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		if neo4jDB == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "neo4j is not configured"})
			return
		}

		origin := strings.ToUpper(strings.TrimSpace(c.Query("origin")))
		originsRaw := strings.TrimSpace(c.Query("origins"))

		origins := []string{}
		seenOrigins := make(map[string]struct{})
		addOrigin := func(code string) {
			code = strings.ToUpper(strings.TrimSpace(code))
			if len(code) != 3 {
				return
			}
			if _, ok := seenOrigins[code]; ok {
				return
			}
			seenOrigins[code] = struct{}{}
			origins = append(origins, code)
		}

		if originsRaw != "" {
			for _, part := range strings.Split(originsRaw, ",") {
				addOrigin(part)
			}
		}
		if origin != "" {
			addOrigin(origin)
		}
		if len(origins) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "origin (or origins) query parameter is required"})
			return
		}

		maxHops := 1
		if h := strings.TrimSpace(c.Query("maxHops")); h != "" {
			if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 3 {
				maxHops = parsed
			}
		}

		maxPrice := 1000.0
		if p := strings.TrimSpace(c.Query("maxPrice")); p != "" {
			if parsed, err := strconv.ParseFloat(p, 64); err == nil && parsed > 0 {
				maxPrice = parsed
			}
		}

		dateFrom := strings.TrimSpace(c.Query("dateFrom"))
		dateTo := strings.TrimSpace(c.Query("dateTo"))

		source := strings.ToLower(strings.TrimSpace(c.Query("source")))
		if source == "" {
			source = "price_point"
		}
		if source != "price_point" && source != "route" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source must be one of: price_point, route"})
			return
		}
		if source == "route" && (dateFrom != "" || dateTo != "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dateFrom/dateTo are only supported with source=price_point"})
			return
		}

		limit := 500
		if l := strings.TrimSpace(c.Query("limit")); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 5000 {
				limit = parsed
			}
		}

		var airlines []string
		if raw := strings.TrimSpace(c.Query("airlines")); raw != "" {
			for _, part := range strings.Split(raw, ",") {
				code := strings.ToUpper(strings.TrimSpace(part))
				if code == "" {
					continue
				}
				if len(code) > 8 {
					continue
				}
				airlines = append(airlines, code)
			}
		}

		query := ""
		switch source {
		case "route":
			query = fmt.Sprintf(`
				MATCH path = (a:Airport)-[:ROUTE*1..%d]->(b:Airport)
				WHERE a <> b
				  AND a.code IN $origins
				  AND all(r IN relationships(path)
					WHERE r.avgPrice IS NOT NULL AND toFloat(r.avgPrice) > 0
					  AND (size($airlines) = 0 OR r.airline IN $airlines)
				  )
				WITH a, b,
				     reduce(total = 0.0, r IN relationships(path) | total + toFloat(r.avgPrice)) AS totalPrice,
				     length(path) AS hops
				WHERE totalPrice <= $maxPrice
				RETURN
					a.code AS originCode,
					coalesce(a.latitude, 0.0) AS originLat,
					coalesce(a.longitude, 0.0) AS originLon,
					b.code AS destCode,
					coalesce(b.name, '') AS destName,
					coalesce(b.city, '') AS destCity,
					coalesce(b.country, '') AS destCountry,
					coalesce(b.latitude, 0.0) AS destLat,
					coalesce(b.longitude, 0.0) AS destLon,
					min(totalPrice) AS cheapestPrice,
					min(hops) AS hops
				ORDER BY cheapestPrice
				LIMIT $limit
			`, maxHops)
		default: // price_point
			query = fmt.Sprintf(`
				MATCH path = (a:Airport)-[:PRICE_POINT*1..%d]->(b:Airport)
				WHERE a <> b
				  AND a.code IN $origins
				  AND all(r IN relationships(path)
					WHERE r.price IS NOT NULL AND toFloat(r.price) > 0
					  AND ($dateFrom = '' OR r.date >= date($dateFrom))
					  AND ($dateTo = '' OR r.date <= date($dateTo))
					  AND (size($airlines) = 0 OR r.airline IN $airlines)
				  )
				WITH a, b,
				     reduce(total = 0.0, r IN relationships(path) | total + toFloat(r.price)) AS totalPrice,
				     length(path) AS hops
				WHERE totalPrice <= $maxPrice
				RETURN
					a.code AS originCode,
					coalesce(a.latitude, 0.0) AS originLat,
					coalesce(a.longitude, 0.0) AS originLon,
					b.code AS destCode,
					coalesce(b.name, '') AS destName,
					coalesce(b.city, '') AS destCity,
					coalesce(b.country, '') AS destCountry,
					coalesce(b.latitude, 0.0) AS destLat,
					coalesce(b.longitude, 0.0) AS destLon,
					min(totalPrice) AS cheapestPrice,
					min(hops) AS hops
				ORDER BY cheapestPrice
				LIMIT $limit
			`, maxHops)
		}

		result, err := neo4jDB.ExecuteReadQuery(c.Request.Context(), query, map[string]interface{}{
			"origins":  origins,
			"maxPrice": maxPrice,
			"dateFrom": dateFrom,
			"dateTo":   dateTo,
			"airlines": airlines,
			"limit":    limit,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer result.Close()

		edges := make([]ExploreEdge, 0, 128)
		for result.Next() {
			rec := result.Record()
			if rec == nil {
				continue
			}

			edge := ExploreEdge{}

			if v, ok := rec.Get("originCode"); ok {
				edge.OriginCode, _ = v.(string)
			}
			if v, ok := rec.Get("originLat"); ok {
				edge.OriginLat, _ = v.(float64)
			}
			if v, ok := rec.Get("originLon"); ok {
				edge.OriginLon, _ = v.(float64)
			}

			if v, ok := rec.Get("destCode"); ok {
				edge.DestCode, _ = v.(string)
			}
			if v, ok := rec.Get("destName"); ok {
				edge.DestName, _ = v.(string)
			}
			if v, ok := rec.Get("destCity"); ok {
				edge.DestCity, _ = v.(string)
			}
			if v, ok := rec.Get("destCountry"); ok {
				edge.DestCountry, _ = v.(string)
			}
			if v, ok := rec.Get("destLat"); ok {
				edge.DestLat, _ = v.(float64)
			}
			if v, ok := rec.Get("destLon"); ok {
				edge.DestLon, _ = v.(float64)
			}

			if v, ok := rec.Get("cheapestPrice"); ok {
				switch vv := v.(type) {
				case float64:
					edge.CheapestPrice = vv
				case int64:
					edge.CheapestPrice = float64(vv)
				}
			}
			if v, ok := rec.Get("hops"); ok {
				if vv, ok := v.(int64); ok {
					edge.Hops = int(vv)
				}
			}

			edges = append(edges, edge)
		}
		if err := result.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ExploreResponse{
			Origin:   origin,
			Origins:  origins,
			MaxHops:  maxHops,
			MaxPrice: maxPrice,
			DateFrom: dateFrom,
			DateTo:   dateTo,
			Limit:    limit,
			Source:   source,
			Count:    len(edges),
			Edges:    edges,
		})
	}
}
