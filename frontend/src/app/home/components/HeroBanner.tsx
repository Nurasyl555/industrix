"use client";

import { useState } from "react";
import { Search, Shield, Phone } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import Link from "next/link"

export function HeroBanner() {
  const [query, setQuery] = useState("");

  return (
    <section
      className="relative overflow-hidden text-center py-18 px-6 max-h-120"
      style={{
        background:
          "linear-gradient(135deg, #0D1F4E 0%, #142260 60%, #1a2d78 100%)",
      }}
    >
      {/* Radial glow */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background:
            "radial-gradient(ellipse 70% 80% at 50% 0%, rgba(245,166,35,0.09) 0%, transparent 70%)",
        }}
      />

      <div className="relative z-10 max-w-4xl mx-auto">
        <h1
          className="text-[clamp(40px,5vw,56px)] font-extrabold leading-tight tracking-tight text-white"
          style={{ fontFamily: "var(--font-inter, 'Outfit', sans-serif)" }}
        >
          The Trusted Marketplace for
          <span className="block text-amber-400">Heavy Machinery</span>
        </h1>

        <p className="mt-4 text-[14px] text-white/60 leading-relaxed max-w-base mx-auto">
          Buy, sell, or rent industrial equipment from verified sellers across the globe.
          <br />
          Professional tools for professional work.
        </p>

        {/* Search bar */}
        <div className="mt-8 mx-auto max-w-140 flex rounded-xl overflow-hidden shadow-2xl bg-white">
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="What equipment are you looking for..."
            className="flex-1 border-none focus-visible:ring-0 focus-visible:ring-offset-0 text-[14px] placeholder:text-gray-400 rounded-none h-13 px-5"
          />
          <Link
            href="/shop/catalog"
            className="inline-flex items-center justify-center rounded-none px-6 bg-amber-500 hover:bg-amber-600 text-white font-bold gap-2 shrink-0 h-13"
          >
            <Search size={15} />
            Search
          </Link>
        </div>

        {/* Trust badges */}
        <div className="mt-5 flex items-center justify-center gap-8 text-[13px] text-white/60">
          <span className="flex items-center gap-2">
            <Shield size={14} className="opacity-75" />
            Safe Transaction Guarantee
          </span>
          <span className="flex items-center gap-2">
            <Phone size={14} className="opacity-75" />
            24/7 Expert Support
          </span>
        </div>
      </div>
    </section>
  );
}
