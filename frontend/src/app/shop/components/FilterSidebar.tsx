"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { type Filters } from "@/types";
import { type Category } from "@/lib/catalog";

const CONDITIONS: { value: "new" | "used"; label: string }[] = [
  { value: "new", label: "New" },
  { value: "used", label: "Used" },
];

const LISTING_TYPES: { value: "sale" | "rental"; label: string }[] = [
  { value: "sale", label: "For Sale" },
  { value: "rental", label: "For Rent" },
];

interface FilterSidebarProps {
  filters:    Filters;
  categories: Category[];
  onChange:   (f: Filters) => void;
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

export function FilterSidebar({ filters, categories, onChange }: FilterSidebarProps) {
  const handleReset = () => {
    onChange({ categoryId: "", priceMin: "", priceMax: "", condition: "", listingType: "" });
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <span className="text-[15px] font-bold text-gray-900">Filters</span>
        <button
          onClick={handleReset}
          className="text-[13px] font-semibold text-blue-600 hover:text-blue-700 bg-transparent border-none cursor-pointer p-0"
        >
          Reset
        </button>
      </div>

      <Section title="Category">
        <div className="flex flex-col gap-2.5">
          {categories.map((cat) => (
            <label key={cat.id} className="flex items-center gap-2.5 cursor-pointer">
              <Checkbox
                checked={filters.categoryId === cat.id}
                onCheckedChange={() =>
                  onChange({ ...filters, categoryId: filters.categoryId === cat.id ? "" : cat.id })
                }
                className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
              />
              <span className="text-[13px] text-gray-700">{cat.name}</span>
            </label>
          ))}
        </div>
      </Section>

      <Section title="Price Range">
        <div className="flex gap-2">
          <Input
            placeholder="Min"
            value={filters.priceMin}
            onChange={(e) => onChange({ ...filters, priceMin: e.target.value })}
            className="h-8 text-[12px] px-2"
          />
          <Input
            placeholder="Max"
            value={filters.priceMax}
            onChange={(e) => onChange({ ...filters, priceMax: e.target.value })}
            className="h-8 text-[12px] px-2"
          />
        </div>
      </Section>

      <Section title="Condition">
        <div className="flex flex-col gap-2.5">
          {CONDITIONS.map((c) => (
            <label key={c.value} className="flex items-center gap-2.5 cursor-pointer">
              <Checkbox
                checked={filters.condition === c.value}
                onCheckedChange={() =>
                  onChange({ ...filters, condition: filters.condition === c.value ? "" : c.value })
                }
                className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
              />
              <span className="text-[13px] text-gray-700">{c.label}</span>
            </label>
          ))}
        </div>
      </Section>

      <Section title="Sale or Rent">
        <div className="flex flex-col gap-2.5">
          {LISTING_TYPES.map((t) => (
            <label key={t.value} className="flex items-center gap-2.5 cursor-pointer">
              <Checkbox
                checked={filters.listingType === t.value}
                onCheckedChange={() =>
                  onChange({ ...filters, listingType: filters.listingType === t.value ? "" : t.value })
                }
                className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
              />
              <span className="text-[13px] text-gray-700">{t.label}</span>
            </label>
          ))}
        </div>
      </Section>

      <Button className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold mt-2 rounded-lg">
        Apply Filters
      </Button>
    </div>
  );
}
