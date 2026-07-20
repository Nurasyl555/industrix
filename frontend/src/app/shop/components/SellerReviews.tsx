"use client";

import { useEffect, useState } from "react";
import { Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  listReviews,
  getReputation,
  postReview,
  type Review,
  type Reputation,
} from "@/lib/review";
import { getCurrentUser } from "@/lib/user";
import { useI18n } from "@/lib/i18n";
import { friendlyError } from "@/lib/api";

const TIER_STYLE: Record<string, string> = {
  gold: "bg-amber-100 text-amber-700",
  silver: "bg-gray-200 text-gray-700",
  bronze: "bg-orange-100 text-orange-700",
  none: "bg-gray-100 text-gray-400",
};

// Row of 5 stars. `value` is filled count; when `onPick` is set it's interactive.
function Stars({
  value,
  size = 16,
  onPick,
}: {
  value: number;
  size?: number;
  onPick?: (n: number) => void;
}) {
  const [hover, setHover] = useState(0);
  return (
    <div className="flex items-center gap-0.5">
      {[1, 2, 3, 4, 5].map((n) => {
        const filled = (hover || value) >= n;
        return (
          <button
            key={n}
            type="button"
            disabled={!onPick}
            onClick={() => onPick?.(n)}
            onMouseEnter={() => onPick && setHover(n)}
            onMouseLeave={() => onPick && setHover(0)}
            className={onPick ? "cursor-pointer" : "cursor-default"}
          >
            <Star
              size={size}
              className={filled ? "fill-amber-400 text-amber-400" : "text-gray-300"}
            />
          </button>
        );
      })}
    </div>
  );
}

export function SellerReviews({ sellerId }: { sellerId: string }) {
  const { t } = useI18n();
  const [reputation, setReputation] = useState<Reputation | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [isSeller, setIsSeller] = useState(false);
  const [signedIn, setSignedIn] = useState(false);

  const [rating, setRating] = useState(0);
  const [comment, setComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function refresh() {
    const [rep, revs] = await Promise.all([getReputation(sellerId), listReviews(sellerId)]);
    setReputation(rep);
    setReviews(revs);
  }

  useEffect(() => {
    refresh();
    getCurrentUser().then((u) => {
      setSignedIn(!!u);
      setIsSeller(!!u && u.id === sellerId);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sellerId]);

  async function handleSubmit() {
    if (rating < 1) { setError(t("reviews.pickRating")); return; }
    setSubmitting(true);
    setError("");
    try {
      await postReview(sellerId, rating, comment);
      setRating(0);
      setComment("");
      await refresh();
    } catch (err) {
      setError(friendlyError(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="space-y-5">
      <h2 className="text-2xl font-semibold text-slate-900">{t("reviews.title")}</h2>

      {/* Reputation summary */}
      <div className="flex items-center gap-4 rounded-2xl border border-slate-200 bg-white p-5">
        {reputation && reputation.review_count > 0 ? (
          <>
            <div className="text-center">
              <div className="text-3xl font-bold text-slate-900">{reputation.average_rating.toFixed(1)}</div>
              <Stars value={Math.round(reputation.average_rating)} />
              <div className="mt-1 text-xs text-slate-500">{reputation.review_count} review(s)</div>
            </div>
            <span className={`rounded-full px-3 py-1 text-xs font-bold uppercase ${TIER_STYLE[reputation.tier]}`}>
              {reputation.tier === "none" ? "unrated" : reputation.tier}
            </span>
          </>
        ) : (
          <p className="text-sm text-slate-400">No ratings yet — be the first to review this seller.</p>
        )}
      </div>

      {/* {t("reviews.leave")} */}
      {signedIn && !isSeller && (
        <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-5">
          <p className="text-sm font-semibold text-slate-800">{t("reviews.leave")}</p>
          <Stars value={rating} size={22} onPick={setRating} />
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={3}
            placeholder={t("reviews.placeholder")}
            className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500"
          />
          {error && <p className="text-sm text-rose-600">{error}</p>}
          <Button onClick={handleSubmit} disabled={submitting}>
            {submitting ? t("reviews.submitting") : t("reviews.submit")}
          </Button>
        </div>
      )}

      {/* Reviews list */}
      {reviews.length > 0 && (
        <div className="space-y-3">
          {reviews.map((r) => (
            <div key={r.id} className="rounded-2xl border border-slate-200 bg-white p-4">
              <div className="mb-1 flex items-center justify-between">
                <Stars value={r.rating} size={14} />
                <span className="text-xs text-slate-400">
                  {new Date(r.created_at).toLocaleDateString()}
                </span>
              </div>
              {r.comment && <p className="text-sm text-slate-700">{r.comment}</p>}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
