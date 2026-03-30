"use client";

import { ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

const CATEGORIES = [
  { icon: "🚜", name: "Excavators",   count: "1 340 units" },
  { icon: "⚡", name: "Generators",   count: "806 units"   },
  { icon: "🏗️", name: "Cranes",       count: "260 units"   },
  { icon: "🚛", name: "Tractors",     count: "612 units"   },
  { icon: "🏭", name: "Forklifts",    count: "986 units"   },
  { icon: "🔧", name: "Compressors",  count: "806 units"   },
];

export function CategoryGrid() {
  return (
    <section className="max-w-7xl mx-auto px-6 py-14">
      <div className="flex items-center justify-between mb-7">
        <h2
          className="text-[20px] font-extrabold text-gray-900"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Browse by Category
        </h2>
        <a
          href="#"
          className="text-[14px] font-semibold text-amber-500 hover:underline flex items-center gap-1 no-underline"
        >
          All Categories <ChevronRight size={15} />
        </a>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        {CATEGORIES.map((cat) => (
          <Card
            key={cat.name}
            className="border border-gray-200 hover:border-amber-400 hover:shadow-md hover:-translate-y-1 transition-all duration-200 cursor-pointer"
          >
            <CardContent className="p-5 text-center">
              <span className="text-3xl block mb-3">{cat.icon}</span>
              <p
                className="text-[13px] font-bold text-gray-900"
                style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
              >
                {cat.name}
              </p>
              <p className="text-[11px] text-gray-400 mt-1">{cat.count}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
