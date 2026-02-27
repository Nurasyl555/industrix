package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/industrix/backend/pkg/logger"
)

// Config holds geo service configuration
type Config struct {
	TwoGISKey     string
	CacheEnabled  bool
	CacheDuration int // hours
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		TwoGISKey:     getEnv("GIS_2GIS_KEY", ""),
		CacheEnabled:  getEnv("GEO_CACHE_ENABLED", "true") == "true",
		CacheDuration: 24,
	}
}

// Region represents a region in Kazakhstan/CIS
type Region struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	NameKZ   string  `json:"name_kz"`
	Country  string  `json:"country"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone string  `json:"timezone"`
	ISOcode  string  `json:"iso_code"`
}

// Address represents a geocoded address
type Address struct {
	FullAddress string  `json:"full_address"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	District    string  `json:"district"`
	Street      string  `json:"street"`
	Building    string  `json:"building"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	PostalCode  string  `json:"postal_code"`
	TwoGISID    string  `json:"2gis_id"`
}

// GeocodingResult represents the result of geocoding
type GeocodingResult struct {
	Address Address
	Err     error
}

// Client handles geographic operations
type Client struct {
	config *Config
	log    *logger.Logger
	mu     sync.RWMutex
	// Cache for region lookups
	regionCache map[string]*Region
	// Cache for geocoding results
	geocodeCache map[string]*GeocodingResult
}

// NewClient creates a new geo client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		config:       cfg,
		log:          logger.New("geo-client"),
		regionCache:  make(map[string]*Region),
		geocodeCache: make(map[string]*GeocodingResult),
	}
}

// RegionData contains pre-populated Kazakhstan regions
var RegionData = []Region{
	{
		ID:       "kz-almaty",
		Name:     "Almaty",
		NameKZ:   "Алматы",
		Country:  "Kazakhstan",
		Lat:      43.2220,
		Lon:      76.8512,
		Timezone: "Asia/Almaty",
		ISOcode:  "ALA",
	},
	{
		ID:       "kz-astana",
		Name:     "Astana",
		NameKZ:   "Астана",
		Country:  "Kazakhstan",
		Lat:      51.1694,
		Lon:      71.4491,
		Timezone: "Asia/Almaty",
		ISOcode:  "AST",
	},
	{
		ID:       "kz-shymkent",
		Name:     "Shymkent",
		NameKZ:   "Шымкент",
		Country:  "Kazakhstan",
		Lat:      42.3175,
		Lon:      69.6519,
		Timezone: "Asia/Almaty",
		ISOcode:  "KZY",
	},
	{
		ID:       "kz-aktobe",
		Name:     "Aktobe",
		NameKZ:   "Ақтөбе",
		Country:  "Kazakhstan",
		Lat:      50.2799,
		Lon:      57.2072,
		Timezone: "Asia/Almaty",
		ISOcode:  "AKT",
	},
	{
		ID:       "kz-karaganda",
		Name:     "Karaganda",
		NameKZ:   "Қарағанды",
		Country:  "Kazakhstan",
		Lat:      49.8068,
		Lon:      73.0875,
		Timezone: "Asia/Almaty",
		ISOcode:  "KAR",
	},
	{
		ID:       "kz-aktau",
		Name:     "Aktau",
		NameKZ:   "Ақтау",
		Country:  "Kazakhstan",
		Lat:      43.6500,
		Lon:      51.1500,
		Timezone: "Asia/Almaty",
		ISOcode:  "MNG",
	},
	{
		ID:       "kz-pavlodar",
		Name:     "Pavlodar",
		NameKZ:   "Павлодар",
		Country:  "Kazakhstan",
		Lat:      52.2873,
		Lon:      76.9504,
		Timezone: "Asia/Almaty",
		ISOcode:  "PAV",
	},
	{
		ID:       "kz-ust-kamenogorsk",
		Name:     "Oskemen",
		NameKZ:   "Өскемен",
		Country:  "Kazakhstan",
		Lat:      49.9783,
		Lon:      82.6059,
		Timezone: "Asia/Almaty",
		ISOcode:  "VOS",
	},
	{
		ID:       "kz-semey",
		Name:     "Semey",
		NameKZ:   "Семей",
		Country:  "Kazakhstan",
		Lat:      50.4100,
		Lon:      80.2500,
		Timezone: "Asia/Almaty",
		ISOcode:  "SEM",
	},
	{
		ID:       "kz-kostanay",
		Name:     "Kostanay",
		NameKZ:   "Қостанай",
		Country:  "Kazakhstan",
		Lat:      53.2143,
		Lon:      63.6246,
		Timezone: "Asia/Almaty",
		ISOcode:  "KUS",
	},
	{
		ID:       "kz-petropavlovsk",
		Name:     "Petropavlovsk",
		NameKZ:   "Петропавл",
		Country:  "Kazakhstan",
		Lat:      54.8753,
		Lon:      69.1368,
		Timezone: "Asia/Almaty",
		ISOcode:  "PET",
	},
	{
		ID:       "kz-uralsk",
		Name:     "Oral",
		NameKZ:   "Орал",
		Country:  "Kazakhstan",
		Lat:      51.2268,
		Lon:      51.3865,
		Timezone: "Asia/Almaty",
		ISOcode:  "ZKO",
	},
	{
		ID:       "ru-moscow",
		Name:     "Moscow",
		NameKZ:   "Москва",
		Country:  "Russia",
		Lat:      55.7558,
		Lon:      37.6173,
		Timezone: "Europe/Moscow",
		ISOcode:  "MOW",
	},
	{
		ID:       "ru-saint-petersburg",
		Name:     "Saint Petersburg",
		NameKZ:   "Санкт-Петербург",
		Country:  "Russia",
		Lat:      59.9343,
		Lon:      30.3351,
		Timezone: "Europe/Moscow",
		ISOcode:  "SPE",
	},
	{
		ID:       "uz-tashkent",
		Name:     "Tashkent",
		NameKZ:   "Ташкент",
		Country:  "Uzbekistan",
		Lat:      41.2995,
		Lon:      69.2401,
		Timezone: "Asia/Tashkent",
		ISOcode:  "TAS",
	},
}

// LookupRegion returns region information by ID or name
func (c *Client) LookupRegion(query string) (*Region, error) {
	// Check cache first
	c.mu.RLock()
	if cached, ok := c.regionCache[query]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	// Search in pre-populated data
	for _, region := range RegionData {
		if region.ID == query || region.Name == query || region.NameKZ == query {
			// Cache the result
			c.mu.Lock()
			c.regionCache[query] = &region
			c.mu.Unlock()
			return &region, nil
		}
	}

	return nil, fmt.Errorf("region not found: %s", query)
}

// GetAllRegions returns all available regions
func (c *Client) GetAllRegions() []Region {
	return RegionData
}

// GetRegionsByCountry returns regions filtered by country
func (c *Client) GetRegionsByCountry(country string) []Region {
	var result []Region
	for _, region := range RegionData {
		if region.Country == country {
			result = append(result, region)
		}
	}
	return result
}

// Geocode performs geocoding using 2GIS API
// Note: No PII is sent to external APIs - only addresses
func (c *Client) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if c.config.TwoGISKey == "" {
		return &GeocodingResult{
			Err: fmt.Errorf("2GIS API key not configured"),
		}, nil
	}

	// Check cache
	c.mu.RLock()
	if cached, ok := c.geocodeCache[address]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	// Make request to 2GIS API
	url := fmt.Sprintf("https://catalog-api.2gis.com/6.0.1/geocoding?q=%s&key=%s", address, c.config.TwoGISKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	var result struct {
		Result struct {
			Items []struct {
				FullAddress string `json:"full_address_name"`
				Point       struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"point"`
				Attrs struct {
					City       string `json:"city_name"`
					District   string `json:"district_name"`
					Street     string `json:"street_name"`
					Building   string `json:"building_name"`
					PostalCode string `json:"postal_code"`
				} `json:"attrs"`
			} `json:"items"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	if len(result.Result.Items) == 0 {
		return &GeocodingResult{Err: fmt.Errorf("no results found")}, nil
	}

	item := result.Result.Items[0]
	geoResult := &GeocodingResult{
		Address: Address{
			FullAddress: item.FullAddress,
			City:        item.Attrs.City,
			District:    item.Attrs.District,
			Street:      item.Attrs.Street,
			Building:    item.Attrs.Building,
			Lat:         item.Point.Lat,
			Lon:         item.Point.Lon,
			PostalCode:  item.Attrs.PostalCode,
		},
	}

	// Cache the result
	c.mu.Lock()
	c.geocodeCache[address] = geoResult
	c.mu.Unlock()

	return geoResult, nil
}

// ReverseGeocode performs reverse geocoding
func (c *Client) ReverseGeocode(ctx context.Context, lat, lon float64) (*GeocodingResult, error) {
	if c.config.TwoGISKey == "" {
		return &GeocodingResult{
			Err: fmt.Errorf("2GIS API key not configured"),
		}, nil
	}

	url := fmt.Sprintf("https://catalog-api.2gis.com/6.0.1/geocoding?lat=%f&lon=%f&key=%s", lat, lon, c.config.TwoGISKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	var result struct {
		Result struct {
			Items []struct {
				FullAddress string `json:"full_address_name"`
				City        string `json:"city_name"`
				District    string `json:"district_name"`
				Street      string `json:"street_name"`
				Building    string `json:"building_name"`
			} `json:"items"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return &GeocodingResult{Err: err}, nil
	}

	if len(result.Result.Items) == 0 {
		return &GeocodingResult{Err: fmt.Errorf("no results found")}, nil
	}

	item := result.Result.Items[0]
	return &GeocodingResult{
		Address: Address{
			FullAddress: item.FullAddress,
			City:        item.City,
			District:    item.District,
			Street:      item.Street,
			Building:    item.Building,
			Lat:         lat,
			Lon:         lon,
		},
	}, nil
}
