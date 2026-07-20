"use client";

import { useEffect, useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  getBookedDates,
  createBooking,
  type DateRange,
} from "@/lib/booking";
import { getCurrentUser } from "@/lib/user";
import { useI18n } from "@/lib/i18n";
import { friendlyError } from "@/lib/api";
import { type ListingView } from "@/lib/listing";

// ── date helpers (work in local time, day-granularity) ──────────────────────
const iso = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
const parseISO = (s: string) => {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
};
const addDays = (d: Date, n: number) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + n);
const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate());

// Every ISO day covered by the booked ranges (inclusive of both ends).
function bookedDaySet(ranges: DateRange[]): Set<string> {
  const set = new Set<string>();
  for (const r of ranges) {
    let cur = parseISO(r.start_date);
    const end = parseISO(r.end_date);
    while (cur <= end) {
      set.add(iso(cur));
      cur = addDays(cur, 1);
    }
  }
  return set;
}

function daysInclusive(a: Date, b: Date) {
  return Math.round((startOfDay(b).getTime() - startOfDay(a).getTime()) / 86400000) + 1;
}

const PERIOD_DAYS: Record<string, number> = { day: 1, week: 7, month: 30 };

export function RentalCalendar({ listing }: { listing: ListingView }) {
  const { t } = useI18n();
  const today = startOfDay(new Date());
  const [view, setView] = useState(() => new Date(today.getFullYear(), today.getMonth(), 1));
  const [booked, setBooked] = useState<Set<string>>(new Set());
  const [start, setStart] = useState<Date | null>(null);
  const [end, setEnd] = useState<Date | null>(null);
  const [signedIn, setSignedIn] = useState(false);
  const [isOwner, setIsOwner] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  async function refresh() {
    setBooked(bookedDaySet(await getBookedDates(listing.id)));
  }

  useEffect(() => {
    refresh();
    getCurrentUser().then((u) => {
      setSignedIn(!!u);
      setIsOwner(!!u && u.id === listing.seller_id);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [listing.id]);

  // Days rendered for the current month grid (leading blanks for alignment).
  const cells = useMemo(() => {
    const first = new Date(view.getFullYear(), view.getMonth(), 1);
    const startWeekday = (first.getDay() + 6) % 7; // Mon=0
    const daysInMonth = new Date(view.getFullYear(), view.getMonth() + 1, 0).getDate();
    const arr: (Date | null)[] = [];
    for (let i = 0; i < startWeekday; i++) arr.push(null);
    for (let d = 1; d <= daysInMonth; d++) arr.push(new Date(view.getFullYear(), view.getMonth(), d));
    return arr;
  }, [view]);

  // Does the [a,b] range contain any booked day?
  function rangeHasBooked(a: Date, b: Date) {
    let cur = a;
    while (cur <= b) {
      if (booked.has(iso(cur))) return true;
      cur = addDays(cur, 1);
    }
    return false;
  }

  function onPick(day: Date) {
    if (day < today || booked.has(iso(day))) return;
    setError("");
    setSuccess(false);
    // No start yet, or both set → begin a new selection.
    if (!start || (start && end)) {
      setStart(day);
      setEnd(null);
      return;
    }
    // Second click.
    if (day < start) {
      setStart(day);
      return;
    }
    if (rangeHasBooked(start, day)) {
      setError("Your range overlaps dates that are already booked.");
      return;
    }
    setEnd(day);
  }

  function inSelection(day: Date) {
    if (start && end) return day >= start && day <= end;
    if (start) return iso(day) === iso(start);
    return false;
  }

  const numDays = start && end ? daysInclusive(start, end) : 0;
  const total = useMemo(() => {
    if (!numDays || !listing.price_period) return 0;
    const perDay = listing.price / (PERIOD_DAYS[listing.price_period] ?? 1);
    return Math.round(perDay * numDays);
  }, [numDays, listing.price, listing.price_period]);

  async function handleBook() {
    if (!start || !end) return;
    setSubmitting(true);
    setError("");
    try {
      await createBooking(listing.id, iso(start), iso(end));
      setSuccess(true);
      setStart(null);
      setEnd(null);
      await refresh();
    } catch (err) {
      setError(friendlyError(err));
    } finally {
      setSubmitting(false);
    }
  }

  const monthLabel = view.toLocaleDateString("en-US", { month: "long", year: "numeric" });
  const canGoPrev = view > new Date(today.getFullYear(), today.getMonth(), 1);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-900">{t("rental.availability")}</h3>
        <div className="flex items-center gap-3 text-xs">
          <span className="flex items-center gap-1 text-rose-500">
            <span className="h-2 w-2 rounded-full bg-rose-400" /> Booked
          </span>
          <span className="flex items-center gap-1 text-sky-600">
            <span className="h-2 w-2 rounded-full bg-sky-500" /> Selected
          </span>
        </div>
      </div>

      <div className="rounded-2xl bg-slate-50 p-4">
        {/* Month nav */}
        <div className="mb-3 flex items-center justify-between">
          <button
            disabled={!canGoPrev}
            onClick={() => setView(new Date(view.getFullYear(), view.getMonth() - 1, 1))}
            className="flex h-7 w-7 items-center justify-center rounded-full text-slate-500 hover:bg-slate-200 disabled:opacity-30"
          >
            <ChevronLeft size={16} />
          </button>
          <span className="text-sm font-semibold text-slate-800">{monthLabel}</span>
          <button
            onClick={() => setView(new Date(view.getFullYear(), view.getMonth() + 1, 1))}
            className="flex h-7 w-7 items-center justify-center rounded-full text-slate-500 hover:bg-slate-200"
          >
            <ChevronRight size={16} />
          </button>
        </div>

        <div className="mb-2 grid grid-cols-7 text-center text-xs font-medium uppercase text-slate-400">
          {["M", "T", "W", "T", "F", "S", "S"].map((d, i) => <span key={i}>{d}</span>)}
        </div>

        <div className="grid grid-cols-7 gap-1.5">
          {cells.map((day, i) => {
            if (!day) return <div key={i} />;
            const isBooked = booked.has(iso(day));
            const isPast = day < today;
            const selected = inSelection(day);
            const disabled = isBooked || isPast;
            return (
              <button
                key={i}
                disabled={disabled}
                onClick={() => onPick(day)}
                className={[
                  "flex aspect-square items-center justify-center rounded-lg text-sm font-medium transition",
                  isBooked && "bg-rose-50 text-rose-400 line-through cursor-not-allowed",
                  isPast && !isBooked && "text-slate-300 cursor-not-allowed",
                  selected && "bg-sky-600 text-white",
                  !disabled && !selected && "text-slate-700 hover:bg-sky-100",
                ].filter(Boolean).join(" ")}
              >
                {day.getDate()}
              </button>
            );
          })}
        </div>
      </div>

      {/* Summary + action */}
      {success && (
        <div className="rounded-xl bg-emerald-50 border border-emerald-200 px-4 py-3 text-sm text-emerald-800">
          Booked! See it under My Bookings.
        </div>
      )}
      {error && <p className="text-sm text-rose-600">{error}</p>}

      {start && end && (
        <div className="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-3 text-sm">
          <div>
            <div className="font-semibold text-slate-800">{iso(start)} → {iso(end)}</div>
            <div className="text-slate-500">{numDays} day(s){total ? ` · ~$${total.toLocaleString()}` : ""}</div>
          </div>
        </div>
      )}

      {isOwner ? (
        <p className="text-center text-xs text-slate-400">{t("rental.ownListing")}</p>
      ) : signedIn ? (
        <Button className="w-full" disabled={!start || !end || submitting} onClick={handleBook}>
          {submitting ? t("rental.booking") : start && end ? t("rental.book") : t("rental.selectDates")}
        </Button>
      ) : (
        <p className="text-center text-xs text-slate-400">{t("rental.signInToBook")}</p>
      )}
    </div>
  );
}
