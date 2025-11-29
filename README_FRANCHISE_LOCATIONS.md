# Franchise Locations API

This API provides franchise location information by integrating with Google Maps Places API to fetch real-time location data based on brand names.

## Features

- **Google Maps Integration**: Fetches real franchise locations using Google Maps Places API
- **Fallback Support**: Falls back to database search if Google Maps API is unavailable
- **Sample Data**: Provides sample location data for demonstration purposes
- **Flexible Search**: Supports brand name and optional city parameters

## API Endpoint

### GET /franchise/locations

Fetches franchise locations by brand name.

#### Query Parameters

- `brand` (required): The brand name to search for
- `city` (optional): City to narrow down the search (e.g., "Jakarta", "Surabaya")

#### Example Request

```bash
GET /franchise/locations?brand=KFC&city=Jakarta
```

#### Example Response

```json
{
  "locations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "franchise_id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "KFC - Jakarta Pusat",
      "address": "Jl. Sudirman No. 123, Jakarta Pusat, DKI Jakarta",
      "lat": -6.2088,
      "lng": 106.8456,
      "type": "Franchise",
      "rating": 4.5,
      "phone": "+62-21-1234-5678",
      "operating_hours": "08:00 - 22:00",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1,
  "message": "Berhasil mengambil lokasi franchise dari Google Maps"
}
```

## Setup

### 1. Environment Variables

Add the following environment variable to your `.env` file:

```bash
GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here
```

Or alternatively:

```bash
REACT_APP_GOOGLE_MAPS_API_KEY=your_google_maps_api_key_here
```

### 2. Google Maps API Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the following APIs:
   - Places API
   - Geocoding API
   - Maps JavaScript API
4. Create credentials (API Key)
5. Restrict the API key to your domain for security

### 3. Database Setup

Run the migration to create the locations table:

```bash
psql -U your_username -d your_database -f migrations/create_locations_table.sql
```

## How It Works

1. **Primary Search**: The API first attempts to fetch locations from Google Maps Places API
2. **Fallback to Database**: If Google Maps API fails or returns no results, it searches the local database
3. **Sample Data**: If both sources fail, it returns sample location data for demonstration

## Google Maps API Integration

The API uses Google Maps Places API with the following features:

- **Text Search**: Searches for establishments by brand name
- **Location Bias**: Biases search results to Indonesia
- **Language Support**: Returns results in Indonesian
- **Establishment Type**: Focuses on business establishments
- **Detailed Information**: Includes ratings, phone numbers, and operating hours

## Error Handling

The API gracefully handles various error scenarios:

- Missing API key
- Google Maps API failures
- Database connection issues
- No results found

## Frontend Integration

The API is designed to work seamlessly with your React frontend:

```javascript
// Example usage in React component
const getFranchiseLocations = async (brandName) => {
  const response = await fetch(`${API_URL}/franchise/locations?brand=${encodeURIComponent(brandName)}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  if (!response.ok) {
    throw new Error('Gagal mengambil lokasi franchise');
  }
  
  return await response.json();
};
```

## Rate Limiting

Google Maps Places API has the following limits:
- **Free Tier**: 1,000 requests per day
- **Paid Tier**: 100,000 requests per day

Consider implementing caching for frequently searched brands to reduce API calls.

## Security Considerations

1. **API Key Protection**: Never expose your Google Maps API key in client-side code
2. **Request Validation**: The API validates all input parameters
3. **Error Messages**: Error messages don't expose sensitive information
4. **CORS**: Configure CORS properly for your frontend domain

## Troubleshooting

### Common Issues

1. **"Google Maps API key not configured"**
   - Check your environment variables
   - Ensure the API key is properly set

2. **"Google Maps API error: REQUEST_DENIED"**
   - Verify your API key is correct
   - Check if the Places API is enabled
   - Verify billing is set up for your Google Cloud project

3. **"Google Maps API error: OVER_QUERY_LIMIT"**
   - You've exceeded your daily quota
   - Consider upgrading to a paid plan or implementing caching

4. **No locations returned**
   - Check if the brand name exists
   - Try adding a city parameter to narrow the search
   - Verify the Google Maps API is working

### Debug Mode

Enable debug logging by checking the console output for detailed error messages and API responses.

## Future Enhancements

- **Caching Layer**: Implement Redis caching for frequently searched brands
- **Geolocation**: Add support for searching by user's current location
- **Advanced Filtering**: Support for filtering by rating, distance, etc.
- **Batch Operations**: Support for searching multiple brands at once
- **Analytics**: Track popular searches and location views 