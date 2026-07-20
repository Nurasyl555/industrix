"use client";

// src/app/shop/deals/page.tsx
// List of the user's deals (as buyer or seller). Each opens a full
// conversation at /shop/deals/[id].

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { MessageSquare } from "lucide-react";
import { listMyDeals, type Deal } from "@/lib/deal";
import { friendlyError } from "@/lib/api";
import { useI18n, type TranslationKey } from "@/lib/i18n";

/** Maps a backend deal status onto its dictionary key. */
function dealStatusKey(status: string): TranslationKey {
  const key = `dealStatus.${status}` as TranslationKey;
  return key;
}

function DealRow({ deal }: { deal: Deal }) {
  const { t } = useI18n();
  return (
    <Link
      href={`/shop/deals/${deal.id}`}
      className="flex items-center justify-between gap-4 rounded-xl border border-gray-200 p-4 no-underline transition-colors hover:border-blue-400 hover:bg-blue-50/30"
    >
      <div className="min-w-0">
        <div className="mb-1 flex items-center gap-2">
          <span className={`rounded px-2 py-0.5 text-[10px] font-bold uppercase ${deal.role === "buyer" ? "bg-blue-100 text-blue-700" : "bg-emerald-100 text-emerald-700"}`}>
            {deal.role === "buyer" ? t("deals.youInquired") : t("deals.inquiryReceived")}
          </span>
          {/* Terminal statuses are greyed; anything still in flight stays amber. */}
          <span className={`rounded px-2 py-0.5 text-[10px] font-bold uppercase ${deal.status === "completed" || deal.status === "cancelled" ? "bg-gray-100 text-gray-500" : "bg-amber-100 text-amber-700"}`}>
            {t(dealStatusKey(deal.status))}
          </span>
        </div>
        <p className="truncate text-sm text-gray-700">
          {deal.message || <span className="text-gray-400">{t("deals.noMessage")}</span>}
        </p>
      </div>
      <MessageSquare size={18} className="shrink-0 text-gray-400" />
    </Link>
  );
}

export default function DealsPage() {
  const { t } = useI18n();
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

  return (
    <div className="min-h-screen bg-white">
      <div className="mx-auto max-w-3xl px-6 py-8">
        <h1 className="mb-6 text-2xl font-extrabold text-gray-900">{t("deals.title")}</h1>

        {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

        {loading ? (
          <div className="py-24 text-center text-[14px] text-gray-400">{t("common.loading")}</div>
        ) : deals.length === 0 ? (
          <div className="py-24 text-center text-[14px] text-gray-400">
            {t("deals.empty")}
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {deals.map((d) => (
              <DealRow key={d.id} deal={d} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
