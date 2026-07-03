"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ProductCard } from "./ProductCard";
import { listActiveListings, type ListingView } from "@/lib/listing";

export function FeaturedEquipment() {
  const [items, setItems] = useState<ListingView[]>([]);

  useEffect(() => {
    listActiveListings({ sort: "newest", limit: 3 })
      .then((res) => setItems(res.items ?? []))
      .catch(() => setItems([]));
  }, []);

  if (items.length === 0) return null;

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
      </div>

      {/* Cards grid */}
      <div className="grid grid-cols-3 gap-6">
        {items.map((item) => (
          <ProductCard key={item.id} item={item} />
        ))}
      </div>

      {/* View all */}
      <div className="text-center mt-10">
        <Button
          asChild
          className="bg-amber-500 hover:bg-amber-600 text-white font-bold px-10 py-5 text-[15px] h-auto rounded-xl"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          <Link href="/shop/catalog">View All Listings</Link>
        </Button>
      </div>
    </section>
  );
}
