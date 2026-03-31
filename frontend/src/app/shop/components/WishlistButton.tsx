"use client";

import { Heart } from "lucide-react";
import { Equipment } from "../catalog/types";
import { useFavorites } from "@/store/favorites-store";

interface WishlistButtonProps {
  item:      Equipment;
  size?:     number;
  className?: string;
}

export function WishlistButton({ item, size = 14, className = "" }: WishlistButtonProps) {
  const { toggle, isFav } = useFavorites();
  const fav = isFav(item.id);

  return (
    <button
      onClick={(e) => { e.stopPropagation(); e.preventDefault(); toggle(item); }}
      className={`rounded-full bg-white shadow flex items-center justify-center border-none cursor-pointer transition-colors hover:bg-red-50 ${className}`}
      aria-label={fav ? "Remove from favorites" : "Add to favorites"}
    >
      <Heart
        size={size}
        className={fav ? "fill-red-500 text-red-500" : "text-gray-400"}
      />
    </button>
  );
}
