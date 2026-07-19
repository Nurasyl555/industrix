"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Bell, CalendarDays, Handshake, Heart, LogOut } from "lucide-react";
import { getCurrentUser, logout, type CurrentUser } from "@/lib/user";
import { unreadCount } from "@/lib/notification";
import { LanguageSwitcher } from "@/components/LanguageSwitcher";
import { useI18n, type TranslationKey } from "@/lib/i18n";

const NAV_LINKS: { labelKey: TranslationKey; href: string }[] = [
  { labelKey: "nav.catalog", href: "/shop/catalog" },
  { labelKey: "nav.sell", href: "/shop/sell" },
  { labelKey: "nav.account", href: "/account/company" },
  { labelKey: "nav.about", href: "/about-us" },
];

const ACTION_LINKS: { icon: typeof CalendarDays; href: string; labelKey: TranslationKey }[] = [
  { icon: CalendarDays, href: "/shop/bookings", labelKey: "nav.bookings" },
  { icon: Handshake, href: "/shop/deals", labelKey: "nav.deals" },
  { icon: Heart, href: "/shop/favorites", labelKey: "nav.favorites" },
];

export function Navbar() {
  const router = useRouter();
  const pathname = usePathname();
  const { t } = useI18n();
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
          {NAV_LINKS.map((link) => (
            <li key={link.labelKey}>
              <Link
                href={link.href}
                className="text-base font-medium text-gray-700 no-underline transition-colors hover:text-amber-500"
              >
                {t(link.labelKey)}
              </Link>
            </li>
          ))}
          {user?.role === "admin" && (
            <li>
              <Link
                href="/admin"
                className="text-base font-bold text-amber-600 no-underline transition-colors hover:text-amber-500"
              >
                {t("nav.admin")}
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

          {ACTION_LINKS.map(({ icon: Icon, href, labelKey }) => (
            <Link
              key={labelKey}
              href={href}
              aria-label={t(labelKey)}
              title={t(labelKey)}
              className="cursor-pointer border-none bg-transparent p-1 text-gray-600 transition-colors hover:text-amber-500"
            >
              <Icon size={20} />
            </Link>
          ))}

          <LanguageSwitcher />

          {/* Auth state — only render once we know it, to avoid a flash */}
          {loaded && user ? (
            <div className="flex items-center gap-3">
              <span className="hidden text-sm font-semibold text-gray-800 sm:inline">
                {displayName}
              </span>
              <button
                type="button"
                onClick={handleLogout}
                aria-label={t("nav.signOut")}
                className="flex items-center gap-1.5 rounded-4xl border border-gray-200 px-4 py-1 text-sm font-semibold text-gray-700 transition-colors hover:bg-gray-100 max-h-8"
              >
                <LogOut size={15} /> {t("nav.signOut")}
              </button>
            </div>
          ) : loaded ? (
            <Link
              href="/auth/login"
              className="rounded-4xl bg-gray-900 px-5 py-1 font-semibold text-white no-underline transition-colors hover:bg-gray-700 max-h-8"
              style={{ fontFamily: "var(--font-gotham, 'Outfit', sans-serif)" }}
            >
              {t("nav.login")}
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
