package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chrisprojs/Franchiso/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// GoogleMapsPlace represents a place result from Google Maps API
type GoogleMapsPlace struct {
	PlaceID          string `json:"place_id"`
	Name             string `json:"name"`
	FormattedAddress string `json:"formatted_address"`
	Geometry         struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
	Rating               *float64 `json:"rating,omitempty"`
	FormattedPhoneNumber *string  `json:"formatted_phone_number,omitempty"`
	OpeningHours         *struct {
		OpenNow     bool     `json:"open_now"`
		WeekdayText []string `json:"weekday_text,omitempty"`
	} `json:"opening_hours,omitempty"`
}

// GoogleMapsPlacesResponse represents the response from Google Maps Places API
type GetFranchiseLocationsResponse struct {
	Results       []GoogleMapsPlace `json:"results"`
	Status        string            `json:"status"`
	NextPageToken *string           `json:"next_page_token,omitempty"`
}

// IndonesiaProvince represents a province from the GeoJSON file
type IndonesiaProvince struct {
	Type     string `json:"type"`
	Geometry struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		ID       int    `json:"ID"`
		Kode     int    `json:"kode"`
		Propinsi string `json:"Propinsi"`
		SUMBER   string `json:"SUMBER"`
	} `json:"properties"`
}

// IndonesiaProvincesCollection represents the complete GeoJSON collection
type IndonesiaProvincesCollection struct {
	Type     string              `json:"type"`
	Features []IndonesiaProvince `json:"features"`
}

// Global variable to store loaded provinces data
var provincesData *IndonesiaProvincesCollection

// Cache for province boundaries to avoid recalculation
var provinceBoundariesCache = make(map[string]struct {
	minLat, minLng, maxLat, maxLng float64
})

const franchiseLocationsCacheTTL = 7 * 24 * time.Hour

// LoadProvincesData loads the Indonesia provinces GeoJSON data
func LoadProvincesData() error {
	file, err := os.Open("indonesia-province-simple.json")
	if err != nil {
		return fmt.Errorf("failed to open provinces file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read provinces file: %v", err)
	}

	var collection IndonesiaProvincesCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse provinces JSON: %v", err)
	}

	provincesData = &collection
	return nil
}

// FindProvinceByName searches for a province by name (case-insensitive)
func FindProvinceByName(name string) *IndonesiaProvince {
	// Lazy loading: load provinces data if not already loaded
	if provincesData == nil {
		if err := LoadProvincesData(); err != nil {
			return nil
		}
	}

	nameLower := strings.ToLower(name)

	// Try exact province name matching first
	for _, province := range provincesData.Features {
		provinceNameLower := strings.ToLower(province.Properties.Propinsi)
		if strings.EqualFold(provinceNameLower, nameLower) {
			return &province
		}
	}

	// If exact match not found, try partial matching
	for _, province := range provincesData.Features {
		provinceNameLower := strings.ToLower(province.Properties.Propinsi)
		if strings.Contains(provinceNameLower, nameLower) || strings.Contains(nameLower, provinceNameLower) {
			return &province
		}
	}

	return nil
}

// PointInPolygon checks if a point is inside a polygon using ray casting algorithm
func PointInPolygon(lat, lng float64, coordinates [][][]float64) bool {
	inside := false
	for _, ring := range coordinates {
		if len(ring) < 3 {
			continue
		}

		j := len(ring) - 1
		for i := 0; i < len(ring); i++ {
			// Coordinates are [longitude, latitude] in GeoJSON
			xi, yi := ring[i][0], ring[i][1] // xi = longitude, yi = latitude
			xj, yj := ring[j][0], ring[j][1] // xj = longitude, yj = latitude

			// Ray casting algorithm: check if ray from point crosses polygon edge
			if ((yi > lat) != (yj > lat)) && (lng < (xj-xi)*(lat-yi)/(yj-yi)+xi) {
				inside = !inside
			}
			j = i
		}
	}
	return inside
}

// IsCoordinateInProvince checks if a coordinate is within a province's boundaries
func IsCoordinateInProvince(lat, lng float64, province *IndonesiaProvince) bool {
	if province == nil {
		return false
	}

	// Parse coordinates based on geometry type
	if province.Geometry.Type == "Polygon" {
		var coordinates [][][]float64
		if err := json.Unmarshal(province.Geometry.Coordinates, &coordinates); err != nil {
			return false
		}
		return PointInPolygon(lat, lng, coordinates)
	} else if province.Geometry.Type == "MultiPolygon" {
		var coordinates [][][][]float64
		if err := json.Unmarshal(province.Geometry.Coordinates, &coordinates); err != nil {
			return false
		}

		// Check if point is in any of the polygons
		for _, polygon := range coordinates {
			if PointInPolygon(lat, lng, polygon) {
				return true
			}
		}
	}

	return false
}

// GetProvinceBoundaries returns the bounding box of a province with caching
func GetProvinceBoundaries(province *IndonesiaProvince) (minLat, minLng, maxLat, maxLng float64, err error) {
	if province == nil {
		return 0, 0, 0, 0, fmt.Errorf("province is nil")
	}

	// Check cache first
	provinceName := province.Properties.Propinsi
	if cached, exists := provinceBoundariesCache[provinceName]; exists {
		return cached.minLat, cached.minLng, cached.maxLat, cached.maxLng, nil
	}

	minLat, minLng = math.Inf(1), math.Inf(1)
	maxLat, maxLng = math.Inf(-1), math.Inf(-1)

	// Parse coordinates based on geometry type
	if province.Geometry.Type == "Polygon" {
		var coordinates [][][]float64
		if err := json.Unmarshal(province.Geometry.Coordinates, &coordinates); err != nil {
			return 0, 0, 0, 0, err
		}

		for _, ring := range coordinates {
			for _, point := range ring {
				if len(point) >= 2 {
					lng, lat := point[0], point[1]
					if lat < minLat {
						minLat = lat
					}
					if lat > maxLat {
						maxLat = lat
					}
					if lng < minLng {
						minLng = lng
					}
					if lng > maxLng {
						maxLng = lng
					}
				}
			}
		}
	} else if province.Geometry.Type == "MultiPolygon" {
		var coordinates [][][][]float64
		if err := json.Unmarshal(province.Geometry.Coordinates, &coordinates); err != nil {
			return 0, 0, 0, 0, err
		}

		for _, polygon := range coordinates {
			for _, ring := range polygon {
				for _, point := range ring {
					if len(point) >= 2 {
						lng, lat := point[0], point[1]
						if lat < minLat {
							minLat = lat
						}
						if lat > maxLat {
							maxLat = lat
						}
						if lng < minLng {
							minLng = lng
						}
						if lng > maxLng {
							maxLng = lng
						}
					}
				}
			}
		}
	}

	// Cache the result
	provinceBoundariesCache[provinceName] = struct {
		minLat, minLng, maxLat, maxLng float64
	}{minLat, minLng, maxLat, maxLng}

	return minLat, minLng, maxLat, maxLng, nil
}

// IsCoordinateInProvinceOptimized checks if a coordinate is within a province's boundaries with optimization
func IsCoordinateInProvinceOptimized(lat, lng float64, province *IndonesiaProvince, minLat, minLng, maxLat, maxLng float64) bool {
	if province == nil {
		return false
	}

	// Quick bounding box check first (much faster than point-in-polygon)
	if lat < minLat || lat > maxLat || lng < minLng || lng > maxLng {
		return false
	}

	// Only do expensive point-in-polygon calculation if within bounding box
	return IsCoordinateInProvince(lat, lng, province)
}

// executeQuery executes a single Google Maps API query with rate limiting
func executeQuery(query string, location string, radius int, apiKey, baseURL string) (*GetFranchiseLocationsResponse, error) {
	// Rate limiting: small delay to prevent throttling
	time.Sleep(100 * time.Millisecond)

	// Handle pagination token URLs (they already contain the full URL)
	var fullURL string
	if strings.Contains(baseURL, "pagetoken=") {
		fullURL = baseURL
	} else {
		params := url.Values{}
		if query != "" {
			params.Add("query", query)
		}
		if location != "" {
			params.Add("location", location)
		}
		if radius > 0 {
			params.Add("radius", strconv.Itoa(radius))
		}
		params.Add("key", apiKey)
		fullURL = fmt.Sprintf("%s%s", baseURL, params.Encode())
	}

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	var result GetFranchiseLocationsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFranchiseLocations fetches franchise locations by brand name using Google Maps Places API
func GetFranchiseLocations(c *gin.Context, app *config.App) {
	brandName := strings.TrimSpace(c.Query("brand"))
	province := strings.TrimSpace(c.Query("province"))

	// Debug logging
	fmt.Printf("DEBUG: brandName=%s, province=%s\n", brandName, province)
	fmt.Printf("DEBUG: Google Maps API Key: %s\n", app.GoogleMaps.ApiKey)
	fmt.Printf("DEBUG: Google Maps Base URL: %s\n", app.GoogleMaps.BaseURL)
	fmt.Printf("DEBUG: Starting GetFranchiseLocations function\n")

	// Validate input
	if brandName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Brand name is required",
		})
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("franchise_locations:%s:%s", strings.ToLower(brandName), strings.ToLower(province))

	if app.Redis != nil {
		cachedData, err := app.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedResponse GetFranchiseLocationsResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
				fmt.Printf("DEBUG: Serving GetFranchiseLocations from Redis cache for key=%s\n", cacheKey)
				c.JSON(http.StatusOK, cachedResponse)
				return
			}
			fmt.Printf("DEBUG: Failed to unmarshal cached response for key=%s: %v\n", cacheKey, err)
		} else if err != redis.Nil {
			fmt.Printf("DEBUG: Redis GET error for key=%s: %v\n", cacheKey, err)
		}
	}

	// If a province is provided, restrict results to within the province area and fetch all pages
	if province != "" {
		// Find province data and get its boundaries
		provinceData := FindProvinceByName(province)
		fmt.Printf("DEBUG: provinceData=%+v\n", provinceData)
		if provinceData == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Province not found"})
			return
		}

		// Get province boundaries
		minLat, minLng, maxLat, maxLng, err := GetProvinceBoundaries(provinceData)
		fmt.Printf("DEBUG: boundaries - minLat=%f, minLng=%f, maxLat=%f, maxLng=%f\n", minLat, minLng, maxLat, maxLng)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get province boundaries: %v", err),
			})
			return
		}

		// Calculate center point of province
		centerLat := (minLat + maxLat) / 2
		centerLng := (minLng + maxLng) / 2
		fmt.Printf("DEBUG: center - lat=%f, lng=%f\n", centerLat, centerLng)

		// Calculate radius based on province size
		latDiff := maxLat - minLat
		lngDiff := maxLng - minLng
		maxDiff := latDiff
		if lngDiff > latDiff {
			maxDiff = lngDiff
		}
		// Convert degrees to meters (approximate: 1 degree ≈ 111,320 meters)
		radiusMeters := int((maxDiff * 111320) / 2)
		if radiusMeters < 10000 {
			radiusMeters = 10000 // minimum 10km radius
		}
		if radiusMeters > 100000 {
			radiusMeters = 100000 // maximum 100km radius
		}
		fmt.Printf("DEBUG: radiusMeters=%d\n", radiusMeters)

		aggregated := GetFranchiseLocationsResponse{Results: []GoogleMapsPlace{}}
		existingPlaceIDs := make(map[string]bool)

		// Optimized Strategy: Use only the most effective queries
		searchQueries := []string{
			fmt.Sprintf("%s %s", brandName, province),
		}

		// Execute queries with early exit strategy
		for queryIndex, query := range searchQueries {
			// Early exit if we have enough results
			if len(aggregated.Results) >= 50 {
				fmt.Printf("DEBUG: Early exit - sufficient results found (%d)\n", len(aggregated.Results))
				break
			}

			fmt.Printf("DEBUG: Optimized query %d: %s\n", queryIndex+1, query)

			// Execute main query
			location := fmt.Sprintf("%f,%f", centerLat, centerLng)
			page, err := executeQuery(query, location, radiusMeters, app.GoogleMaps.ApiKey, app.GoogleMaps.BaseURL)
			if err != nil {
				fmt.Printf("DEBUG: Error in query %d: %v\n", queryIndex+1, err)
				continue
			}

			fmt.Printf("DEBUG: Query %d - Status=%s, Results count=%d\n", queryIndex+1, page.Status, len(page.Results))
			if page.Status != "OK" && page.Status != "ZERO_RESULTS" {
				fmt.Printf("DEBUG: API error for query %d: %s\n", queryIndex+1, page.Status)
				continue
			}

			// Filter using optimized province boundaries check
			for i, r := range page.Results {
				if existingPlaceIDs[r.PlaceID] {
					continue // Skip duplicates
				}

				lat := r.Geometry.Location.Lat
				lng := r.Geometry.Location.Lng

				// Use optimized boundary check
				withinProvince := IsCoordinateInProvinceOptimized(lat, lng, provinceData, minLat, minLng, maxLat, maxLng)
				fmt.Printf("DEBUG: Query %d, Result %d - %s at lat=%f, lng=%f, withinProvince=%t\n", queryIndex+1, i, r.Name, lat, lng, withinProvince)

				// Include only if within province boundaries
				if withinProvince {
					aggregated.Results = append(aggregated.Results, r)
					existingPlaceIDs[r.PlaceID] = true
				}
			}

			// Fetch only 2 additional pages (reduced from 5)
			nextToken := page.NextPageToken
			for i := 0; i < 2 && nextToken != nil && *nextToken != ""; i++ {
				// Reduced wait time
				time.Sleep(1 * time.Second)

				tokenParams := url.Values{}
				tokenParams.Add("pagetoken", *nextToken)
				tokenParams.Add("key", app.GoogleMaps.ApiKey)
				tokenURL := fmt.Sprintf("%s%s", app.GoogleMaps.BaseURL, tokenParams.Encode())

				page2, err := executeQuery("", "", 0, app.GoogleMaps.ApiKey, tokenURL)
				if err != nil {
					break
				}

				if page2.Status != "OK" && page2.Status != "ZERO_RESULTS" {
					break
				}

				// Apply same optimized filtering to subsequent pages
				for _, r := range page2.Results {
					if existingPlaceIDs[r.PlaceID] {
						continue // Skip duplicates
					}

					lat := r.Geometry.Location.Lat
					lng := r.Geometry.Location.Lng

					// Use optimized boundary check
					withinProvince := IsCoordinateInProvinceOptimized(lat, lng, provinceData, minLat, minLng, maxLat, maxLng)

					// Include only if within province boundaries
					if withinProvince {
						aggregated.Results = append(aggregated.Results, r)
						existingPlaceIDs[r.PlaceID] = true
					}
				}
				nextToken = page2.NextPageToken
			}

			fmt.Printf("DEBUG: After query %d, total unique results count=%d\n", queryIndex+1, len(aggregated.Results))
		}

		// Optimized Strategy: Only try multiple center points if results are very low
		if len(aggregated.Results) < 5 {
			fmt.Printf("DEBUG: Results count is very low (%d), trying limited center points strategy\n", len(aggregated.Results))

			// Use only 3 most effective center points instead of 5
			centerPoints := []struct {
				lat, lng float64
				name     string
			}{
				{centerLat, centerLng, "center"},
				{minLat + (maxLat-minLat)*0.25, minLng + (maxLng-minLng)*0.25, "northwest"},
				{minLat + (maxLat-minLat)*0.75, minLng + (maxLng-minLng)*0.75, "southeast"},
			}

			for _, centerPoint := range centerPoints {
				// Early exit if we have enough results
				if len(aggregated.Results) >= 20 {
					break
				}

				fmt.Printf("DEBUG: Searching from center point: %s (%.6f, %.6f)\n", centerPoint.name, centerPoint.lat, centerPoint.lng)

				// Use smaller radius for multiple center points
				smallerRadius := radiusMeters / 2
				if smallerRadius < 5000 {
					smallerRadius = 5000
				}

				location := fmt.Sprintf("%f,%f", centerPoint.lat, centerPoint.lng)
				page, err := executeQuery(brandName, location, smallerRadius, app.GoogleMaps.ApiKey, app.GoogleMaps.BaseURL)
				if err != nil {
					continue
				}

				if page.Status != "OK" && page.Status != "ZERO_RESULTS" {
					continue
				}

				// Filter and add unique results using optimized check
				for _, r := range page.Results {
					if !existingPlaceIDs[r.PlaceID] {
						lat := r.Geometry.Location.Lat
						lng := r.Geometry.Location.Lng

						// Use optimized boundary check
						withinProvince := IsCoordinateInProvinceOptimized(lat, lng, provinceData, minLat, minLng, maxLat, maxLng)

						// Include only if within province boundaries
						if withinProvince {
							aggregated.Results = append(aggregated.Results, r)
							existingPlaceIDs[r.PlaceID] = true
						}
					}
				}
			}
		}

		aggregated.Status = "OK"
		if app.Redis != nil {
			cacheFranchiseLocationsResponse(ctx, app.Redis, cacheKey, &aggregated)
		}
		c.JSON(http.StatusOK, aggregated)
		return
	}

	// Optimized default behavior: brand-only text search (first page only)
	placesResponse, err := executeQuery(brandName, "", 0, app.GoogleMaps.ApiKey, app.GoogleMaps.BaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to make request to Google Maps API: %v", err),
		})
		return
	}

	// Check if the API request was successful
	if placesResponse.Status != "OK" && placesResponse.Status != "ZERO_RESULTS" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Google Maps API error: %s", placesResponse.Status),
		})
		return
	}

	// Return the locations
	if app.Redis != nil && (placesResponse.Status == "OK" || placesResponse.Status == "ZERO_RESULTS") {
		cacheFranchiseLocationsResponse(ctx, app.Redis, cacheKey, placesResponse)
	}
	c.JSON(http.StatusOK, placesResponse)
}

func cacheFranchiseLocationsResponse(ctx context.Context, client *redis.Client, key string, response *GetFranchiseLocationsResponse) {
	if client == nil || response == nil {
		return
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("DEBUG: Failed to marshal response for cache key=%s: %v\n", key, err)
		return
	}

	if err := client.Set(ctx, key, data, franchiseLocationsCacheTTL).Err(); err != nil {
		fmt.Printf("DEBUG: Failed to set Redis cache for key=%s: %v\n", key, err)
	}
}
