# Coverage Gaps Analysis Feature

This feature identifies areas within cities that have no franchise spots within a specified radius (default 500 meters). It helps franchise owners and analysts understand where there are opportunities for expansion.

## API Endpoint

```
GET /franchise/coverage-gaps
```

## Query Parameters

- `city` (required): The city name to analyze (e.g., "Jakarta", "Surabaya")
- `brand` (optional): Specific brand name to analyze. If not provided, analyzes all franchises in the city
- `radius` (optional): Analysis radius in meters. Default is 500 meters

## Example Requests

### Analyze all franchises in Jakarta with 500m radius
```
GET /franchise/coverage-gaps?city=Jakarta
```

### Analyze specific brand in Surabaya with 300m radius
```
GET /franchise/coverage-gaps?city=Surabaya&brand=KFC&radius=300
```

### Analyze all franchises in Bandung with 1km radius
```
GET /franchise/coverage-gaps?city=Bandung&radius=1000
```

## Response Format

```json
{
  "city": "Jakarta",
  "total_gaps": 15,
  "total_gap_area": 785398.1633974483,
  "coverage_gaps": [
    {
      "center": {
        "lat": -6.2088,
        "lng": 106.8456
      },
      "radius": 500,
      "area": 785398.1633974483,
      "nearest_distance": 750.5
    }
  ],
  "analysis_radius": 500,
  "status": "OK"
}
```

## Response Fields

- `city`: The analyzed city name
- `total_gaps`: Number of coverage gap areas found
- `total_gap_area`: Total area of all gaps in square meters
- `coverage_gaps`: Array of gap areas with their details
  - `center`: Latitude and longitude of the gap center
  - `radius`: Analysis radius used (in meters)
  - `area`: Area of this gap in square meters
  - `nearest_distance`: Distance to the nearest franchise (in meters). -1 if no franchises found
- `analysis_radius`: The radius used for analysis (in meters)
- `status`: Response status

## How It Works

1. **City Geocoding**: The system first geocodes the city to get its boundaries and center coordinates
2. **Franchise Discovery**: Searches for all franchise locations within the city using Google Maps Places API
3. **Grid Analysis**: Creates a grid across the city area and analyzes each grid point
4. **Gap Identification**: Identifies areas where no franchises exist within the specified radius
5. **Gap Merging**: Merges nearby gaps to avoid fragmentation and provide meaningful results

## Use Cases

- **Franchise Expansion Planning**: Identify underserved areas for new franchise locations
- **Market Analysis**: Understand franchise density and coverage patterns
- **Competitive Analysis**: Find gaps in competitor coverage
- **Investment Decisions**: Support data-driven decisions for franchise investments

## Technical Details

- Uses Google Maps Geocoding API to get city boundaries
- Uses Google Maps Places API to find franchise locations
- Implements grid-based analysis with configurable spacing
- Merges nearby gaps using clustering algorithm
- Calculates accurate distances using Haversine formula
- Handles edge cases like cities with no franchises

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- `400 Bad Request`: Missing required parameters or invalid city
- `500 Internal Server Error`: Google Maps API errors or processing failures

## Rate Limiting

This feature makes multiple API calls to Google Maps services. Consider implementing rate limiting for production use to avoid hitting API quotas.

## Performance Considerations

- Analysis time increases with city size and grid density
- Larger analysis radius may require more processing time
- Consider caching results for frequently requested cities 