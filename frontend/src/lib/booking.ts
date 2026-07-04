// lib/booking.ts
// Calls to the Booking Module (backend/modules/booking/) — rental reservations.

import { publicGet, authGet, authPost, authPut } from "./api";

export interface Booking {
  id: string;
  listing_id: string;
  renter_id: string;
  owner_id: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  status: "confirmed" | "cancelled";
  created_at: string;
}

export interface DateRange {
  start_date: string;
  end_date: string;
}

/** Booked (unavailable) date ranges for a listing — public, drives the calendar. */
export async function getBookedDates(listingId: string): Promise<DateRange[]> {
  return (await publicGet<DateRange[] | null>(`/listings/${listingId}/booked-dates`)) ?? [];
}

export function createBooking(listing_id: string, start_date: string, end_date: string) {
  return authPost<Booking>("/bookings", { listing_id, start_date, end_date });
}

export async function listMyBookings(): Promise<Booking[]> {
  return (await authGet<Booking[] | null>("/my-bookings")) ?? [];
}

export function cancelBooking(id: string) {
  return authPut<void>(`/bookings/${id}/cancel`);
}
