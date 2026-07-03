"use client";

import { useState, useCallback, useEffect } from "react";
import { Filters, SortOption, ViewMode } from "@/types";
import { FilterSidebar } from "../components/FilterSidebar";
import { CatalogHeader } from "../components/CatalogHeader";
import { ActiveFilters } from "../components/ActiveFilters";
import { EquipmentGrid } from "../components/EquipmentGrid";
import { Pagination } from "../components/Pagination";
import { listCategories, type Category } from "@/lib/catalog";
import { listActiveListings, type ListingView } from "@/lib/listing";
import { friendlyError } from "@/lib/api";

const DEFAULT_FILTERS: Filters = {
  categoryId: "",
  priceMin: "",
  priceMax: "",
  condition: "",
  listingType: "",
};

const PER_PAGE = 12;

export default function CatalogPage() {
  const [filters, setFilters] = useState<Filters>(DEFAULT_FILTERS);
  const [sort, setSort] = useState<SortOption>("newest");
  const [view, setView] = useState<ViewMode>("grid");
  const [page, setPage] = useState(1);

  const [categories, setCategories] = useState<Category[]>([]);
  const [items, setItems] = useState<ListingView[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");

    listActiveListings({
      category_id: filters.categoryId || undefined,
      condition: filters.condition || undefined,
      listing_type: filters.listingType || undefined,
      price_min: filters.priceMin ? Number(filters.priceMin) : undefined,
      price_max: filters.priceMax ? Number(filters.priceMax) : undefined,
      sort,
      page,
      limit: PER_PAGE,
    })
      .then((res) => {
        if (cancelled) return;
        setItems(res.items ?? []);
        setTotal(res.total);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(friendlyError(err));
        setItems([]);
        setTotal(0);
      })
      .finally(() => !cancelled && setLoading(false));

    return () => { cancelled = true; };
  }, [filters, sort, page]);

  const handleFilterChange = useCallback((next: Filters) => {
    setFilters(next);
    setPage(1);
  }, []);

  const handleFilterRemove = useCallback((key: keyof Filters) => {
    setFilters((prev) => ({ ...prev, [key]: "" }));
    setPage(1);
  }, []);

  const handleClearAll = useCallback(() => {
    setFilters(DEFAULT_FILTERS);
    setPage(1);
  }, []);

  const categoryName = categories.find((c) => c.id === filters.categoryId)?.name ?? "";
  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <div className="flex-1 max-w-7xl mx-auto w-full px-6 py-8 flex gap-8">
        <aside className="w-55 shrink-0">
          <FilterSidebar filters={filters} categories={categories} onChange={handleFilterChange} />
        </aside>

        <main className="flex-1 min-w-0">
          <CatalogHeader
            total={total}
            sort={sort}
            view={view}
            onSortChange={setSort}
            onViewChange={setView}
          />

          <ActiveFilters
            filters={filters}
            categoryName={categoryName}
            onRemove={handleFilterRemove}
            onClearAll={handleClearAll}
          />

          {error && (
            <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {loading ? (
            <div className="py-24 text-center text-gray-400 text-[14px]">Loading…</div>
          ) : (
            <>
              <EquipmentGrid view={view} items={items} />
              <Pagination
                currentPage={page}
                totalPages={totalPages}
                totalResults={total}
                perPage={PER_PAGE}
                onPageChange={setPage}
              />
            </>
          )}
        </main>
      </div>
    </div>
  );
}
