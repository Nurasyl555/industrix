"use client";

import { useState } from "react";
import { Heart, Calendar, Clock, MapPin } from "lucide-react";
import { Button } from "@/components/ui/button";
import Image from "next/image";
import { ALL_EQUIPMENT, applyFiltersAndSort, paginate } from "../catalog/data"
import { type ViewMode, type SortOption, type Filters, type Equipment } from "@/types";
import Link from "next/link";

const PER_PAGE = 6; // 6 items visible in the Figma screenshots

function formatPrice(n: number) {
  return "$" + n.toLocaleString("en-US");
}

function Badge({ label }: { label: string }) {
  const isVerified = label === "VERIFIED";
  return (
    <span
      className={`absolute top-3 left-3 text-[10px] font-bold px-2 py-0.5 rounded ${
        isVerified
          ? "bg-green-600 text-white"
          : "bg-blue-600 text-white"
      }`}
    >
      {label}
    </span>
  );
}

/* ─── GRID CARD ─────────────────────────────────────────────── */
function GridCard({ item, seed }: { item: Equipment; seed: number }) {
  const [wished, setWished] = useState(false);
  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden bg-white hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 cursor-pointer flex flex-col">
      {/* Image */}
      <div className="relative h-46.25 overflow-hidden bg-sky-50">
        <Image
          src="/pics/sample.jpg"
          alt="Equipment image"
          fill
          className="object-cover"
        />
        {item.badge && <Badge label={item.badge} />}
        <button
          onClick={(e) => { e.stopPropagation(); setWished(!wished); }}
          className="absolute top-3 right-3 w-8 h-8 rounded-full bg-white shadow flex items-center justify-center border-none cursor-pointer hover:bg-red-50 transition-colors"
        >
          <Heart size={14} className={wished ? "fill-red-500 text-red-500" : "text-gray-400"} />
        </button>
      </div>

      {/* Body */}
      <div className="p-4 flex flex-col flex-1">
        <div className="flex items-start justify-between gap-2 mb-3">
          <h3
            className="text-[14px] font-bold text-gray-900 leading-snug"
            style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
          >
            {item.title}
          </h3>
          <span className="text-[16px] font-extrabold text-blue-600 shrink-0" style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}>
            {formatPrice(item.price)}
          </span>
        </div>

        <div className="flex flex-wrap gap-x-4 gap-y-1.5 text-[12px] text-gray-500 mb-3">
          <span className="flex items-center gap-1.5">
            <Calendar size={12} className="text-gray-400 shrink-0" />
            Year: {item.year}
          </span>
          <span className="flex items-center gap-1.5">
            <Clock size={12} className="text-gray-400 shrink-0" />
            Hours: {item.hours.toLocaleString()}
          </span>
        </div>

        <div className="flex items-center gap-1.5 text-[12px] text-gray-500 mb-4">
          <MapPin size={12} className="text-gray-400 shrink-0" />
          {item.location}
        </div>

        <Button
          asChild
          variant="outline"
          className="w-full h-9 text-[13px] font-semibold border-gray-200 hover:border-gray-400 mt-auto"
        >
          <Link href="/shop/details">
            View Details
          </Link>
        </Button>
      </div>
    </div>
  );
}

/* ─── LIST CARD ──────────────────────────────────────────────── */
function ListCard({ item, seed }: { item: Equipment; seed: number }) {
  const [wished, setWished] = useState(false);
  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden bg-white hover:shadow-md transition-all duration-200 cursor-pointer flex">
      {/* Image */}
      <div className="relative w-55 shrink-0 overflow-hidden bg-sky-50">
        <Image
          src="/pics/sample.jpg"
          alt="Equipment image"
          fill
          className="object-cover"
        />
        {item.badge && <Badge label={item.badge} />}
        <button
          onClick={(e) => { e.stopPropagation(); setWished(!wished); }}
          className="absolute top-3 right-3 w-7 h-7 rounded-full bg-white shadow flex items-center justify-center border-none cursor-pointer hover:bg-red-50 transition-colors"
        >
          <Heart size={13} className={wished ? "fill-red-500 text-red-500" : "text-gray-400"} />
        </button>
      </div>

      {/* Body */}
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
            {formatPrice(item.price)}
          </span>
        </div>

        <div className="flex flex-wrap gap-x-5 gap-y-1 text-[12px] text-gray-500 mb-2">
          <span className="flex items-center gap-1.5">
            <Calendar size={12} className="text-gray-400" />
            Year: {item.year}
          </span>
          <span className="flex items-center gap-1.5">
            <Clock size={12} className="text-gray-400" />
            Hours: {item.hours.toLocaleString()}
          </span>
          <span className="flex items-center gap-1.5">
            <MapPin size={12} className="text-gray-400" />
            {item.location}
          </span>
        </div>

        <div className="mt-auto">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="h-8 text-[12px] font-semibold border-gray-200 hover:border-gray-400 px-5"
          >
            <Link href="/shop/details">
              View Details
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}

/* ─── GRID WRAPPER ───────────────────────────────────────────── */
interface EquipmentGridProps {
  view:    ViewMode;
  page:    number;
  sort:    SortOption;
  filters: Filters;
}

export function EquipmentGrid({ view, page, sort, filters }: EquipmentGridProps) {
  const filtered = applyFiltersAndSort(ALL_EQUIPMENT, filters, sort);
  const items    = paginate(filtered, page, PER_PAGE);

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
        {items.map((item, i) => (
          <GridCard key={item.id} item={item} seed={Number(item.image)} />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 mb-8">
      {items.map((item) => (
        <ListCard key={item.id} item={item} seed={Number(item.image)} />
      ))}
    </div>
  );
}
