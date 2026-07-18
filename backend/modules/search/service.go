package search

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"github.com/industrix/backend/pkg/logger"
	"github.com/industrix/backend/pkg/redis"
)

// cacheTTL is how long a hot query's result is cached in Redis.
const cacheTTL = 60 * time.Second

// Service exposes buyer-facing search plus the index-mutation operations the
// Kafka consumer drives.
type Service interface {
	Search(ctx context.Context, q Query) (*Result, error)

	// Index mutations, driven by domain events.
	UpsertEquipment(ctx context.Context, d Doc) error
	DeleteEquipment(ctx context.Context, equipmentID string) error
	SetListingActive(ctx context.Context, equipmentID, listingID, listingType string, price float64, pricePeriod string) error
	SetListingInactive(ctx context.Context, equipmentID string) error
}

type service struct {
	os    *OpenSearchClient
	cache *redis.Client
	log   *logger.Logger
}

// NewService wires the OpenSearch client and an optional Redis cache (may be nil).
func NewService(os *OpenSearchClient, cache *redis.Client) Service {
	return &service{os: os, cache: cache, log: logger.New("search-service")}
}

func (s *service) Search(ctx context.Context, q Query) (*Result, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}

	key := cacheKey(q)
	if cached := s.cacheGet(ctx, key); cached != nil {
		return cached, nil
	}

	body := buildQuery(q)
	resp, err := s.os.Search(ctx, body)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Items:  make([]Doc, 0, len(resp.Hits.Hits)),
		Total:  resp.Hits.Total.Value,
		Page:   q.Page,
		Limit:  q.Limit,
		Facets: map[string]Facet{},
	}
	for _, h := range resp.Hits.Hits {
		res.Items = append(res.Items, h.Source)
	}
	for name, agg := range resp.Aggregations {
		f := Facet{}
		for _, b := range agg.Buckets {
			f[fmt.Sprintf("%v", b.Key)] = b.Count
		}
		res.Facets[name] = f
	}

	s.cacheSet(ctx, key, res)
	return res, nil
}

// === Index mutations ===

func (s *service) UpsertEquipment(ctx context.Context, d Doc) error {
	return s.os.Upsert(ctx, d.EquipmentID, map[string]any{
		"equipment_id": d.EquipmentID,
		"title":        d.Title,
		"description":  d.Description,
		"category_id":  d.CategoryID,
		"region":       d.Region,
		"condition":    d.Condition,
		"image_url":    d.ImageURL,
		"seller_id":    d.SellerID,
	})
}

func (s *service) DeleteEquipment(ctx context.Context, equipmentID string) error {
	return s.os.Delete(ctx, equipmentID)
}

func (s *service) SetListingActive(ctx context.Context, equipmentID, listingID, listingType string, price float64, pricePeriod string) error {
	return s.os.Upsert(ctx, equipmentID, map[string]any{
		"equipment_id": equipmentID,
		"listing_id":   listingID,
		"listing_type": listingType,
		"price":        price,
		"price_period": pricePeriod,
		"active":       true,
	})
}

func (s *service) SetListingInactive(ctx context.Context, equipmentID string) error {
	return s.os.Upsert(ctx, equipmentID, map[string]any{"active": false})
}

// === Query building ===

// buildQuery translates a Query into an OpenSearch request body: a bool query
// (full-text must + keyword/range filters) plus terms aggregations for facets.
func buildQuery(q Query) map[string]any {
	var must []any
	if q.Text != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q.Text,
				"fields": []string{"title^2", "description"},
			},
		})
	}

	filter := []any{map[string]any{"term": map[string]any{"active": true}}}
	addTerm := func(field, val string) {
		if val != "" {
			filter = append(filter, map[string]any{"term": map[string]any{field: val}})
		}
	}
	addTerm("category_id", q.CategoryID)
	addTerm("region", q.Region)
	addTerm("condition", q.Condition)
	addTerm("listing_type", q.ListingType)

	if q.PriceMin > 0 || q.PriceMax > 0 {
		rng := map[string]any{}
		if q.PriceMin > 0 {
			rng["gte"] = q.PriceMin
		}
		if q.PriceMax > 0 {
			rng["lte"] = q.PriceMax
		}
		filter = append(filter, map[string]any{"range": map[string]any{"price": rng}})
	}

	if must == nil {
		must = []any{map[string]any{"match_all": map[string]any{}}}
	}

	body := map[string]any{
		"from": (q.Page - 1) * q.Limit,
		"size": q.Limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
		"aggs": map[string]any{
			"category_id":  termsAgg("category_id"),
			"region":       termsAgg("region"),
			"condition":    termsAgg("condition"),
			"listing_type": termsAgg("listing_type"),
		},
	}

	switch q.Sort {
	case "price_asc":
		body["sort"] = []any{map[string]any{"price": "asc"}}
	case "price_desc":
		body["sort"] = []any{map[string]any{"price": "desc"}}
	}
	return body
}

func termsAgg(field string) map[string]any {
	return map[string]any{"terms": map[string]any{"field": field, "size": 50}}
}

// === Redis cache ===

func cacheKey(q Query) string {
	b, _ := json.Marshal(q)
	return fmt.Sprintf("search:%x", sha1.Sum(b))
}

func (s *service) cacheGet(ctx context.Context, key string) *Result {
	if s.cache == nil {
		return nil
	}
	raw, err := s.cache.Get(ctx, key)
	if err != nil || raw == "" {
		return nil
	}
	var r Result
	if json.Unmarshal([]byte(raw), &r) != nil {
		return nil
	}
	return &r
}

func (s *service) cacheSet(ctx context.Context, key string, r *Result) {
	if s.cache == nil {
		return
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	if err := s.cache.Set(ctx, key, b, cacheTTL); err != nil {
		s.log.Debug().Err(err).Msg("search cache set failed")
	}
}
