"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { listCategories, type Category } from "@/lib/catalog";

// Icon per known seed category slug; fallback for anything new.
const ICONS: Record<string, string> = {
  "excavators": "🚜",
  "cranes": "🏗️",
  "generators": "⚡",
  "compressors": "🔧",
  "welding-equipment": "🔥",
  "concrete-mixers": "🏭",
  "loaders": "🚛",
  "trucks-transport": "🚚",
};

export function CategoryGrid() {
  const [categories, setCategories] = useState<Category[]>([]);

  useEffect(() => {
    listCategories()
      .then((cats) => setCategories(cats.slice(0, 6)))
      .catch(() => setCategories([]));
  }, []);

  if (categories.length === 0) return null;

  return (
    <section className="max-w-7xl mx-auto px-6 py-14">
      <div className="flex items-center justify-between mb-7">
        <h2
          className="text-[20px] font-extrabold text-gray-900"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Browse by Category
        </h2>
        <Link
          href="/shop/catalog"
          className="text-[14px] font-semibold text-amber-500 hover:underline flex items-center gap-1 no-underline"
        >
          All Categories <ChevronRight size={15} />
        </Link>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        {categories.map((cat) => (
          <Link key={cat.id} href="/shop/catalog" className="no-underline">
            <Card className="border border-gray-200 hover:border-amber-400 hover:shadow-md hover:-translate-y-1 transition-all duration-200 cursor-pointer h-full">
              <CardContent className="p-5 text-center">
                <span className="text-3xl block mb-3">{ICONS[cat.slug] ?? "⚙️"}</span>
                <p
                  className="text-[13px] font-bold text-gray-900"
                  style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
                >
                  {cat.name}
                </p>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </section>
  );
}
