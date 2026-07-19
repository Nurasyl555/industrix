"use client";

import { LOCALES, LOCALE_NAMES, useI18n } from "@/lib/i18n";

/** Two-way toggle between Russian and Kazakh. */
export function LanguageSwitcher({ className = "" }: { className?: string }) {
  const { locale, setLocale } = useI18n();

  return (
    <div className={`flex items-center rounded-md border border-gray-200 overflow-hidden ${className}`}>
      {LOCALES.map((code) => (
        <button
          key={code}
          type="button"
          onClick={() => setLocale(code)}
          aria-pressed={locale === code}
          className={`px-2 py-1 text-[12px] font-semibold cursor-pointer border-none transition-colors ${
            locale === code
              ? "bg-blue-600 text-white"
              : "bg-transparent text-gray-600 hover:bg-gray-100"
          }`}
        >
          {LOCALE_NAMES[code]}
        </button>
      ))}
    </div>
  );
}
