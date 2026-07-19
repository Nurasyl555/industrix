"use client";

import { X } from "lucide-react";
import { type Filters } from "@/types";
import { useI18n } from "@/lib/i18n";

interface ActiveFiltersProps {
  filters:      Filters;
  categoryName: string;
  onRemove:     (key: keyof Filters) => void;
  onClearAll:   () => void;
}

export function ActiveFilters({ filters, categoryName, onRemove, onClearAll }: ActiveFiltersProps) {
  const { t } = useI18n();
  const chips: { label: string; key: keyof Filters }[] = [];

  if (filters.categoryId) chips.push({ label: `${t("filters.category")}: ${categoryName}`, key: "categoryId" });
  if (filters.region) chips.push({ label: `${t("filters.region")}: ${filters.region}`, key: "region" });
  if (filters.condition) {
    chips.push({
      label: `${t("filters.condition")}: ${filters.condition === "new" ? t("condition.new") : t("condition.used")}`,
      key: "condition",
    });
  }
  if (filters.listingType) {
    chips.push({
      label: filters.listingType === "rental" ? t("listingType.rental") : t("listingType.sale"),
      key: "listingType",
    });
  }
  if (filters.priceMin) chips.push({ label: `${t("filters.min")} ${filters.priceMin} ₸`, key: "priceMin" });
  if (filters.priceMax) chips.push({ label: `${t("filters.max")} ${filters.priceMax} ₸`, key: "priceMax" });

  if (chips.length === 0) return null;

  return (
    <div className="flex items-center flex-wrap gap-2 mb-5">
      {chips.map((chip) => (
        <span
          key={chip.key}
          className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full border border-gray-300 bg-white text-[12px] font-medium text-gray-700"
        >
          {chip.label}
          <button
            onClick={() => onRemove(chip.key)}
            aria-label={t("filters.reset")}
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
        {t("filters.clearAll")}
      </button>
    </div>
  );
}
