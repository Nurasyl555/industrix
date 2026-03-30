"use client";

import { Bell, Heart, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";

const NAV_LINKS = ["Home", "About Us", "Products", "Help"];

export function Navbar() {
  return (
    <nav className="sticky top-0 z-50 bg-white border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <span
          className="text-[22px] font-extrabold tracking-tight"
          style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
        >
          Industri<span className="text-amber-500">X</span>
        </span>

        {/* Nav links */}
        <ul className="hidden md:flex items-center gap-8 list-none">
          {NAV_LINKS.map((link) => (
            <li key={link}>
              <a
                href="#"
                className="text-[14px] font-medium text-gray-700 hover:text-amber-500 transition-colors no-underline"
              >
                {link}
              </a>
            </li>
          ))}
        </ul>

        {/* Actions */}
        <div className="flex items-center gap-4">
          {[Bell, Heart, ShoppingCart].map((Icon, i) => (
            <button
              key={i}
              className="text-gray-600 hover:text-amber-500 transition-colors bg-transparent border-none cursor-pointer p-1"
            >
              <Icon size={20} />
            </button>
          ))}
          <Button
            size="sm"
            className="bg-gray-900 hover:bg-gray-700 text-white font-semibold px-5"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            Sign In
          </Button>
        </div>
      </div>
    </nav>
  );
}
