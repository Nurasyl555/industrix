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

export interface DealMessage {
  id: string;
  deal_id: string;
  sender_id: string;
  body: string;
  created_at: string;
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

export function listDealMessages(dealId: string) {
  return authGet<DealMessage[]>(`/deals/${dealId}/messages`);
}

export function postDealMessage(dealId: string, body: string) {
  return authPost<DealMessage>(`/deals/${dealId}/messages`, { body });
}

// WebSocket URL for a deal's realtime thread. Derived from the public API URL
// (strip /api/v1, swap http→ws) so it points at the backend directly. The
// access_token cookie is host-scoped and rides along with the handshake, so
// no token needs to be exposed to JS.
export function dealSocketURL(dealId: string): string {
  const api = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";
  const base = api.replace(/\/api\/v1\/?$/, "").replace(/^http/, "ws");
  return `${base}/ws/deals/${dealId}`;
}
