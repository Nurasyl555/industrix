"use client";

import Link from "next/link";
import { Bell, Heart, ShoppingCart } from "lucide-react";

const NAV_LINKS = [
  { label: "Home", href: "/home" },
  { label: "About Us", href: "/about-us" },
  { label: "Products", href: "/products" },
  { label: "Help", href: "/help" },
];

export function Navbar() {
  return (
    <nav className="sticky top-0 z-50 border-b border-gray-200 bg-white">
      <div className="mx-auto grid h-16 max-w-7xl grid-cols-[auto_1fr_auto] items-center px-6">
        {/* Logo */}
        <div className="justify-self-start mr-22">
          <Link
            href="/home"
            className="text-[22px] font-extrabold tracking-tight text-black no-underline"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            Industri<span className="text-amber-500">X</span>
          </Link>
        </div>

        {/* Nav links */}
        <ul className="hidden list-none items-center justify-center gap-8 md:flex">
          {NAV_LINKS.map((link) => (
            <li key={link.label}>
              <Link
                href={link.href}
                className="text-[14px] font-medium text-gray-700 no-underline transition-colors hover:text-amber-500"
              >
                {link.label}
              </Link>
            </li>
          ))}
        </ul>

        {/* Actions */}
        <div className="flex items-center gap-4 justify-self-end">
          {[Bell, Heart, ShoppingCart].map((Icon, i) => (
            <button
              key={i}
              type="button"
              className="cursor-pointer border-none bg-transparent p-1 text-gray-600 transition-colors hover:text-amber-500"
            >
              <Icon size={20} />
            </button>
          ))}

          <Link
            href="/auth/login"
            className="rounded-4xl bg-gray-900 px-5 py-1 font-semibold text-white no-underline transition-colors hover:bg-gray-700 max-h-8"
            style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
          >
            Sign In
          </Link>
        </div>
      </div>
    </nav>
  );
}