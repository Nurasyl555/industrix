"use client";

import { useI18n } from "@/lib/i18n";
import Link from "next/link";
import { Button } from "@/components/ui/button";

export function CtaBanner() {
  const { t } = useI18n();
  return (
    <div className="max-w-7xl mx-auto px-6 mb-14">
      <div
        className="relative overflow-hidden rounded-2xl px-12 py-10 flex items-center justify-between"
        style={{ background: "linear-gradient(110deg, #F5A623 0%, #f0b03a 100%)" }}
      >
        {/* Decorative circles */}
        <div className="absolute right-24 -top-15 w-60 h-60 rounded-full bg-white/6 pointer-events-none" />
        <div className="absolute -right-5 -bottom-7.5 w-52 h-52 rounded-full bg-white/8 pointer-events-none" />

        <div className="relative z-10 max-w-md">
          <h2
            className="text-[24px] font-extrabold text-white mb-2"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            Sell your equipment faster
          </h2>
          <p className="text-[14px] text-white/85 leading-relaxed mb-6">
            List your machines on the world&apos;s leading industrial marketplace
            and reach thousands of verified buyers daily.
          </p>
          <div className="flex gap-3">
            <Button
              asChild
              className="bg-gray-900 hover:bg-gray-700 text-white font-bold px-6"
              style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
            >
              <Link href="/shop/sell">{t("home.startSelling")}</Link>
            </Button>
            <Button
              asChild
              variant="outline"
              className="border-2 border-white/70 text-white bg-transparent hover:bg-white/15 hover:border-white font-bold px-6"
              style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
            >
              <Link href="/shop/catalog">{t("home.learnMore")}</Link>
            </Button>
          </div>
        </div>

        {/* Gear decoration */}
        <span className="absolute right-12 bottom-0 text-[130px] leading-none opacity-[0.12] pointer-events-none select-none">
          ⚙️
        </span>
      </div>
    </div>
  );
}
