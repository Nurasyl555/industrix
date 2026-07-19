"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { EMPTY_FILTERS, type Filters } from "@/types";
import { type Category } from "@/lib/catalog";
import { type SearchResult } from "@/lib/search";
import { useI18n, type TranslationKey } from "@/lib/i18n";

const CONDITIONS: { value: "new" | "used"; labelKey: TranslationKey }[] = [
  { value: "new", labelKey: "condition.new" },
  { value: "used", labelKey: "condition.used" },
];

const LISTING_TYPES: { value: "sale" | "rental"; labelKey: TranslationKey }[] = [
  { value: "sale", labelKey: "listingType.sale" },
  { value: "rental", labelKey: "listingType.rental" },
];

interface FilterSidebarProps {
  filters: Filters;
  categories: Category[];
  /** Per-value result counts from the search response. */
  facets: SearchResult["facets"];
  onChange: (f: Filters) => void;
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(true);
  return (
    <div className="border-b border-gray-100 pb-4 mb-4 last:border-none last:mb-0">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center justify-between w-full text-left bg-transparent border-none cursor-pointer p-0 mb-3"
      >
        <span className="text-[13px] font-bold text-gray-900">{title}</span>
        {open ? <ChevronUp size={15} className="text-gray-400" /> : <ChevronDown size={15} className="text-gray-400" />}
      </button>
      {open && children}
    </div>
  );
}

/** A checkbox row showing how many results the option would yield. */
function FacetOption({
  label,
  count,
  checked,
  onToggle,
}: {
  label: string;
  count?: number;
  checked: boolean;
  onToggle: () => void;
}) {
  // An option with no matches is greyed out rather than hidden, so the list
  // doesn't reshuffle every time a filter changes.
  const empty = count === undefined || count === 0;
  return (
    <label className={`flex items-center gap-2.5 ${empty && !checked ? "opacity-45" : "cursor-pointer"}`}>
      <Checkbox
        checked={checked}
        onCheckedChange={onToggle}
        className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
      />
      <span className="text-[13px] text-gray-700 flex-1">{label}</span>
      {count !== undefined && <span className="text-[11px] text-gray-400 tabular-nums">{count}</span>}
    </label>
  );
}

export function FilterSidebar({ filters, categories, facets, onChange }: FilterSidebarProps) {
  const { t } = useI18n();

  const toggle = <K extends keyof Filters>(key: K, value: Filters[K]) =>
    onChange({ ...filters, [key]: filters[key] === value ? ("" as Filters[K]) : value });

  // Regions aren't a fixed list — they're whatever is actually indexed.
  const regions = Object.keys(facets.region ?? {}).sort();

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <span className="text-[15px] font-bold text-gray-900">{t("filters.title")}</span>
        <button
          onClick={() => onChange(EMPTY_FILTERS)}
          className="text-[13px] font-semibold text-blue-600 hover:text-blue-700 bg-transparent border-none cursor-pointer p-0"
        >
          {t("filters.reset")}
        </button>
      </div>

      <Section title={t("filters.category")}>
        <div className="flex flex-col gap-2.5">
          {categories.map((cat) => (
            <FacetOption
              key={cat.id}
              label={cat.name}
              count={facets.category_id?.[cat.id]}
              checked={filters.categoryId === cat.id}
              onToggle={() => toggle("categoryId", cat.id)}
            />
          ))}
        </div>
      </Section>

      {regions.length > 0 && (
        <Section title={t("filters.region")}>
          <div className="flex flex-col gap-2.5">
            {regions.map((region) => (
              <FacetOption
                key={region}
                label={region}
                count={facets.region?.[region]}
                checked={filters.region === region}
                onToggle={() => toggle("region", region)}
              />
            ))}
          </div>
        </Section>
      )}

      <Section title={t("filters.priceRange")}>
        <div className="flex gap-2">
          <Input
            placeholder={t("filters.min")}
            inputMode="numeric"
            value={filters.priceMin}
            onChange={(e) => onChange({ ...filters, priceMin: e.target.value.replace(/\D/g, "") })}
            className="h-8 text-[12px] px-2"
          />
          <Input
            placeholder={t("filters.max")}
            inputMode="numeric"
            value={filters.priceMax}
            onChange={(e) => onChange({ ...filters, priceMax: e.target.value.replace(/\D/g, "") })}
            className="h-8 text-[12px] px-2"
          />
        </div>
      </Section>

      <Section title={t("filters.condition")}>
        <div className="flex flex-col gap-2.5">
          {CONDITIONS.map((c) => (
            <FacetOption
              key={c.value}
              label={t(c.labelKey)}
              count={facets.condition?.[c.value]}
              checked={filters.condition === c.value}
              onToggle={() => toggle("condition", c.value)}
            />
          ))}
        </div>
      </Section>

      <Section title={t("filters.saleOrRent")}>
        <div className="flex flex-col gap-2.5">
          {LISTING_TYPES.map((lt) => (
            <FacetOption
              key={lt.value}
              label={t(lt.labelKey)}
              count={facets.listing_type?.[lt.value]}
              checked={filters.listingType === lt.value}
              onToggle={() => toggle("listingType", lt.value)}
            />
          ))}
        </div>
      </Section>
    </div>
  );
}
