"use client";

import Link from "next/link";
import { Building2, ShieldCheck, Handshake, CalendarDays } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n, type TranslationKey } from "@/lib/i18n";

// Keys rather than text: this array is built at module load, so translated
// strings here would be frozen to whichever locale happened to load first.
const FEATURES: { icon: typeof Building2; titleKey: TranslationKey; textKey: TranslationKey }[] = [
  { icon: Building2, titleKey: "about.verifiedCompanies", textKey: "about.verifiedCompaniesText" },
  { icon: ShieldCheck, titleKey: "about.moderatedListings", textKey: "about.moderatedListingsText" },
  { icon: Handshake, titleKey: "about.directDeals", textKey: "about.directDealsText" },
  { icon: CalendarDays, titleKey: "about.rentalBookings", textKey: "about.rentalBookingsText" },
];

export default function AboutUsPage() {
  const { t } = useI18n();

  return (
    <div className="min-h-screen bg-white">
      {/* Hero */}
      <section className="bg-gray-900 px-6 py-20 text-center text-white">
        <div className="mx-auto max-w-3xl">
          <h1 className="text-4xl font-extrabold sm:text-5xl">
            {t("about.heroTitle")} Industri<span className="text-amber-500">X</span>
          </h1>
          <p className="mt-4 text-lg leading-8 text-gray-300">
            {t("about.heroText")}
          </p>
        </div>
      </section>

      {/* Mission */}
      <section className="mx-auto max-w-3xl px-6 py-16">
        <h2 className="mb-4 text-2xl font-bold text-gray-900">{t("about.mission")}</h2>
        <p className="text-base leading-8 text-gray-600">{t("about.missionText")}</p>
      </section>

      {/* Features */}
      <section className="bg-gray-50 px-6 py-16">
        <div className="mx-auto grid max-w-4xl gap-6 sm:grid-cols-2">
          {FEATURES.map((f) => (
            <div key={f.titleKey} className="rounded-2xl border border-gray-200 bg-white p-6">
              <f.icon size={24} className="mb-3 text-amber-500" />
              <h3 className="mb-1 text-lg font-semibold text-gray-900">{t(f.titleKey)}</h3>
              <p className="text-sm leading-6 text-gray-500">{t(f.textKey)}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="mx-auto max-w-3xl px-6 py-16 text-center">
        <h2 className="mb-3 text-2xl font-bold text-gray-900">{t("about.ctaTitle")}</h2>
        <p className="mb-6 text-gray-500">{t("about.ctaText")}</p>
        <div className="flex justify-center gap-3">
          <Button asChild>
            <Link href="/shop/catalog">{t("about.browseCatalog")}</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/shop/sell">{t("about.sellEquipment")}</Link>
          </Button>
        </div>
      </section>
    </div>
  );
}
