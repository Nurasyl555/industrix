"use client";

import { useState } from "react";
import { Breadcrumb } from "./components/Breadcrumb";
import { PriceAlert } from "./components/PriceAlert";
import { CategoryTabs } from "./components/CategoryTabs";
import { FavoritesGrid } from "./components/FavoritesGrid";
import { useFavorites } from "@/store/favorites-store";

export default function FavoritesPage() {
  const { items } = useFavorites();
  const [activeTab, setActiveTab] = useState("All");
  const [dismissed, setDismissed] = useState(false);
  const [visibleCount, setVisibleCount] = useState(9);

  // Derive unique categories from favorited items
  // In a real app, items would have a `category` field.
  // For now we use a fixed set matching the Figma.
  const tabs = ["All", "Excavators", "Generators"];

  const filtered = activeTab === "All" ? items : items; // extend when items have category field
  const visible  = filtered.slice(0, visibleCount);
  const hasMore  = visibleCount < filtered.length;

  return (
    <div className="min-h-screen bg-white flex flex-col">

      <div className="flex-1 max-w-7xl mx-auto w-full px-6 py-8">
        {/* Top row: breadcrumb + tabs */}
        <div className="flex items-center justify-between mb-5">
          <Breadcrumb />
          <CategoryTabs
            tabs={tabs}
            active={activeTab}
            total={items.length}
            onSelect={setActiveTab}
          />
        </div>

        {/* Price alert banner */}
        {!dismissed && items.length > 0 && (
          <PriceAlert onDismiss={() => setDismissed(true)} />
        )}

        {/* Empty state */}
        {items.length === 0 && (
          <div className="py-32 text-center">
            <p className="text-[16px] font-semibold text-gray-700 mb-1">
              No favorites yet
            </p>
            <p className="text-[13px] text-gray-400">
              Press the heart icon on any listing to save it here.
            </p>
          </div>
        )}

        {/* Grid */}
        {items.length > 0 && (
          <>
            <FavoritesGrid items={visible} />

            {hasMore && (
              <div className="text-center mt-8">
                <button
                  onClick={() => setVisibleCount((c) => c + 9)}
                  className="px-10 py-3 rounded-full bg-amber-500 hover:bg-amber-600 text-white text-[14px] font-bold transition-colors border-none cursor-pointer"
                  style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
                >
                  Load more Items
                </button>
              </div>
            )}
          </>
        )}
      </div>

    </div>
  );
}
