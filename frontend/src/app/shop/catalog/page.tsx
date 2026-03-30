"use client";

import { useState, useCallback } from "react";
import { Filters, SortOption, ViewMode } from "./types";
import { FilterSidebar } from "../components/FilterSidebar";
import { CatalogHeader } from "../components/CatalogHeader";
import { EquipmentGrid } from "../components/EquipmentGrid";
import { Pagination } from "../components/Pagination";

const DEFAULT_FILTERS: Filters = {
  categories: ["Excavators"],
  priceMin: "",
  priceMax: "",
  conditions: ["Used"],
  yearMin: 2000,
  yearMax: 2024,
};

export default function CatalogPage() {
  const [filters, setFilters] = useState<Filters>(DEFAULT_FILTERS);
  const [sort, setSort] = useState<SortOption>("relevance");
  const [view, setView] = useState<ViewMode>("grid");
  const [page, setPage] = useState(1);

  const handleFilterChange = useCallback((next: Filters) => {
    setFilters(next);
    setPage(1);
  }, []);

  const handleFilterRemove = useCallback(
    (key: keyof Filters, value?: string) => {
      setFilters((prev) => {
        const next = { ...prev };
        if (key === "categories" || key === "conditions") {
          next[key] = (prev[key] as string[]).filter((v) => v !== value);
        } else {
          (next as Record<string, unknown>)[key] =
            key === "priceMin" || key === "priceMax" ? "" : prev[key];
        }
        return next;
      });
      setPage(1);
    },
    []
  );

  const handleClearAll = useCallback(() => {
    setFilters({ categories: [], priceMin: "", priceMax: "", conditions: [], yearMin: 2000, yearMax: 2024 });
    setPage(1);
  }, []);

  return (
    <div className="min-h-screen bg-white flex flex-col">

      <div className="flex-1 max-w-7xl mx-auto w-full px-6 py-8 flex gap-8">
        {/* Sidebar */}
        <aside className="w-55 shrink-0">
          <FilterSidebar filters={filters} onChange={handleFilterChange} />
        </aside>

        {/* Main content */}
        <main className="flex-1 min-w-0">
          <CatalogHeader
            total={124}
            query="Used Excavators"
            sort={sort}
            view={view}
            onSortChange={setSort}
            onViewChange={setView}
          />

          {/* <ActiveFilters
            filters={filters}
            onRemove={handleFilterRemove}
            onClearAll={handleClearAll}
          /> */}

          <EquipmentGrid view={view} page={page} sort={sort} filters={filters} />

          <Pagination
            currentPage={page}
            totalPages={11}
            totalResults={124}
            perPage={12}
            onPageChange={setPage}
          />
        </main>
      </div>
    </div>
  );
}
