"use client";

// src/app/admin/page.tsx
// Admin moderation console: company verification queue + listing moderation
// queue. Access is enforced server-side (admin-role gated /admin/*); this page
// also redirects non-admins away.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Check, X } from "lucide-react";
import {
  listCompanies,
  verifyCompany,
  rejectCompany,
  listModerationQueue,
  approveListing,
  rejectListing,
} from "@/lib/admin";
import { type Company } from "@/lib/company";
import { type ListingView } from "@/lib/listing";
import { getCurrentUser } from "@/lib/user";
import { friendlyError } from "@/lib/api";
import { Button } from "@/components/ui/button";

export default function AdminPage() {
  const router = useRouter();
  const [tab, setTab] = useState<"companies" | "listings">("companies");
  const [companies, setCompanies] = useState<Company[]>([]);
  const [listings, setListings] = useState<ListingView[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [denied, setDenied] = useState(false);

  async function loadCompanies() {
    setCompanies((await listCompanies("pending")) ?? []);
  }
  async function loadListings() {
    setListings((await listModerationQueue()) ?? []);
  }

  useEffect(() => {
    (async () => {
      const user = await getCurrentUser();
      if (!user) { router.push("/auth/login"); return; }
      try {
        await Promise.all([loadCompanies(), loadListings()]);
      } catch (err) {
        // 403 → not an admin
        if (friendlyError(err) === "Please sign in to continue." || friendlyError(err).includes("Admin")) {
          setDenied(true);
        } else {
          setError(friendlyError(err));
        }
      } finally {
        setLoading(false);
      }
    })();
  }, [router]);

  async function act(fn: () => Promise<void>, reload: () => Promise<void>) {
    try {
      await fn();
      await reload();
    } catch (err) {
      setError(friendlyError(err));
    }
  }

  if (loading) return <div className="py-24 text-center text-gray-400">Loading…</div>;

  if (denied) {
    return (
      <div className="py-24 text-center">
        <p className="text-lg font-semibold text-gray-700">Admin access required</p>
        <p className="mt-1 text-sm text-gray-400">This page is only available to administrators.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <h1 className="mb-6 text-2xl font-extrabold text-gray-900">Admin console</h1>

      {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      <div className="mb-6 flex gap-2">
        <button
          onClick={() => setTab("companies")}
          className={`rounded-full px-4 py-1.5 text-sm font-semibold ${tab === "companies" ? "bg-gray-900 text-white" : "bg-gray-100 text-gray-600"}`}
        >
          Companies ({companies.length})
        </button>
        <button
          onClick={() => setTab("listings")}
          className={`rounded-full px-4 py-1.5 text-sm font-semibold ${tab === "listings" ? "bg-gray-900 text-white" : "bg-gray-100 text-gray-600"}`}
        >
          Listings ({listings.length})
        </button>
      </div>

      {tab === "companies" && (
        <div className="flex flex-col gap-3">
          {companies.length === 0 && <p className="py-16 text-center text-sm text-gray-400">No companies awaiting verification.</p>}
          {companies.map((c) => (
            <div key={c.id} className="flex items-center justify-between gap-4 rounded-xl border border-gray-200 p-4">
              <div className="min-w-0">
                <p className="font-semibold text-gray-900">{c.name}</p>
                <p className="text-sm text-gray-500">BIN {c.bin} · {c.email} · {c.phone}</p>
                <p className="text-xs text-gray-400">{c.address}</p>
              </div>
              <div className="flex shrink-0 gap-2">
                <Button size="sm" onClick={() => act(() => verifyCompany(c.id), loadCompanies)}>
                  <Check size={15} /> Verify
                </Button>
                <Button size="sm" variant="outline" onClick={() => act(() => rejectCompany(c.id, "Rejected by admin"), loadCompanies)}>
                  <X size={15} /> Reject
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {tab === "listings" && (
        <div className="flex flex-col gap-3">
          {listings.length === 0 && <p className="py-16 text-center text-sm text-gray-400">No listings awaiting moderation.</p>}
          {listings.map((l) => (
            <div key={l.id} className="flex items-center justify-between gap-4 rounded-xl border border-gray-200 p-4">
              <div className="min-w-0">
                <Link href={`/shop/details?id=${l.id}`} className="font-semibold text-gray-900 hover:underline">
                  {l.title}
                </Link>
                <p className="text-sm text-gray-500">
                  {l.listing_type === "rental" ? "For Rent" : "For Sale"} · ${l.price.toLocaleString()} · {l.condition} · {l.region || "—"}
                </p>
              </div>
              <div className="flex shrink-0 gap-2">
                <Button size="sm" onClick={() => act(() => approveListing(l.id), loadListings)}>
                  <Check size={15} /> Approve
                </Button>
                <Button size="sm" variant="outline" onClick={() => act(() => rejectListing(l.id), loadListings)}>
                  <X size={15} /> Reject
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
