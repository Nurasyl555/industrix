"use client";

import { X } from "lucide-react";
import { type Filters } from "@/types";

interface ActiveFiltersProps {
  filters:     Filters;
  categoryName: string;
  onRemove:    (key: keyof Filters) => void;
  onClearAll:  () => void;
}

export function ActiveFilters({ filters, categoryName, onRemove, onClearAll }: ActiveFiltersProps) {
  const chips: { label: string; key: keyof Filters }[] = [];

  if (filters.categoryId) chips.push({ label: `Category: ${categoryName}`, key: "categoryId" });
  if (filters.condition) chips.push({ label: `Condition: ${filters.condition}`, key: "condition" });
  if (filters.listingType) chips.push({ label: filters.listingType === "rental" ? "For Rent" : "For Sale", key: "listingType" });
  if (filters.priceMin) chips.push({ label: `Min $${filters.priceMin}`, key: "priceMin" });
  if (filters.priceMax) chips.push({ label: `Max $${filters.priceMax}`, key: "priceMax" });

  if (chips.length === 0) return null;

  return (
    <div className="flex items-center flex-wrap gap-2 mb-5">
      <span className="text-[13px] text-gray-500 mr-1">Active Filters:</span>

      {chips.map((chip) => (
        <span
          key={chip.key}
          className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border border-gray-300 bg-white text-[12px] font-medium text-gray-700"
        >
          {chip.label}
          <button
            onClick={() => onRemove(chip.key)}
            className="text-gray-400 hover:text-gray-700 transition-colors bg-transparent border-none cursor-pointer p-0 flex items-center"
          >
            <X size={12} />
          </button>
        </span>
      ))}

      <button
        onClick={onClearAll}
        className="text-[12px] font-semibold text-blue-600 hover:text-blue-700 bg-transparent border-none cursor-pointer p-0 ml-1"
      >
        Clear All
      </button>
    </div>
  );
}
