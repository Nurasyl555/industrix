"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Slider } from "@/components/ui/slider";
import { Input } from "@/components/ui/input";
import { type Filters } from "@/types";

const CATEGORIES = [
  { label: "Excavators", count: 124 },
  { label: "Loaders",    count: 86  },
  { label: "Dozers",     count: 42  },
];

const CONDITIONS = ["New", "Used"];

interface FilterSidebarProps {
  filters:  Filters;
  onChange: (f: Filters) => void;
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  const [open, setOpen] = useState(true);
  return (
    <div className="border-b border-gray-100 pb-4 mb-4 last:border-none last:mb-0">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center justify-between w-full text-left bg-transparent border-none cursor-pointer p-0 mb-3"
      >
        <span className="text-[13px] font-bold text-gray-900">{title}</span>
        {open ? (
          <ChevronUp size={15} className="text-gray-400" />
        ) : (
          <ChevronDown size={15} className="text-gray-400" />
        )}
      </button>
      {open && children}
    </div>
  );
}

export function FilterSidebar({ filters, onChange }: FilterSidebarProps) {
  const toggle = (key: "categories" | "conditions", value: string) => {
    const arr = filters[key];
    onChange({
      ...filters,
      [key]: arr.includes(value)
        ? arr.filter((v) => v !== value)
        : [...arr, value],
    });
  };

  const handleReset = () => {
    onChange({
      categories: [],
      priceMin: "",
      priceMax: "",
      conditions: [],
      yearMin: 2000,
      yearMax: 2024,
    });
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <span className="text-[15px] font-bold text-gray-900">Filters</span>
        <button
          onClick={handleReset}
          className="text-[13px] font-semibold text-blue-600 hover:text-blue-700 bg-transparent border-none cursor-pointer p-0"
        >
          Reset
        </button>
      </div>

      {/* Category */}
      <Section title="Category">
        <div className="flex flex-col gap-2.5">
          {CATEGORIES.map(({ label, count }) => (
            <label
              key={label}
              className="flex items-center gap-2.5 cursor-pointer"
            >
              <Checkbox
                checked={filters.categories.includes(label)}
                onCheckedChange={() => toggle("categories", label)}
                className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
              />
              <span className="text-[13px] text-gray-700">
                {label}{" "}
                <span className="text-gray-400">({count})</span>
              </span>
            </label>
          ))}
        </div>
      </Section>

      {/* Price Range */}
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

      {/* Condition */}
      <Section title="Condition">
        <div className="flex flex-col gap-2.5">
          {CONDITIONS.map((c) => (
            <label key={c} className="flex items-center gap-2.5 cursor-pointer">
              <Checkbox
                checked={filters.conditions.includes(c)}
                onCheckedChange={() => toggle("conditions", c)}
                className="data-[state=checked]:bg-blue-600 data-[state=checked]:border-blue-600"
              />
              <span className="text-[13px] text-gray-700">{c}</span>
            </label>
          ))}
        </div>
      </Section>

      {/* Year */}
      <Section title="Year">
        <div className="px-1">
          <Slider
            min={2000}
            max={2024}
            step={1}
            value={[filters.yearMin, filters.yearMax]}
            onValueChange={([min, max]) =>
              onChange({ ...filters, yearMin: min, yearMax: max })
            }
            className="mb-3"
          />
          <div className="flex justify-between text-[12px] text-gray-500">
            <span>{filters.yearMin}</span>
            <span>{filters.yearMax}</span>
          </div>
        </div>
      </Section>

      <Button className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold mt-2 rounded-lg">
        Apply Filters
      </Button>
    </div>
  );
}
