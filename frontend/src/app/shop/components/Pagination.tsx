"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";

interface PaginationProps {
  currentPage:  number;
  totalPages:   number;
  totalResults: number;
  perPage:      number;
  onPageChange: (p: number) => void;
}

export function Pagination({
  currentPage,
  totalPages,
  totalResults,
  perPage,
  onPageChange,
}: PaginationProps) {
  const start = (currentPage - 1) * perPage + 1;
  const end   = Math.min(currentPage * perPage, totalResults);

  // Build page buttons: always show 1, 2, 3, …, last
  const pages: (number | "…")[] = [];
  if (totalPages <= 5) {
    for (let i = 1; i <= totalPages; i++) pages.push(i);
  } else {
    pages.push(1, 2, 3, "…", totalPages);
  }

  const btn = (
    label: React.ReactNode,
    onClick: () => void,
    active = false,
    disabled = false
  ) => (
    <button
      key={String(label)}
      onClick={onClick}
      disabled={disabled}
      className={`
        w-9 h-9 flex items-center justify-center rounded-lg text-[13px] font-semibold
        border transition-colors cursor-pointer
        ${active
          ? "bg-blue-600 text-white border-blue-600"
          : disabled
          ? "text-gray-300 border-gray-100 cursor-not-allowed bg-white"
          : "text-gray-700 border-gray-200 bg-white hover:border-gray-400"}
      `}
    >
      {label}
    </button>
  );

  return (
    <div className="flex items-center justify-between py-6 border-t border-gray-100 mt-2">
      <span className="text-[13px] text-gray-500">
        Showing {start} to {end} of {totalResults} results
      </span>

      <div className="flex items-center gap-1.5">
        {btn(
          <ChevronLeft size={15} />,
          () => onPageChange(currentPage - 1),
          false,
          currentPage === 1
        )}

        {pages.map((p, i) =>
          p === "…" ? (
            <span key={`ellipsis-${i}`} className="w-9 text-center text-[13px] text-gray-400">
              …
            </span>
          ) : (
            btn(p, () => onPageChange(p as number), p === currentPage)
          )
        )}

        {btn(
          <ChevronRight size={15} />,
          () => onPageChange(currentPage + 1),
          false,
          currentPage === totalPages
        )}
      </div>
    </div>
  );
}
