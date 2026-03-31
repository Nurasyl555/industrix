"use client";

import { useState } from "react";
import { Heart, MapPin, Star, ShoppingCart } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { ExcavatorIllustration } from "./ExcavatorIllustration";

export interface Product {
  name:   string;
  desc:   string;
  loc:    string;
  rating: number;
  price:  string;
}

interface ProductCardProps {
  product: Product;
  index:   number;
}

export function ProductCard({ product, index }: ProductCardProps) {
  const [wished, setWished] = useState(false);

  return (
    <Card className="overflow-hidden border border-gray-200 hover:shadow-xl hover:-translate-y-1 transition-all duration-200 cursor-pointer">
      {/* Image */}
      <div className="relative h-[200px] overflow-hidden bg-sky-100">
        <ExcavatorIllustration seed={index} />
        <button
          onClick={(e) => {
            e.stopPropagation();
            setWished(!wished);
          }}
          className="absolute top-3 right-3 w-8 h-8 rounded-full bg-white shadow-md flex items-center justify-center hover:bg-red-50 transition-colors border-none cursor-pointer"
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
          className="font-semibold text-[15px] text-gray-900 mb-1 leading-tight"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          {product.name}
        </p>
        <p className="text-[12px] text-gray-500 leading-relaxed mb-3">
          {product.desc}
        </p>

        <div className="flex items-center gap-1.5 text-[12px] text-gray-500 mb-1.5">
          <MapPin size={12} className="shrink-0" />
          {product.loc}
        </div>

        <div className="flex items-center gap-1 text-[12px] font-semibold text-amber-500 mb-4">
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
            <Button
              variant="outline"
              size="sm"
              className="text-[12px] h-8 px-3 border-gray-200 hover:border-amber-500 hover:text-amber-600"
            >
              View Details
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
      </CardContent>
    </Card>
  );
}
