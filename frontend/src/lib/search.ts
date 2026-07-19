// lib/search.ts
// Client for the OpenSearch-backed /search endpoint.
//
// This is the faceted search: full-text over title/description plus filters,
// and it returns per-value counts so the sidebar can show how many results
// each option would yield. The plain /listings endpoint (lib/listing.ts) is a
// straight SQL filter with no relevance ranking and no facets.

import { publicGet } from "./api";
import type { ListingView, ListingType, PricePeriod } from "./listing";

/** One indexed listing. Mirrors the backend's search.Doc. */
export interface SearchDoc {
  equipment_id: string;
  listing_id: string;
  title: string;
  description?: string;
  category_id: string;
  region: string;
  condition: "new" | "used";
  image_url?: string;
  seller_id: string;
  listing_type: ListingType;
  price: number;
  price_period?: PricePeriod;
  active: boolean;
}

/** value → number of matching results. */
export type Facet = Record<string, number>;

export interface SearchResult {
  items: SearchDoc[];
  total: number;
  page: number;
  limit: number;
  facets: {
    category_id?: Facet;
    region?: Facet;
    condition?: Facet;
    listing_type?: Facet;
  };
}

export interface SearchQuery {
  q?: string;
  category_id?: string;
  region?: string;
  condition?: string;
  listing_type?: string;
  min_price?: number;
  max_price?: number;
  sort?: "price_asc" | "price_desc" | "newest";
  page?: number;
  limit?: number;
}

function toQuery(params: SearchQuery): string {
  const qs = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "" && value !== null) {
      qs.set(key, String(value));
    }
  }
  const s = qs.toString();
  return s ? `?${s}` : "";
}

export function search(params: SearchQuery = {}) {
  return publicGet<SearchResult>(`/search${toQuery(params)}`);
}

/**
 * Adapts a search hit to the shape the listing cards render.
 *
 * The grid is driven by ListingView, and a search hit carries the same fields
 * under different names — notably listing_id, which is the id the details page
 * expects. Reusing the card keeps search and browse visually identical.
 */
export function docToListingView(doc: SearchDoc): ListingView {
  return {
    id: doc.listing_id,
    equipment_id: doc.equipment_id,
    title: doc.title,
    description: doc.description,
    category_id: doc.category_id,
    region: doc.region ?? "",
    condition: doc.condition,
    image_url: doc.image_url,
    seller_id: doc.seller_id,
    listing_type: doc.listing_type,
    price: doc.price,
    price_period: doc.price_period,
    status: "active",
    // The index doesn't carry created_at — nothing on the card shows it, and
    // ordering comes from the query, not this field.
    created_at: "",
  };
}
