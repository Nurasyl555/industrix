// lib/deal.ts
// Calls to the Deal Module (backend/modules/deal/) — buyer inquiries.

import { authGet, authPost, authPut } from "./api";

export interface Deal {
  id: string;
  listing_id: string;
  buyer_id: string;
  seller_id: string;
  message: string;
  status: "inquiry" | "closed";
  created_at: string;
  updated_at: string;
  role: "buyer" | "seller";
}

export function createDeal(listing_id: string, message: string) {
  return authPost<Deal>("/deals", { listing_id, message });
}

export function listMyDeals() {
  return authGet<Deal[]>("/my-deals");
}

export function getDeal(id: string) {
  return authGet<Deal>(`/deals/${id}`);
}

export function closeDeal(id: string) {
  return authPut<void>(`/deals/${id}/close`);
}
