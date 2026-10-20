"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { useLocale } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import { settingsService, type PublicSettings } from "@/services/settings";
import { getRememberedLocaleChoice } from "@/lib/localePreference";
import { PUBLIC_SETTINGS_UPDATED_EVENT } from "@/lib/settingsEvents";

interface PublicSettingsContextValue {
  /** The storefront settings, or null while loading / on failure. */
  settings: PublicSettings | null;
  loading: boolean;
}

const PublicSettingsContext = createContext<PublicSettingsContextValue>({
  settings: null,
  loading: true,
});

export function PublicSettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<PublicSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const locale = useLocale();
  const pathname = usePathname();
  const router = useRouter();

  const loadSettings = useCallback(async () => {
    try {
      const res = await settingsService.getPublicSettings();
      setSettings(res?.data ?? null);
    } catch {
      // Storefront must keep working if settings cannot be loaded; the UI
      // simply falls back to its built-in defaults.
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  // Keep storefront data fresh after an admin save and when returning to the tab.
  useEffect(() => {
    const refresh = () => {
      if (document.visibilityState === "visible") loadSettings();
    };
    const onSettingsUpdated = () => loadSettings();

    window.addEventListener("focus", refresh);
    document.addEventListener("visibilitychange", refresh);
    window.addEventListener(PUBLIC_SETTINGS_UPDATED_EVENT, onSettingsUpdated);
    return () => {
      window.removeEventListener("focus", refresh);
      document.removeEventListener("visibilitychange", refresh);
      window.removeEventListener(PUBLIC_SETTINGS_UPDATED_EVENT, onSettingsUpdated);
    };
  }, [loadSettings]);

  // Apply the storefront default language (or the visitor's saved choice) only
  // on unprefixed URLs. Explicit /ar and /en URLs always win.
  useEffect(() => {
    if (loading || !settings) return;
    if (typeof window === "undefined") return;

    const path = window.location.pathname;
    const hasLocalePrefix =
      path === "/ar" || path.startsWith("/ar/") ||
      path === "/en" || path.startsWith("/en/");
    if (hasLocalePrefix) return;

    const target = getRememberedLocaleChoice() ?? settings.defaultLocale;
    if (target !== "ar" && target !== "en") return;
    if (target === locale) return;

    const query = window.location.search;
    router.replace(`${pathname}${query}`, { locale: target });
  }, [loading, settings, locale, pathname, router]);

  return (
    <PublicSettingsContext.Provider value={{ settings, loading }}>
      {children}
    </PublicSettingsContext.Provider>
  );
}

export function usePublicSettings() {
  return useContext(PublicSettingsContext);
}
