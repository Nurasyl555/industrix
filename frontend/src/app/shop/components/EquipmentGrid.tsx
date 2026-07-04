"use client";

import { MapPin, Tag } from "lucide-react";
import { Button } from "@/components/ui/button";
import Image from "next/image";
import Link from "next/link";
import { type ViewMode } from "@/types";
import { type ListingView } from "@/lib/listing";
import { WishlistButton } from "./WishlistButton";

function formatPrice(item: ListingView) {
  const price = "$" + item.price.toLocaleString("en-US");
  if (item.listing_type === "rental" && item.price_period) {
    return `${price} / ${item.price_period}`;
  }
  return price;
}

function ConditionBadge({ condition }: { condition: ListingView["condition"] }) {
  const isNew = condition === "new";
  return (
    <span
      className={`absolute top-3 left-3 text-[10px] font-bold px-2 py-0.5 rounded uppercase ${
        isNew ? "bg-blue-600 text-white" : "bg-gray-700 text-white"
      }`}
    >
      {condition}
    </span>
  );
}

/* ─── GRID CARD ─────────────────────────────────────────────── */
function GridCard({ item }: { item: ListingView }) {
  return (
    <Link
      href={`/shop/details?id=${item.id}`}
      className="border border-gray-200 rounded-xl overflow-hidden bg-white hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 flex flex-col no-underline text-inherit"
    >
      <div className="relative h-46.25 overflow-hidden bg-sky-50">
        <Image src={item.image_url || "/pics/sample.jpg"} alt={item.title} fill className="object-cover" unoptimized />
        <ConditionBadge condition={item.condition} />
        <WishlistButton item={item} size={14} className="absolute top-3 right-3 w-8 h-8" />
      </div>

      <div className="p-4 flex flex-col flex-1">
        <div className="flex items-start justify-between gap-2 mb-3">
          <h3
            className="text-[14px] font-bold text-gray-900 leading-snug"
            style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
          >
            {item.title}
          </h3>
          <span className="text-[16px] font-extrabold text-blue-600 shrink-0" style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}>
            {formatPrice(item)}
          </span>
        </div>

        <div className="flex flex-wrap gap-x-4 gap-y-1.5 text-[12px] text-gray-500 mb-3">
          <span className="flex items-center gap-1.5">
            <Tag size={12} className="text-gray-400 shrink-0" />
            {item.listing_type === "rental" ? "For Rent" : "For Sale"}
          </span>
          {item.region && (
            <span className="flex items-center gap-1.5">
              <MapPin size={12} className="text-gray-400 shrink-0" />
              {item.region}
            </span>
          )}
        </div>

        <Button
          asChild
          variant="outline"
          className="w-full h-9 text-[13px] font-semibold border-gray-200 hover:border-gray-400 mt-auto"
        >
          <span>View Details</span>
        </Button>
      </div>
    </Link>
  );
}

/* ─── LIST CARD ──────────────────────────────────────────────── */
function ListCard({ item }: { item: ListingView }) {
  return (
    <Link
      href={`/shop/details?id=${item.id}`}
      className="border border-gray-200 rounded-xl overflow-hidden bg-white hover:shadow-md transition-all duration-200 flex no-underline text-inherit"
    >
      <div className="relative w-55 shrink-0 overflow-hidden bg-sky-50">
        <Image src={item.image_url || "/pics/sample.jpg"} alt={item.title} fill className="object-cover" unoptimized />
        <ConditionBadge condition={item.condition} />
        <WishlistButton item={item} size={13} className="absolute top-3 right-3 w-7 h-7" />
      </div>

      <div className="p-4 flex flex-col flex-1 min-w-0">
        <div className="flex items-start justify-between gap-4 mb-2">
          <h3
            className="text-[15px] font-bold text-gray-900 leading-snug"
            style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
          >
            {item.title}
          </h3>
          <span
            className="text-[18px] font-extrabold text-blue-600 shrink-0"
            style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
          >
            {formatPrice(item)}
          </span>
        </div>

        <div className="flex flex-wrap gap-x-5 gap-y-1 text-[12px] text-gray-500 mb-2">
          <span className="flex items-center gap-1.5">
            <Tag size={12} className="text-gray-400" />
            {item.listing_type === "rental" ? "For Rent" : "For Sale"}
          </span>
          {item.region && (
            <span className="flex items-center gap-1.5">
              <MapPin size={12} className="text-gray-400" />
              {item.region}
            </span>
          )}
        </div>

        <div className="mt-auto">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="h-8 text-[12px] font-semibold border-gray-200 hover:border-gray-400 px-5"
          >
            <span>View Details</span>
          </Button>
        </div>
      </div>
    </Link>
  );
}

/* ─── GRID WRAPPER ───────────────────────────────────────────── */
interface EquipmentGridProps {
  view:  ViewMode;
  items: ListingView[];
}

export function EquipmentGrid({ view, items }: EquipmentGridProps) {
  if (items.length === 0) {
    return (
      <div className="py-24 text-center text-gray-400 text-[14px]">
        No equipment matches your filters.
      </div>
    );
  }

  if (view === "grid") {
    return (
      <div className="grid grid-cols-3 gap-5 mb-8">
        {items.map((item) => (
          <GridCard key={item.id} item={item} />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 mb-8">
      {items.map((item) => (
        <ListCard key={item.id} item={item} />
      ))}
    </div>
  );
}
