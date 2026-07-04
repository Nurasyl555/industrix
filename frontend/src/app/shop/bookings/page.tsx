"use client";

// src/app/shop/bookings/page.tsx
// The current user's rental bookings (as renter). Cancellable.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { CalendarDays } from "lucide-react";
import { listMyBookings, cancelBooking, type Booking } from "@/lib/booking";
import { friendlyError } from "@/lib/api";
import { Button } from "@/components/ui/button";

export default function BookingsPage() {
  const router = useRouter();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setBookings(await listMyBookings());
  }

  useEffect(() => {
    load()
      .catch((err) => {
        if (friendlyError(err) === "Please sign in to continue.") {
          router.push("/auth/login");
          return;
        }
        setError(friendlyError(err));
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function handleCancel(id: string) {
    try {
      await cancelBooking(id);
      setBookings((prev) => prev.map((b) => (b.id === id ? { ...b, status: "cancelled" } : b)));
    } catch (err) {
      setError(friendlyError(err));
    }
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="mx-auto max-w-3xl px-6 py-8">
        <h1 className="mb-6 text-2xl font-extrabold text-gray-900">My Bookings</h1>

        {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

        {loading ? (
          <div className="py-24 text-center text-[14px] text-gray-400">Loading…</div>
        ) : bookings.length === 0 ? (
          <div className="py-24 text-center text-[14px] text-gray-400">
            No bookings yet. Rent equipment from a rental listing to see it here.
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {bookings.map((b) => (
              <div key={b.id} className="flex items-center justify-between gap-4 rounded-xl border border-gray-200 p-4">
                <div className="flex items-center gap-3">
                  <CalendarDays size={20} className="shrink-0 text-gray-400" />
                  <div>
                    <p className="text-sm font-semibold text-gray-900">{b.start_date} → {b.end_date}</p>
                    <div className="flex items-center gap-2">
                      <span className={`text-[10px] font-bold uppercase ${b.status === "cancelled" ? "text-gray-400" : "text-emerald-600"}`}>
                        {b.status}
                      </span>
                      <Link href={`/shop/details?id=${b.listing_id}`} className="text-xs font-semibold text-blue-600 hover:underline">
                        View listing
                      </Link>
                    </div>
                  </div>
                </div>
                {b.status === "confirmed" && (
                  <Button variant="outline" size="sm" onClick={() => handleCancel(b.id)}>
                    Cancel
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
