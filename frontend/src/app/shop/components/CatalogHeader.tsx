"use client";

import { LayoutGrid, List } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type SortOption, type ViewMode } from "@/types";

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: "relevance",  label: "Relevance"       },
  { value: "price-asc",  label: "Price: Low → High" },
  { value: "price-desc", label: "Price: High → Low" },
  { value: "date-desc",  label: "Newest first"    },
  { value: "date-asc",   label: "Oldest first"    },
  { value: "hours-asc",  label: "Fewest hours"    },
  { value: "hours-desc", label: "Most hours"      },
];

interface CatalogHeaderProps {
  total:         number;
  query:         string;
  sort:          SortOption;
  view:          ViewMode;
  onSortChange:  (s: SortOption) => void;
  onViewChange:  (v: ViewMode)   => void;
}

export function CatalogHeader({
  total,
  query,
  sort,
  view,
  onSortChange,
  onViewChange,
}: CatalogHeaderProps) {
  return (
    <div className="mb-4">
      {/* Title row */}
      <div className="flex items-start justify-between">
        <div>
          <h1
            className="text-[28px] font-extrabold text-gray-900 leading-tight"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            {total} Excavators for Sale
          </h1>
          <p className="text-[13px] text-gray-500 mt-0.5">
            Showing search results for &ldquo;{query}&rdquo;
          </p>
        </div>

        {/* Sort + view toggle */}
        <div className="flex items-center gap-3 mt-1">
          <div className="flex items-center gap-2">
            <span className="text-[13px] text-gray-500 whitespace-nowrap">
              Sort by:
            </span>
            <Select value={sort} onValueChange={(v) => onSortChange(v as SortOption)}>
              <SelectTrigger className="h-9 text-[13px] w-[160px] border-gray-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SORT_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value} className="text-[13px]">
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* View toggle */}
          <div className="flex items-center border border-gray-200 rounded-lg overflow-hidden">
            <button
              onClick={() => onViewChange("grid")}
              className={`p-2 transition-colors cursor-pointer border-none ${
                view === "grid"
                  ? "bg-gray-100 text-gray-900"
                  : "bg-white text-gray-400 hover:bg-gray-50"
              }`}
            >
              <LayoutGrid size={16} />
            </button>
            <button
              onClick={() => onViewChange("list")}
              className={`p-2 transition-colors cursor-pointer border-none border-l border-gray-200 ${
                view === "list"
                  ? "bg-gray-100 text-gray-900"
                  : "bg-white text-gray-400 hover:bg-gray-50"
              }`}
            >
              <List size={16} />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
