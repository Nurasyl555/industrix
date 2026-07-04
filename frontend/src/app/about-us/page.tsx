import Link from "next/link";
import { Building2, ShieldCheck, Handshake, CalendarDays } from "lucide-react";
import { Button } from "@/components/ui/button";

const FEATURES = [
  { icon: Building2, title: "Verified companies", text: "Sellers register with a 12-digit BIN and are verified by our team before their listings go live." },
  { icon: ShieldCheck, title: "Moderated listings", text: "Every listing passes moderation, so the catalog stays trustworthy and spam-free." },
  { icon: Handshake, title: "Direct deals & chat", text: "Buyers inquire and negotiate with sellers in real time — no middlemen." },
  { icon: CalendarDays, title: "Rental bookings", text: "Rent equipment by the day, week or month with conflict-free calendar booking." },
];

export default function AboutUsPage() {
  return (
    <div className="min-h-screen bg-white">
      {/* Hero */}
      <section className="bg-gray-900 px-6 py-20 text-center text-white">
        <div className="mx-auto max-w-3xl">
          <h1 className="text-4xl font-extrabold sm:text-5xl">
            About Industri<span className="text-amber-500">X</span>
          </h1>
          <p className="mt-4 text-lg leading-8 text-gray-300">
            Industrix is a digital marketplace for buying, selling and renting
            industrial equipment across the CIS region — connecting industrial
            enterprises, equipment suppliers and service companies in one place.
          </p>
        </div>
      </section>

      {/* Mission */}
      <section className="mx-auto max-w-3xl px-6 py-16">
        <h2 className="mb-4 text-2xl font-bold text-gray-900">Our mission</h2>
        <p className="text-base leading-8 text-gray-600">
          Industrial equipment transactions have long been slow, opaque and
          built on personal connections. Industrix brings that market online:
          verified sellers, moderated listings, transparent pricing, and safe
          communication — so buyers can find the right machine and sellers can
          reach thousands of qualified buyers. All infrastructure is
          self-hosted and data resides in-region for compliance.
        </p>
      </section>

      {/* Features */}
      <section className="bg-gray-50 px-6 py-16">
        <div className="mx-auto grid max-w-4xl gap-6 sm:grid-cols-2">
          {FEATURES.map((f) => (
            <div key={f.title} className="rounded-2xl border border-gray-200 bg-white p-6">
              <f.icon size={24} className="mb-3 text-amber-500" />
              <h3 className="mb-1 text-lg font-semibold text-gray-900">{f.title}</h3>
              <p className="text-sm leading-6 text-gray-500">{f.text}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="mx-auto max-w-3xl px-6 py-16 text-center">
        <h2 className="mb-3 text-2xl font-bold text-gray-900">Ready to get started?</h2>
        <p className="mb-6 text-gray-500">Browse the catalog or list your own equipment today.</p>
        <div className="flex justify-center gap-3">
          <Button asChild>
            <Link href="/shop/catalog">Browse catalog</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/shop/sell">Sell equipment</Link>
          </Button>
        </div>
      </section>
    </div>
  );
}
