"use client";

import { useState } from "react";
import { Heart, MapPin, Star, ShoppingCart } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import Image from "next/image";
import Link from "next/link";

export interface Product {
  name: string;
  desc: string;
  loc: string;
  rating: number;
  price: string;
}

interface ProductCardProps {
  product: Product;
  index: number;
}

export function ProductCard({ product, index }: ProductCardProps) {
  const [wished, setWished] = useState(false);

  return (
    <Card className="overflow-hidden rounded-3xl border border-gray-200 p-0 hover:-translate-y-1 hover:shadow-xl transition-all duration-200 cursor-pointer">
      <div className="relative h-56 w-full overflow-hidden rounded-t-3xl bg-sky-100">
        <Image
          src="/pics/sample.jpg"
          alt="Equipment image"
          fill
          className="object-cover object-center"
          sizes="(max-width: 768px) 100vw, 33vw"
        />

        <button
          onClick={(e) => {
            e.stopPropagation();
            setWished(!wished);
          }}
          className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-full bg-white shadow-md transition-colors hover:bg-red-50 border-none cursor-pointer"
          aria-label="Add to wishlist"
        >
          <Heart
            size={15}
            className={wished ? "fill-red-500 text-red-500" : "text-gray-400"}
          />
        </button>
      </div>

      <CardContent className="p-4">
        <p
          className="mb-1 text-[15px] font-semibold leading-tight text-gray-900"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          {product.name}
        </p>

        <p className="mb-3 text-[12px] leading-relaxed text-gray-500">
          {product.desc}
        </p>

        <div className="mb-1.5 flex items-center gap-1.5 text-[12px] text-gray-500">
          <MapPin size={12} className="shrink-0" />
          {product.loc}
        </div>

        <div className="mb-4 flex items-center gap-1 text-[12px] font-semibold text-amber-500">
          <Star size={12} className="fill-amber-500" />
          {product.rating}
        </div>

        <Separator className="mb-3" />

        <div className="flex items-center justify-between">
          <span
            className="text-[18px] font-bold text-gray-900"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            {product.price}
          </span>

          <div className="flex items-center gap-2">
            <Link
              href="/shop/details"
              className="flex items-center h-8 border border-gray-200 px-3 text-[12px] hover:border-amber-500 hover:text-amber-600 rounded-md"
            >
              View Details
            </Link>

            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8 border-gray-200 transition-colors hover:border-amber-500 hover:bg-amber-500 hover:text-white"
            >
              <ShoppingCart size={14} />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}