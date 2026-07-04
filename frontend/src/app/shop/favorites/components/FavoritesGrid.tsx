"use client";

import Link from "next/link";
import Image from "next/image";
import { MapPin, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { type ListingView } from "@/lib/listing";
import { WishlistButton } from "../../components/WishlistButton";

function formatPrice(n: number) {
  return "$" + n.toLocaleString("en-US").replace(",", " ");
}

function FavoriteCard({ item }: { item: ListingView }) {
  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden bg-white hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 flex flex-col">
      {/* Image */}
      <div className="relative h-46.25 overflow-hidden bg-sky-50">
        <Image
          src={item.image_url || "/pics/sample.jpg"}
          alt={item.title}
          fill
          className="object-cover"
          unoptimized
        />
        <WishlistButton
          item={item}
          size={15}
          className="absolute top-3 right-3 w-8 h-8"
        />
      </div>

      {/* Body */}
      <div className="p-4 flex flex-col flex-1">
        <h3
          className="text-[14px] font-bold text-gray-900 leading-snug mb-1"
          style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
        >
          {item.title}
        </h3>
        {item.region && (
          <div className="flex items-center gap-1.5 text-[12px] text-gray-500 mb-4">
            <MapPin size={12} className="shrink-0 text-gray-400" />
            {item.region}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between mt-auto pt-3 border-t border-gray-100">
          <span
            className="text-[17px] font-extrabold text-gray-900"
            style={{ fontFamily: "var(--font-gotham,'Outfit',sans-serif)" }}
          >
            {formatPrice(item.price)}
          </span>
          <div className="flex items-center gap-2">
            <Button
              asChild
              variant="outline"
              size="sm"
              className="text-[12px] h-8 px-3 border-gray-200 hover:border-gray-400"
            >
              <Link href={`/shop/details?id=${item.id}`}>View Details</Link>
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 border-gray-200 hover:bg-amber-500 hover:border-amber-500 hover:text-white transition-colors"
            >
              <ShoppingCart size={14} />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

interface FavoritesGridProps {
  items: ListingView[];
}

export function FavoritesGrid({ items }: FavoritesGridProps) {
  return (
    <div className="grid grid-cols-3 gap-5 mt-6">
      {items.map((item) => (
        <FavoriteCard key={item.id} item={item} />
      ))}
    </div>
  );
}
