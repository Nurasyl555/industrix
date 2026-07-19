"use client";

import { useState, useCallback, useEffect } from "react";
import { Search, X } from "lucide-react";
import { EMPTY_FILTERS, type Filters, type SortOption, type ViewMode } from "@/types";
import { FilterSidebar } from "../components/FilterSidebar";
import { CatalogHeader } from "../components/CatalogHeader";
import { ActiveFilters } from "../components/ActiveFilters";
import { EquipmentGrid } from "../components/EquipmentGrid";
import { Pagination } from "../components/Pagination";
import { listCategories, type Category } from "@/lib/catalog";
import { type ListingView } from "@/lib/listing";
import { search, docToListingView, type SearchResult } from "@/lib/search";
import { useI18n } from "@/lib/i18n";

const PER_PAGE = 12;

/** Debounce so typing doesn't fire a request per keystroke. */
function useDebounced<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(id);
  }, [value, delay]);
  return debounced;
}

export default function CatalogPage() {
  const { t } = useI18n();

  const [query, setQuery] = useState("");
  const debouncedQuery = useDebounced(query);

  const [filters, setFilters] = useState<Filters>(EMPTY_FILTERS);
  const [sort, setSort] = useState<SortOption>("newest");
  const [view, setView] = useState<ViewMode>("grid");
  const [page, setPage] = useState(1);

  const [categories, setCategories] = useState<Category[]>([]);
  const [items, setItems] = useState<ListingView[]>([]);
  const [facets, setFacets] = useState<SearchResult["facets"]>({});
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    listCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  // A new query or filter set restarts paging — otherwise you can land on
  // page 5 of a result set that now has one page.
  useEffect(() => {
    setPage(1);
  }, [debouncedQuery, filters, sort]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");

    search({
      q: debouncedQuery || undefined,
      category_id: filters.categoryId || undefined,
      region: filters.region || undefined,
      condition: filters.condition || undefined,
      listing_type: filters.listingType || undefined,
      min_price: filters.priceMin ? Number(filters.priceMin) : undefined,
      max_price: filters.priceMax ? Number(filters.priceMax) : undefined,
      sort,
      page,
      limit: PER_PAGE,
    })
      .then((res) => {
        if (cancelled) return;
        setItems((res.items ?? []).map(docToListingView));
        setFacets(res.facets ?? {});
        setTotal(res.total);
      })
      .catch(() => {
        if (cancelled) return;
        // The search subsystem is optional infrastructure; say so plainly
        // rather than surfacing a raw transport error.
        setError(t("catalog.searchUnavailable"));
        setItems([]);
        setFacets({});
        setTotal(0);
      })
      .finally(() => !cancelled && setLoading(false));

    return () => {
      cancelled = true;
    };
  }, [debouncedQuery, filters, sort, page, t]);

  const handleFilterRemove = useCallback((key: keyof Filters) => {
    setFilters((prev) => ({ ...prev, [key]: "" }));
  }, []);

  const handleClearAll = useCallback(() => {
    setFilters(EMPTY_FILTERS);
    setQuery("");
  }, []);

  const categoryName = categories.find((c) => c.id === filters.categoryId)?.name ?? "";
  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <div className="flex-1 max-w-7xl mx-auto w-full px-6 py-8">
        {/* Full-text search over title and description */}
        <div className="relative mb-6">
          <Search size={18} className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("catalog.searchPlaceholder")}
            aria-label={t("catalog.search")}
            className="w-full h-12 pl-11 pr-11 rounded-xl border border-gray-200 text-[14px] outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 transition"
          />
          {query && (
            <button
              type="button"
              onClick={() => setQuery("")}
              aria-label={t("catalog.clearSearch")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 bg-transparent border-none cursor-pointer p-1"
            >
              <X size={16} />
            </button>
          )}
        </div>

        <div className="flex gap-8">
          <aside className="w-55 shrink-0">
            <FilterSidebar
              filters={filters}
              categories={categories}
              facets={facets}
              onChange={setFilters}
            />
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
              <div className="py-24 text-center text-gray-400 text-[14px]">{t("catalog.loading")}</div>
            ) : items.length === 0 && !error ? (
              <div className="py-24 text-center">
                <p className="text-[15px] font-semibold text-gray-700">{t("catalog.nothingFound")}</p>
                <p className="mt-1 text-[13px] text-gray-400">{t("catalog.nothingFoundHint")}</p>
              </div>
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
    </div>
  );
}
