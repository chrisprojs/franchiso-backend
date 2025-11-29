package config

import (
	"os"
)

type GoogleMapsConfig struct {
	ApiKey     string
	BaseURL    string
	GeocodeURL string
}

func NewGoogleMaps() *GoogleMapsConfig {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		// Fallback to a panic if not available
		panic("Unable to get Google Maps API KEY")
	}

	return &GoogleMapsConfig{
		ApiKey:     apiKey,
		BaseURL:    "https://maps.googleapis.com/maps/api/place/textsearch/json?",
		GeocodeURL: "https://maps.googleapis.com/maps/api/geocode/json?",
	}
}
