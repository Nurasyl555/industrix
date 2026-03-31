"use client";

import { X, ArrowDownUp } from "lucide-react";

interface PriceAlertProps {
  onDismiss: () => void;
}

export function PriceAlert({ onDismiss }: PriceAlertProps) {
  return (
    <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-xl px-4 py-3 mb-6">
      <div className="flex items-center gap-2 text-[13px] text-green-800">
        <ArrowDownUp size={15} className="text-green-600 shrink-0" />
        <span>
          <span className="font-bold">Price alert!</span>{" "}
          Price decreased since you saved this item
        </span>
      </div>
      <button
        onClick={onDismiss}
        className="text-green-500 hover:text-green-700 transition-colors bg-transparent border-none cursor-pointer p-0.5"
        aria-label="Dismiss"
      >
        <X size={16} />
      </button>
    </div>
  );
}
