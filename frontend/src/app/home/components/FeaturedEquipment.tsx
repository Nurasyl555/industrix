"use client";

import { ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ProductCard, type Product } from "./ProductCard";

const PRODUCTS: Product[] = [
  {
    name:   "Caterpillar 320 GC Excavator",
    desc:   "20-ton excavator with hydraulic thumb. Low hours, well maintained.",
    loc:    "Almaty, KZ",
    rating: 4.6,
    price:  "$82 500",
  },
  {
    name:   "Caterpillar 320 GC Excavator",
    desc:   "20-ton excavator with hydraulic thumb. Low hours, well maintained.",
    loc:    "Almaty, KZ",
    rating: 4.6,
    price:  "$82 500",
  },
  {
    name:   "Caterpillar 320 GC Excavator",
    desc:   "20-ton excavator with hydraulic thumb. Low hours, well maintained.",
    loc:    "Almaty, KZ",
    rating: 4.6,
    price:  "$82 500",
  },
];

export function FeaturedEquipment() {
  return (
    <section className="max-w-7xl mx-auto px-6 pb-14">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h2
          className="text-[20px] font-extrabold text-gray-900"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Featured Equipment
        </h2>
        <Button
          className="bg-amber-500 hover:bg-amber-600 text-white font-semibold gap-2 text-[13px] h-9"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Sort by: Newest first <ChevronDown size={14} />
        </Button>
      </div>

      {/* Cards grid */}
      <div className="grid grid-cols-3 gap-6">
        {PRODUCTS.map((product, i) => (
          <ProductCard key={i} product={product} index={i} />
        ))}
      </div>

      {/* View all */}
      <div className="text-center mt-10">
        <Button
          className="bg-amber-500 hover:bg-amber-600 text-white font-bold px-10 py-5 text-[15px] h-auto rounded-xl"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          View All Listings
        </Button>
      </div>
    </section>
  );
}
