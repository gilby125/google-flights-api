package api

import (
	"net/http"
	"strconv"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gin-gonic/gin"
)

// GetCheapestPath finds multi-hop routes between two airports
// GET /api/v1/graph/path?origin=ORD&dest=LHR&maxHops=2&maxPrice=1000
func GetCheapestPath(neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Query("origin")
		dest := c.Query("dest")
		if origin == "" || dest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "origin and dest query parameters are required"})
			return
		}

		maxHops := 2 // default
		if h := c.Query("maxHops"); h != "" {
			if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 5 {
				maxHops = parsed
			}
		}

		maxPrice := 10000.0 // default (effectively no limit)
		if p := c.Query("maxPrice"); p != "" {
			if parsed, err := strconv.ParseFloat(p, 64); err == nil && parsed > 0 {
				maxPrice = parsed
			}
		}

		paths, err := neo4jDB.FindCheapestPath(c.Request.Context(), origin, dest, maxHops, maxPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if paths == nil {
			paths = []db.PathResult{}
		}

		c.JSON(http.StatusOK, gin.H{
			"origin":   origin,
			"dest":     dest,
			"maxHops":  maxHops,
			"maxPrice": maxPrice,
			"paths":    paths,
		})
	}
}

// GetConnections finds all reachable destinations from an origin
// GET /api/v1/graph/connections?origin=ORD&maxHops=2&maxPrice=500
func GetConnections(neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Query("origin")
		if origin == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "origin query parameter is required"})
			return
		}

		maxHops := 1 // default
		if h := c.Query("maxHops"); h != "" {
			if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 3 {
				maxHops = parsed
			}
		}

		maxPrice := 5000.0 // default
		if p := c.Query("maxPrice"); p != "" {
			if parsed, err := strconv.ParseFloat(p, 64); err == nil && parsed > 0 {
				maxPrice = parsed
			}
		}

		connections, err := neo4jDB.FindConnections(c.Request.Context(), origin, maxHops, maxPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if connections == nil {
			connections = []db.Connection{}
		}

		c.JSON(http.StatusOK, gin.H{
			"origin":      origin,
			"maxHops":     maxHops,
			"maxPrice":    maxPrice,
			"count":       len(connections),
			"connections": connections,
		})
	}
}

// GetRouteStats returns price statistics for a specific route
// GET /api/v1/graph/route-stats?origin=ORD&dest=LHR
func GetRouteStats(neo4jDB db.Neo4jDatabase) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Query("origin")
		dest := c.Query("dest")
		if origin == "" || dest == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "origin and dest query parameters are required"})
			return
		}

		stats, err := neo4jDB.GetRouteStats(c.Request.Context(), origin, dest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if stats == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no price data found for this route"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
