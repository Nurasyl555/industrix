"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export function Newsletter() {
  const [email, setEmail] = useState("");

  const handleSubscribe = () => {
    if (!email) return;
    // TODO: wire up to your API
    console.log("Subscribed:", email);
    setEmail("");
  };

  return (
    <section className="bg-gray-100 py-20 px-6 text-center">
      <h2
        className="text-[clamp(22px,3vw,36px)] font-extrabold text-gray-900 mb-3"
        style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
      >
        Never Miss a Heavy Machinery Deal
      </h2>
      <p className="text-lg font-weight-600 text-gray-500 max-w-md mx-auto leading-relaxed mb-8">
        Subscribe to get monthly reports on equipment pricing trends and new
        arrivals in your category.
      </p>

      <div className="flex items-stretch max-w-108 h-12 mx-auto border border-gray-200 rounded-xl overflow-hidden bg-white shadow-sm">
        <Input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSubscribe()}
          placeholder="Your email"
          className="flex-1 h-full border-none focus-visible:ring-0 focus-visible:ring-offset-0 text-[14px] rounded-none px-5"
        />
        <Button
          onClick={handleSubscribe}
          className="rounded-none rounded-r-xl px-7 bg-amber-500 hover:bg-amber-600 text-white font-bold shrink-0 h-full"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Subscribe
        </Button>
      </div>
    </section>
  );
}
