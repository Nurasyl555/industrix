"use client";

import { useI18n } from "@/lib/i18n";
import { MapPin, Tag } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import Image from "next/image";
import Link from "next/link";
import { type ListingView } from "@/lib/listing";
import { WishlistButton } from "@/app/shop/components/WishlistButton";

function formatPrice(item: ListingView) {
  const price = "$" + item.price.toLocaleString("en-US");
  if (item.listing_type === "rental" && item.price_period) {
    return `${price} / ${item.price_period}`;
  }
  return price;
}

interface ProductCardProps {
  item: ListingView;
}

export function ProductCard({ item }: ProductCardProps) {
  const { t } = useI18n();
  return (
    <Card className="overflow-hidden rounded-3xl border border-gray-200 p-0 hover:-translate-y-1 hover:shadow-xl transition-all duration-200">
      <div className="relative h-56 w-full overflow-hidden rounded-t-3xl bg-sky-100">
        <Image
          src={item.image_url || "/pics/sample.jpg"}
          alt={item.title}
          fill
          className="object-cover object-center"
          sizes="(max-width: 768px) 100vw, 33vw"
          unoptimized
        />
        <WishlistButton item={item} size={15} className="absolute top-3 right-3 h-8 w-8" />
      </div>

      <CardContent className="p-4">
        <p
          className="mb-3 text-[15px] font-semibold leading-tight text-gray-900"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          {item.title}
        </p>

        <div className="mb-1.5 flex items-center gap-1.5 text-[12px] text-gray-500">
          <Tag size={12} className="shrink-0" />
          {item.listing_type === "rental" ? t("listingType.rental") : t("listingType.sale")}
        </div>

        {item.region && (
          <div className="mb-4 flex items-center gap-1.5 text-[12px] text-gray-500">
            <MapPin size={12} className="shrink-0" />
            {item.region}
          </div>
        )}

        <Separator className="mb-3" />

        <div className="flex items-center justify-between">
          <span
            className="text-[18px] font-bold text-gray-900"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            {formatPrice(item)}
          </span>

          <Link
            href={`/shop/details?id=${item.id}`}
            className="flex items-center h-8 border border-gray-200 px-3 text-[12px] hover:border-amber-500 hover:text-amber-600 rounded-md no-underline text-gray-700"
          >
            View Details
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
