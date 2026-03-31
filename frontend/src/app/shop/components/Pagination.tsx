"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";

interface PaginationProps {
  currentPage:  number;
  totalPages:   number;
  totalResults: number;
  perPage:      number;
  onPageChange: (p: number) => void;
}

const btnBase = `
  w-9 h-9 flex items-center justify-center rounded-lg text-[13px] font-semibold
  border transition-colors cursor-pointer
`;

export function Pagination({
  currentPage,
  totalPages,
  totalResults,
  perPage,
  onPageChange,
}: PaginationProps) {
  const start = (currentPage - 1) * perPage + 1;
  const end   = Math.min(currentPage * perPage, totalResults);

  const pages: (number | "…")[] = [];
  if (totalPages <= 5) {
    for (let i = 1; i <= totalPages; i++) pages.push(i);
  } else {
    pages.push(1, 2, 3, "…", totalPages);
  }

  return (
    <div className="flex items-center justify-between py-6 border-t border-gray-100 mt-2">
      <span className="text-[13px] text-gray-500">
        Showing {start} to {end} of {totalResults} results
      </span>

      <div className="flex items-center gap-1.5">
        {/* Prev */}
        <button
          key="prev"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          className={btnBase + (currentPage === 1
            ? "text-gray-300 border-gray-100 cursor-not-allowed bg-white"
            : "text-gray-700 border-gray-200 bg-white hover:border-gray-400"
          )}
        >
          <ChevronLeft size={15} />
        </button>

        {/* Page numbers */}
        {pages.map((p, i) =>
          p === "…" ? (
            <span
              key={`ellipsis-${i}`}
              className="w-9 text-center text-[13px] text-gray-400"
            >
              …
            </span>
          ) : (
            <button
              key={p}
              onClick={() => onPageChange(p as number)}
              className={btnBase + (p === currentPage
                ? "bg-blue-600 text-white border-blue-600"
                : "text-gray-700 border-gray-200 bg-white hover:border-gray-400"
              )}
            >
              {p}
            </button>
          )
        )}

        {/* Next */}
        <button
          key="next"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          className={btnBase + (currentPage === totalPages
            ? "text-gray-300 border-gray-100 cursor-not-allowed bg-white"
            : "text-gray-700 border-gray-200 bg-white hover:border-gray-400"
          )}
        >
          <ChevronRight size={15} />
        </button>
      </div>
    </div>
  );
}
