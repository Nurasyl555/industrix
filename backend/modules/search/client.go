package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/industrix/backend/pkg/logger"
)

// indexName is the single OpenSearch index backing marketplace search.
const indexName = "listings"

// OpenSearchClient is a thin REST wrapper over OpenSearch. It intentionally
// uses net/http (no extra dependency) since we only need index/update/delete
// and a search-with-aggregations query.
type OpenSearchClient struct {
	baseURL  string
	username string
	password string
	http     *http.Client
	log      *logger.Logger
}

// NewOpenSearchClient builds a client from a comma-separated hosts string
// (only the first host is used) and optional basic-auth credentials.
func NewOpenSearchClient(hosts, username, password string) *OpenSearchClient {
	return &OpenSearchClient{
		baseURL:  strings.TrimRight(firstHost(hosts), "/"),
		username: username,
		password: password,
		http:     &http.Client{Timeout: 10 * time.Second},
		log:      logger.New("opensearch"),
	}
}

func firstHost(hosts string) string {
	parts := strings.Split(hosts, ",")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "http://localhost:9200"
	}
	return strings.TrimSpace(parts[0])
}

func (c *OpenSearchClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	return c.http.Do(req)
}

// EnsureIndex creates the index with an explicit mapping if it does not exist.
// Best-effort: called once at startup; a failure is logged, not fatal.
func (c *OpenSearchClient) EnsureIndex(ctx context.Context) error {
	head, err := c.do(ctx, http.MethodHead, "/"+indexName, nil)
	if err != nil {
		return err
	}
	head.Body.Close()
	if head.StatusCode == http.StatusOK {
		return nil // already exists
	}

	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"equipment_id": map[string]any{"type": "keyword"},
				"listing_id":   map[string]any{"type": "keyword"},
				"title":        map[string]any{"type": "text"},
				"description":  map[string]any{"type": "text"},
				"category_id":  map[string]any{"type": "keyword"},
				"region":       map[string]any{"type": "keyword"},
				"condition":    map[string]any{"type": "keyword"},
				"image_url":    map[string]any{"type": "keyword", "index": false},
				"seller_id":    map[string]any{"type": "keyword"},
				"listing_type": map[string]any{"type": "keyword"},
				"price":        map[string]any{"type": "double"},
				"price_period": map[string]any{"type": "keyword"},
				"active":       map[string]any{"type": "boolean"},
			},
		},
	}
	resp, err := c.do(ctx, http.MethodPut, "/"+indexName, mapping)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return c.errFrom(resp)
	}
	c.log.Info().Str("index", indexName).Msg("OpenSearch index ensured")
	return nil
}

// Upsert merges the given partial fields into the doc keyed by id, creating it
// if absent (doc_as_upsert). Used by both equipment.* and listing.* consumers.
func (c *OpenSearchClient) Upsert(ctx context.Context, id string, fields map[string]any) error {
	body := map[string]any{"doc": fields, "doc_as_upsert": true}
	resp, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/%s/_update/%s", indexName, id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return c.errFrom(resp)
	}
	return nil
}

// Delete removes the doc; a 404 is treated as success (already gone).
func (c *OpenSearchClient) Delete(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/%s/_doc/%s", indexName, id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 300 {
		return c.errFrom(resp)
	}
	return nil
}

// searchResponse maps the parts of the OpenSearch _search response we consume.
type searchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source Doc `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]struct {
		Buckets []struct {
			Key   any   `json:"key"`
			Count int64 `json:"doc_count"`
		} `json:"buckets"`
	} `json:"aggregations"`
}

// Search runs the query body and returns hits, total and facet buckets.
func (c *OpenSearchClient) Search(ctx context.Context, body map[string]any) (*searchResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/%s/_search", indexName), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, c.errFrom(resp)
	}
	var out searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *OpenSearchClient) errFrom(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("opensearch %s: %s", resp.Status, strings.TrimSpace(string(b)))
}
