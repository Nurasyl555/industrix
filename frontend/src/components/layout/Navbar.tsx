"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Bell, CalendarDays, Handshake, Heart, LogOut } from "lucide-react";
import { getCurrentUser, logout, type CurrentUser } from "@/lib/user";
import { unreadCount } from "@/lib/notification";

const NAV_LINKS = [
  { label: "Home", href: "/home" },
  { label: "Products", href: "/shop/catalog" },
  { label: "Sell Equipment", href: "/shop/sell" },
  { label: "My Company", href: "/account/company" },
  { label: "About Us", href: "/about-us" },
];

const ACTION_LINKS = [
  { icon: CalendarDays, href: "/shop/bookings", label: "My Bookings" },
  { icon: Handshake, href: "/shop/deals", label: "My Deals" },
  { icon: Heart, href: "/shop/favorites", label: "Favorites" },
];

export function Navbar() {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [loaded, setLoaded] = useState(false);
  const [unread, setUnread] = useState(0);

  // Re-check on every route change: the navbar lives in the root layout and
  // never remounts on client-side navigation, so a mount-only effect would
  // keep showing "Sign In" after login until a hard refresh.
  useEffect(() => {
    getCurrentUser().then((u) => {
      setUser(u);
      setLoaded(true);
    });
  }, [pathname]);

  // Poll the unread notification count (refreshes on nav + every 20s).
  useEffect(() => {
    let active = true;
    const tick = () => unreadCount().then((n) => active && setUnread(n));
    tick();
    const id = setInterval(tick, 20000);
    return () => { active = false; clearInterval(id); };
  }, [pathname]);

  async function handleLogout() {
    await logout();
    setUser(null);
    router.push("/home");
    router.refresh();
  }

  const displayName = user?.first_name?.trim() || user?.email || "Account";

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
          {NAV_LINKS.map((link) =>
            link.href ? (
              <li key={link.label}>
                <Link
                  href={link.href}
                  className="text-base font-medium text-gray-700 no-underline transition-colors hover:text-amber-500"
                >
                  {link.label}
                </Link>
              </li>
            ) : (
              <li key={link.label}>
                <span
                  title="Coming soon"
                  className="cursor-not-allowed text-base font-medium text-gray-400 select-none"
                >
                  {link.label}
                </span>
              </li>
            )
          )}
          {user?.role === "admin" && (
            <li>
              <Link
                href="/admin"
                className="text-base font-bold text-amber-600 no-underline transition-colors hover:text-amber-500"
              >
                Admin
              </Link>
            </li>
          )}
        </ul>

        {/* Actions */}
        <div className="flex items-center gap-4 justify-self-end">
          {/* Notifications bell with unread badge */}
          <Link
            href="/notifications"
            aria-label="Notifications"
            className="relative cursor-pointer border-none bg-transparent p-1 text-gray-600 transition-colors hover:text-amber-500"
          >
            <Bell size={20} />
            {unread > 0 && (
              <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
                {unread > 9 ? "9+" : unread}
              </span>
            )}
          </Link>

          {ACTION_LINKS.map(({ icon: Icon, href, label }) =>
            href ? (
              <Link
                key={label}
                href={href}
                aria-label={label}
                className="cursor-pointer border-none bg-transparent p-1 text-gray-600 transition-colors hover:text-amber-500"
              >
                <Icon size={20} />
              </Link>
            ) : (
              <span
                key={label}
                title="Coming soon"
                aria-label={`${label} (coming soon)`}
                className="cursor-not-allowed border-none bg-transparent p-1 text-gray-300"
              >
                <Icon size={20} />
              </span>
            )
          )}

          {/* Auth state — only render once we know it, to avoid a flash */}
          {loaded && user ? (
            <div className="flex items-center gap-3">
              <span className="hidden text-sm font-semibold text-gray-800 sm:inline">
                {displayName}
              </span>
              <button
                type="button"
                onClick={handleLogout}
                aria-label="Sign out"
                className="flex items-center gap-1.5 rounded-4xl border border-gray-200 px-4 py-1 text-sm font-semibold text-gray-700 transition-colors hover:bg-gray-100 max-h-8"
              >
                <LogOut size={15} /> Sign Out
              </button>
            </div>
          ) : loaded ? (
            <Link
              href="/auth/login"
              className="rounded-4xl bg-gray-900 px-5 py-1 font-semibold text-white no-underline transition-colors hover:bg-gray-700 max-h-8"
              style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
            >
              Sign In
            </Link>
          ) : (
            // Reserve space while auth state is loading
            <span className="h-8 w-20" />
          )}
        </div>
      </div>
    </nav>
  );
}
