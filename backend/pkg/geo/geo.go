package geo

import (
	"context"
	"fmt"
)

type Region struct {
	ID   string
	Name string
	Lat  float64
	Lng  float64
}

type GeoService interface {
	LookupRegion(ctx context.Context, lat, lng float64) (*Region, error)
	Geocode(ctx context.Context, address string) (float64, float64, error)
}

type geoService struct {
	// Add DB or other dependencies here
}

func NewService() GeoService {
	return &geoService{}
}

func (s *geoService) LookupRegion(ctx context.Context, lat, lng float64) (*Region, error) {
	// Mock implementation
	return &Region{
		ID:   "kz-ala",
		Name: "Almaty",
		Lat:  43.238949,
		Lng:  76.889709,
	}, nil
}

func (s *geoService) Geocode(ctx context.Context, address string) (float64, float64, error) {
	// Mock implementation (2GIS integration stub)
	fmt.Printf("Geocoding address: %s\n", address)
	return 43.238949, 76.889709, nil
}
