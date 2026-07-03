"use client";

// src/app/shop/deals/page.tsx
// Inbox (as seller) / outbox (as buyer) of listing inquiries.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { listMyDeals, closeDeal, type Deal } from "@/lib/deal";
import { friendlyError } from "@/lib/api";
import { Button } from "@/components/ui/button";

function DealCard({ deal, onClose }: { deal: Deal; onClose: (id: string) => void }) {
  return (
    <div className="border border-gray-200 rounded-xl p-4 flex items-start justify-between gap-4">
      <div className="min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded ${deal.role === "buyer" ? "bg-blue-100 text-blue-700" : "bg-emerald-100 text-emerald-700"}`}>
            {deal.role === "buyer" ? "You inquired" : "Inquiry received"}
          </span>
          <span className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded ${deal.status === "closed" ? "bg-gray-100 text-gray-500" : "bg-amber-100 text-amber-700"}`}>
            {deal.status}
          </span>
        </div>
        <p className="text-sm text-gray-700 mb-2">{deal.message || <span className="text-gray-400">No message</span>}</p>
        <Link href={`/shop/details?id=${deal.listing_id}`} className="text-sm font-semibold text-blue-600 hover:underline">
          View listing
        </Link>
      </div>
      {deal.status === "inquiry" && (
        <Button variant="outline" size="sm" onClick={() => onClose(deal.id)}>
          Close
        </Button>
      )}
    </div>
  );
}

export default function DealsPage() {
  const router = useRouter();
  const [deals, setDeals] = useState<Deal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listMyDeals()
      .then(setDeals)
      .catch((err) => {
        if (friendlyError(err) === "Please sign in to continue.") {
          router.push("/auth/login");
          return;
        }
        setError(friendlyError(err));
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function handleClose(id: string) {
    try {
      await closeDeal(id);
      setDeals((prev) => prev.map((d) => (d.id === id ? { ...d, status: "closed" } : d)));
    } catch (err) {
      setError(friendlyError(err));
    }
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-3xl mx-auto px-6 py-8">
        <h1 className="text-2xl font-extrabold text-gray-900 mb-6">My Deals</h1>

        {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

        {loading ? (
          <div className="py-24 text-center text-gray-400 text-[14px]">Loading…</div>
        ) : deals.length === 0 ? (
          <div className="py-24 text-center text-gray-400 text-[14px]">
            No deals yet. Inquire about a listing to get started.
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {deals.map((d) => (
              <DealCard key={d.id} deal={d} onClose={handleClose} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
