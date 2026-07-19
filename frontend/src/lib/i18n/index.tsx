"use client";

// Lightweight i18n: a React context over the dictionaries, no routing changes
// and no extra dependency. The chosen locale is kept in localStorage so it
// survives reloads, and the <html lang> attribute is updated to match.

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import {
  DEFAULT_LOCALE,
  LOCALES,
  dictionaries,
  type Locale,
  type TranslationKey,
} from "./dictionaries";

const STORAGE_KEY = "industrix.locale";

interface I18nValue {
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (key: TranslationKey) => string;
}

const I18nContext = createContext<I18nValue | null>(null);

function isLocale(value: string | null): value is Locale {
  return value !== null && (LOCALES as readonly string[]).includes(value);
}

export function I18nProvider({ children }: { children: React.ReactNode }) {
  // Always start from the default so the server and the first client render
  // agree; the stored preference is applied in the effect below, after
  // hydration. Reading localStorage during render would cause a mismatch.
  const [locale, setLocaleState] = useState<Locale>(DEFAULT_LOCALE);

  useEffect(() => {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (isLocale(stored)) setLocaleState(stored);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  const setLocale = useCallback((next: Locale) => {
    setLocaleState(next);
    window.localStorage.setItem(STORAGE_KEY, next);
  }, []);

  const t = useCallback(
    (key: TranslationKey) => dictionaries[locale][key] ?? dictionaries[DEFAULT_LOCALE][key] ?? key,
    [locale],
  );

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

/**
 * Returns the translator and current locale.
 *
 * Falls back to the default dictionary when used outside the provider, so a
 * component rendered in isolation shows Russian text instead of crashing.
 */
export function useI18n(): I18nValue {
  const ctx = useContext(I18nContext);
  if (ctx) return ctx;
  return {
    locale: DEFAULT_LOCALE,
    setLocale: () => {},
    t: (key) => dictionaries[DEFAULT_LOCALE][key] ?? key,
  };
}

export { LOCALES, LOCALE_NAMES, DEFAULT_LOCALE } from "./dictionaries";
export type { Locale, TranslationKey } from "./dictionaries";
