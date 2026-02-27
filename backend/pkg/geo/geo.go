package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/industrix/pkg/logger"
)

// Config holds geo service configuration
type Config struct {
	2GISAPIKey  string
	CacheTTL    time.Duration
	RegionTable map[string]Region
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		2GISAPIKey: getEnv("2GIS_API_KEY", ""),
		CacheTTL:   24 * time.Hour,
		RegionTable: getDefaultRegions(),
	}
}

// Region represents a KZ/CIS region
type Region struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	NameKK  string  `json:"name_kk"`
	NameRU  string  `json:"name_ru"`
	Code    string  `json:"code"`
	Center  LatLng `json:"center"`
	Country string  `json:"country"`
}

// LatLng represents latitude and longitude
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GeocodingResult represents the result of geocoding
type GeocodingResult struct {
	Address   string  `json:"address"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	LatLng    LatLng  `json:"lat_lng"`
	Formatted string  `json:"formatted"`
}

// Client handles geographic operations
type Client struct {
	config *Config
	log    *logger.Logger
	client *http.Client
}

func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Client{
		config: cfg,
		log:    logger.New("geo-client"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// getDefaultRegions returns default KZ regions
func getDefaultRegions() map[string]Region {
	return map[string]Region{
		"almaty": {
			ID:      "almaty",
			Name:    "Almaty",
			NameKK:  "Алматы",
			NameRU:  "Алматы",
			Code:    "ALA",
			Center:  LatLng{Lat: 43.2220, Lng: 76.8512},
			Country: "KZ",
		},
		"astana": {
			ID:      "astana",
			Name:    "Astana",
			NameKK:  "Астана",
			NameRU:  "Астана",
			Code:    "AST",
			Center:  LatLng{Lat: 51.1694, Lng: 71.4491},
			Country: "KZ",
		},
		"akmola": {
			ID:      "akmola",
			Name:    "Akmola Region",
			NameKK:  "Ақмола облысы",
			NameRU:  "Акмолинская область",
			Code:    "AKM",
			Center:  LatLng{Lat: 51.7892, Lng: 69.3684},
			Country: "KZ",
		},
		"aktobe": {
			ID:      "aktobe",
			Name:    "Aktobe Region",
			NameKK:  "Ақтөбе облысы",
			NameRU:  "Актюбинская область",
			Code:    "AKT",
			Center:  LatLng{Lat: 50.2797, Lng: 57.2072},
			Country: "KZ",
		},
		"almaty_region": {
			ID:      "almaty_region",
			Name:    "Almaty Region",
			NameKK:  "Алматы облысы",
			NameRU:  "Алматинская область",
			Code:    "ALM",
			Center:  LatLng{Lat: 44.6994, Lng: 78.4118},
			Country: "KZ",
		},
		"atyrau": {
			ID:      "atyrau",
			Name:    "Atyrau Region",
			NameKK:  "Атырау облысы",
			NameRU:  "Атырауская область",
			Code:    "ATY",
			Center:  LatLng{Lat: 47.1064, Lng: 51.9245},
			Country: "KZ",
		},
		"east_kazakhstan": {
			ID:      "east_kazakhstan",
			Name:    "East Kazakhstan Region",
			NameKK:  "Шығыс Қазақстан облысы",
			NameRU:  "Восточно-Казахстанская область",
			Code:    "VKO",
			Center:  LatLng{Lat: 49.9430, Lng: 82.6210},
			Country: "KZ",
		},
		"zhambyl": {
			ID:      "zhambyl",
			Name:    "Zhambyl Region",
			NameKK:  "Жамбыл облысы",
			NameRU:  "Жамбылская область",
			Code:    "ZHA",
			Center:  LatLng{Lat: 42.8734, Lng: 71.3693},
			Country: "KZ",
		},
		"west_kazakhstan": {
			ID:      "west_kazakhstan",
			Name:    "West Kazakhstan Region",
			NameKK: "Батыс Қазақстан облысы",
			NameRU: "Западно-Казахстанская область",
			Code:   "ZAP",
			Center: LatLng{Lat: 51.2254, Lng: 51.3126},
			Country: "KZ",
		},
		"karaganda": {
			ID:      "karaganda",
			Name:    "Karaganda Region",
			NameKK:  "Қарағанды облысы",
			NameRU:  "Карагандинская область",
			Code:    "KAR",
			Center:  LatLng{Lat: 49.8069, Lng: 73.0781},
			Country: "KZ",
		},
		"kostanay": {
			ID:      "kostanay",
			Name:    "Kostanay Region",
			NameKK:  "Қостанай облысы",
			NameRU:  "Костанайская область",
			Code:    "KUS",
			Center:  LatLng{Lat: 53.2141, Lng: 63.6246},
			Country: "KZ",
		},
		"kyzylorda": {
			ID:      "kyzylorda",
			Name:    "Kyzylorda Region",
			NameKK:  "Қызылорда облысы",
			NameRU:  "Кызылординская область",
			Code:   "KZY",
			Center: LatLng{Lat: 44.8528, Lng: 65.5089},
			Country: "KZ",
		},
		"mangystau": {
			ID:      "mangystau",
			Name:    "Mangystau Region",
			NameKK:  "Маңғыстау облысы",
			NameRU:  "Мангистауская область",
			Code:    "MAN",
			Center:  LatLng{Lat: 43.5903, Lng: 51.9207},
			Country: "KZ",
		},
		"pavlodar": {
			ID:      "pavlodar",
			Name:    "Pavlodar Region",
			NameKK:  "Павлодар облысы",
			NameRU:  "Павлодарская область",
			Code:    "PAV",
			Center:  LatLng{Lat: 52.2873, Lng: 76.9754},
			Country: "KZ",
		},
		"north_kazakhstan": {
			ID:      "north_kazakhstan",
			Name:    "North Kazakhstan Region",
			NameKK:  "Солтүстік Қазақстан облысы",
			NameRU:  "Северо-Казахстанская область",
			Code:    "SEV",
			Center:  LatLng{Lat: 54.8764, Lng: 69.1571},
			Country: "KZ",
		},
		"turkestan": {
			ID:      "turkestan",
			Name:    "Turkestan Region",
			NameKK:  "Түркістан облысы",
			NameRU:  "Туркестанская область",
			Code:    "TUR",
			Center:  LatLng{Lat: 43.3000, Lng: 68.2500},
			Country: "KZ",
		},
		"shymkent": {
			ID:      "shymkent",
			Name:    "Shymkent",
			NameKK:  "Шымкент",
			NameRU:  "Шымкент",
			Code:    "SHY",
			Center:  LatLng{Lat: 42.3155, Lng: 69.2789},
			Country: "KZ",
		},
	}
}

// LookupRegion returns region info by ID or name
func (c *Client) LookupRegion(ctx context.Context, regionID string) (*Region, error) {
	region, ok := c.config.RegionTable[regionID]
	if !ok {
		return nil, fmt.Errorf("region not found: %s", regionID)
	}
	return &region, nil
}

// GetAllRegions returns all available regions
func (c *Client) GetAllRegions(ctx context.Context) []Region {
	regions := make([]Region, 0, len(c.config.RegionTable))
	for _, r := range c.config.RegionTable {
		regions = append(regions, r)
	}
	return regions
}

// Geocode converts an address to coordinates using 2GIS API
// Note: No PII is sent - only address coordinates are returned
func (c *Client) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if c.config.2GISAPIKey == "" {
		c.log.Warn().Msg("2GIS API key not configured, returning empty result")
		return &GeocodingResult{}, nil
	}

	url := fmt.Sprintf(
		"https://catalog-api.2gis.com/6.0/_geocoding?q=%s&key=%s",
		url.QueryEscape(address),
		c.config.2GISAPIKey,
	)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call geocoding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			City    string `json:"city_name"`
			Region  string `json:"region_name"`
			LatLng  LatLng `json:"point"`
			Address string `json:"full_name"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode geocoding response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", address)
	}

	item := result.Items[0]
	return &GeocodingResult{
		City:      item.City,
		Region:    item.Region,
		LatLng:    item.LatLng,
		Address:   item.Address,
		Formatted: item.Address,
	}, nil
}

// ReverseGeocode converts coordinates to address
func (c *Client) ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodingResult, error) {
	if c.config.2GISAPIKey == "" {
		c.log.Warn().Msg("2GIS API key not configured, returning empty result")
		return &GeocodingResult{}, nil
	}

	url := fmt.Sprintf(
		"https://catalog-api.2gis.com/6.0/reverse_geocoding?point=%f,%f&key=%s",
		lng, lat, c.config.2GISAPIKey,
	)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call reverse geocoding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reverse geocoding API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			City    string `json:"city_name"`
			Region  string `json:"region_name"`
			LatLng  LatLng `json:"point"`
			Address string `json:"full_name"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode reverse geocoding response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no results found for coordinates: %f, %f", lat, lng)
	}

	item := result.Items[0]
	return &GeocodingResult{
		City:      item.City,
		Region:    item.Region,
		LatLng:    LatLng{Lat: lat, Lng: lng},
		Address:   item.Address,
		Formatted: item.Address,
	}, nil
}
