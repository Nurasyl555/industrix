// lib/listing.ts
// Calls to the Listing Module (backend/modules/listing/)

import { publicGet, authGet, authPost, authPut, authDelete } from "./api";

export type ListingType = "sale" | "rental";
export type PricePeriod = "day" | "week" | "month" | "";
export type ListingStatus = "draft" | "active" | "archived";

export interface ListingView {
  id: string;
  equipment_id: string;
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
  status: ListingStatus;
  created_at: string;
}

export interface Listing {
  id: string;
  equipment_id: string;
  seller_id: string;
  listing_type: ListingType;
  price: number;
  price_period?: PricePeriod;
  status: ListingStatus;
  created_at: string;
  updated_at: string;
}

export interface ListingFilters {
  category_id?: string;
  region?: string;
  listing_type?: ListingType;
  condition?: "new" | "used";
  search?: string;
  price_min?: number;
  price_max?: number;
  sort?: "price_asc" | "price_desc" | "newest";
  page?: number;
  limit?: number;
}

export interface ListingPage {
  items: ListingView[];
  total: number;
  page: number;
  limit: number;
}

function toQuery(filters: ListingFilters): string {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value !== undefined && value !== "" && value !== 0) {
      params.set(key, String(value));
    }
  });
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export function listActiveListings(filters: ListingFilters = {}) {
  return publicGet<ListingPage>(`/listings${toQuery(filters)}`);
}

export function getListing(id: string) {
  return publicGet<ListingView>(`/listings/${id}`);
}

export function createListing(input: {
  equipment_id: string;
  listing_type: ListingType;
  price: number;
  price_period?: PricePeriod;
}) {
  return authPost<Listing>("/listings", input);
}

export async function listMyListings(): Promise<Listing[]> {
  return (await authGet<Listing[] | null>("/my-listings")) ?? [];
}

export function updateListingPrice(id: string, price: number, price_period?: PricePeriod) {
  return authPut<Listing>(`/my-listings/${id}`, { price, price_period });
}

export function publishListing(id: string) {
  return authPut<void>(`/my-listings/${id}/publish`);
}

export function archiveListing(id: string) {
  return authPut<void>(`/my-listings/${id}/archive`);
}

export function deleteListing(id: string) {
  return authDelete<void>(`/my-listings/${id}`);
}
