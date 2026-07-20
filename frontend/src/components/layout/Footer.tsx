"use client";

import Link from "next/link";
import { useI18n, type TranslationKey } from "@/lib/i18n";

// Links carry a key, not text: these arrays are built once at module load, so
// storing translated labels would pin them to whichever locale loaded first.
type FooterLink = { labelKey: TranslationKey; href: string };

const marketplaceLinks: FooterLink[] = [
  { labelKey: "footer.buyEquipment", href: "/shop/catalog" },
  { labelKey: "footer.sellEquipment", href: "/shop/sell" },
  { labelKey: "footer.rentEquipment", href: "/shop/catalog" },
  { labelKey: "footer.auctionCalendar", href: "#" },
];

const companyLinks: FooterLink[] = [
  { labelKey: "footer.aboutUs", href: "/about-us" },
  { labelKey: "footer.careers", href: "#" },
  { labelKey: "footer.partners", href: "#" },
  { labelKey: "footer.news", href: "#" },
];

const supportLinks: FooterLink[] = [
  { labelKey: "footer.helpCenter", href: "#" },
  { labelKey: "footer.safety", href: "#" },
  { labelKey: "footer.contact", href: "#" },
  { labelKey: "footer.terms", href: "#" },
];

const legalLinks: FooterLink[] = [
  { labelKey: "footer.privacyPolicy", href: "#" },
  { labelKey: "footer.cookiePolicy", href: "#" },
  { labelKey: "footer.accessibility", href: "#" },
];

export function Footer() {
  const { t } = useI18n();
  return (
    <footer className="border-t border-[#E5E7EB] bg-[#F9FAFB] text-[#64748B]">
      <div className="mx-auto max-w-7xl px-8 py-10 lg:px-12">
        <div className="grid grid-cols-1 gap-10 md:grid-cols-2 lg:grid-cols-[1.7fr_1fr_1fr_1fr] lg:gap-8">
          <div className="max-w-[320px]">
            <h2 className="text-[20px] font-bold uppercase tracking-tight text-[#0F172A]">
              Industrix
            </h2>

            <p className="mt-3 text-sm leading-7 text-[#94A3B8]">
              Global leader in industrial equipment commerce. Facilitating the
              trade of heavy machinery for construction, mining, and agriculture
              since 2010.
            </p>

            <div className="mt-6 flex items-center gap-3">
              <SocialButton
                href="#"
                label="Instagram"
                icon={
                  <svg
                    viewBox="0 0 24 24"
                    fill="none"
                    className="h-4 w-4"
                    aria-hidden="true"
                  >
                    <rect
                      x="3"
                      y="3"
                      width="18"
                      height="18"
                      rx="5"
                      stroke="currentColor"
                      strokeWidth="2"
                    />
                    <circle
                      cx="12"
                      cy="12"
                      r="4"
                      stroke="currentColor"
                      strokeWidth="2"
                    />
                    <circle cx="17.5" cy="6.5" r="1.2" fill="currentColor" />
                  </svg>
                }
              />

              <SocialButton
                href="#"
                label="Telegram"
                icon={
                  <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="h-4 w-4"
                    aria-hidden="true"
                  >
                    <path d="M21.5 4.5 18.4 19a1.2 1.2 0 0 1-1.8.7l-4.4-3.2-2.2 2.1c-.2.2-.4.3-.7.3l.4-5 9.1-8.2c.4-.4-.1-.6-.6-.3L7 12.6 2.8 11.3c-.9-.3-.9-1.5.1-1.8L20 3.7c.8-.3 1.7.4 1.5 1.2Z" />
                  </svg>
                }
              />

              <SocialButton
                href="#"
                label="Facebook"
                icon={
                  <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="h-4 w-4"
                    aria-hidden="true"
                  >
                    <path d="M13.5 21v-7h2.4l.4-3h-2.8V9.1c0-.9.3-1.6 1.6-1.6h1.4V4.8c-.2 0-1.1-.1-2.1-.1-2.1 0-3.6 1.3-3.6 3.8V11H8.5v3h2.2v7h2.8Z" />
                  </svg>
                }
              />
            </div>
          </div>

          <FooterColumn titleKey="footer.marketplace" links={marketplaceLinks} />
          <FooterColumn titleKey="footer.company" links={companyLinks} />
          <FooterColumn titleKey="footer.support" links={supportLinks} />
        </div>
      </div>

      <div className="border-t border-[#E5E7EB]">
        <div className="mx-auto flex max-w-7xl flex-col gap-3 px-8 py-4 text-sm text-[#94A3B8] sm:flex-row sm:items-center sm:justify-between lg:px-12">
          <p>© 2026 Industrix International. All rights reserved.</p>

          <div className="flex flex-wrap items-center gap-4 sm:gap-6">
            {legalLinks.map((link) => (
              <Link
                key={link.labelKey}
                href={link.href}
                className="transition hover:text-[#0F172A]"
              >
                {t(link.labelKey)}
              </Link>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}

type FooterColumnProps = {
  titleKey: TranslationKey;
  links: FooterLink[];
};

function FooterColumn({ titleKey, links }: FooterColumnProps) {
  const { t } = useI18n();
  return (
    <div>
      <h3 className="text-lg font-bold uppercase tracking-[0.12em] text-[#0F172A]">
        {t(titleKey)}
      </h3>

      <ul className="mt-4 space-y-3">
        {links.map((link) => (
          <li key={link.labelKey}>
            <Link
              href={link.href}
              className="text-[14px] text-[#94A3B8] transition hover:text-[#0F172A]"
            >
              {t(link.labelKey)}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

type SocialButtonProps = {
  href: string;
  label: string;
  icon: React.ReactNode;
};

function SocialButton({ href, label, icon }: SocialButtonProps) {
  return (
    <Link
      href={href}
      aria-label={label}
      className="flex h-8 w-8 items-center justify-center rounded-full bg-[#F1F5F9] text-[#94A3B8] transition hover:bg-[#E2E8F0] hover:text-[#0F172A]"
    >
      {icon}
    </Link>
  );
}