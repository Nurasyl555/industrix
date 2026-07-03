"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { Mail, MapPin, ShieldCheck, Tag } from "lucide-react";
import { getListing, type ListingView } from "@/lib/listing";
import { createDeal } from "@/lib/deal";
import { friendlyError } from "@/lib/api";
import { WishlistButton } from "../components/WishlistButton";

function formatPrice(item: ListingView) {
  const price = "$" + item.price.toLocaleString("en-US");
  if (item.listing_type === "rental" && item.price_period) {
    return `${price} / ${item.price_period}`;
  }
  return price;
}

function DetailsContent() {
  const params = useSearchParams();
  const router = useRouter();
  const id = params.get("id") ?? "";

  const [listing, setListing] = useState<ListingView | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const [message, setMessage] = useState("");
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [dealError, setDealError] = useState("");

  useEffect(() => {
    if (!id) {
      setError("No listing specified.");
      setLoading(false);
      return;
    }
    getListing(id)
      .then(setListing)
      .catch((err) => setError(friendlyError(err)))
      .finally(() => setLoading(false));
  }, [id]);

  async function handleContactSeller() {
    setDealError("");
    setSending(true);
    try {
      await createDeal(id, message);
      setSent(true);
    } catch (err) {
      if (friendlyError(err) === "Please sign in to continue.") {
        router.push(`/auth/login`);
        return;
      }
      setDealError(friendlyError(err));
    } finally {
      setSending(false);
    }
  }

  if (loading) {
    return <main className="min-h-screen bg-slate-50 flex items-center justify-center text-slate-400">Loading…</main>;
  }

  if (error || !listing) {
    return (
      <main className="min-h-screen bg-slate-50 flex items-center justify-center text-slate-500">
        {error || "Listing not found."}
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="mx-auto max-w-7xl px-4 pb-16 pt-6 sm:px-6 lg:px-8">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px] xl:grid-cols-[minmax(0,1fr)_380px]">

          {/* ── LEFT COLUMN ── */}
          <section className="space-y-6">
            <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="relative aspect-video w-full bg-linear-to-br from-sky-300 via-sky-500 to-sky-700">
                <img src="/pics/sample.jpg" alt={listing.title} className="h-full w-full object-cover" />
                <div className="absolute bottom-4 left-4 flex flex-wrap gap-2">
                  <span className="rounded-md bg-sky-600 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
                    {listing.listing_type === "rental" ? "For Rent" : "For Sale"}
                  </span>
                  <span className="rounded-md bg-black/70 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
                    {listing.condition}
                  </span>
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
                {listing.title}
              </h1>
              <WishlistButton
                item={listing}
                size={18}
                className="h-9 w-9 border border-slate-200 shrink-0"
              />
            </div>

            <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-slate-500">
              {listing.region && (
                <span className="flex items-center gap-1.5">
                  <MapPin size={14} /> {listing.region}
                </span>
              )}
              <span className="flex items-center gap-1.5">
                <Tag size={14} /> {listing.condition === "new" ? "New" : "Used"}
              </span>
            </div>

            {listing.description && (
              <section className="space-y-4">
                <h2 className="text-2xl font-semibold text-slate-900">Description</h2>
                <p className="max-w-4xl text-base leading-8 text-slate-600">{listing.description}</p>
              </section>
            )}
          </section>

          {/* ── RIGHT COLUMN ── */}
          <aside className="space-y-4">
            <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
              <div className="space-y-4">
                <h2 className="text-4xl font-bold tracking-tight text-slate-900">{formatPrice(listing)}</h2>

                {sent ? (
                  <div className="rounded-xl bg-emerald-50 border border-emerald-200 px-4 py-3 text-sm text-emerald-800">
                    Your message was sent to the seller.
                  </div>
                ) : (
                  <div className="space-y-3">
                    <textarea
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      placeholder="Introduce yourself and ask about this listing…"
                      rows={3}
                      className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
                    />
                    {dealError && <p className="text-sm text-rose-600">{dealError}</p>}
                    <button
                      onClick={handleContactSeller}
                      disabled={sending}
                      className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-sky-600 text-sm font-semibold text-white transition hover:bg-sky-700 disabled:opacity-60"
                    >
                      <Mail size={16} /> {sending ? "Sending…" : "Contact Seller"}
                    </button>
                  </div>
                )}
              </div>
            </section>

            <section className="rounded-2xl border border-amber-200 bg-amber-50 p-4 shadow-sm">
              <div className="flex items-start gap-3">
                <ShieldCheck size={18} className="mt-0.5 shrink-0 text-amber-500" />
                <div>
                  <h3 className="text-sm font-semibold text-amber-900">Trade Safely</h3>
                  <p className="mt-1 text-sm leading-6 text-amber-800/80">
                    Always inspect machinery in person before making any payment. Industrix does not currently
                    process payments on your behalf — arrange delivery and payment directly with the seller.
                  </p>
                </div>
              </div>
            </section>
          </aside>

        </div>
      </div>
    </main>
  );
}

export default function DetailsPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading…</div>}>
      <DetailsContent />
    </Suspense>
  );
}
