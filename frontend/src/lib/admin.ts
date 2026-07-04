// lib/admin.ts
// Admin-only calls (backend: admin-role gated /admin/*). All via the proxy.

import { authGet, authPut } from "./api";
import { type Company } from "./company";
import { type ListingView } from "./listing";

export async function listCompanies(status?: string): Promise<Company[]> {
  const q = status ? `?status=${status}` : "";
  // Go serializes an empty slice as null — coerce so callers can rely on [].
  return (await authGet<Company[] | null>(`/admin/companies${q}`)) ?? [];
}

export function verifyCompany(id: string) {
  return authPut<void>(`/admin/companies/${id}/verify`);
}

export function rejectCompany(id: string, note: string) {
  return authPut<void>(`/admin/companies/${id}/reject`, { note });
}

export async function listModerationQueue(): Promise<ListingView[]> {
  return (await authGet<ListingView[] | null>("/admin/listings/moderation")) ?? [];
}

export function approveListing(id: string) {
  return authPut<void>(`/admin/listings/${id}/approve`);
}

export function rejectListing(id: string) {
  return authPut<void>(`/admin/listings/${id}/reject`);
}
