"use client";

import { Navbar } from "./components/Navbar";
import { HeroBanner } from "./components/HeroBanner";
import { CategoryGrid } from "./components/CategoryGrid";
import { FeaturedEquipment } from "./components/FeaturedEquipment";
import { CtaBanner } from "./components/CtaBanner";
import { Newsletter } from "./components/Newsletter";

export default function HomePage() {
  return (
    <main className="min-h-screen bg-white">
      <Navbar />
      <HeroBanner />
      <CategoryGrid />
      <FeaturedEquipment />
      <CtaBanner />
      <Newsletter />
    </main>
  );
}
