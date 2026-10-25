"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { settingsService, type PublicSettings } from "@/modules/settings/api/settingsService";
import { PUBLIC_SETTINGS_UPDATED_EVENT } from "@/shared/lib/settingsEvents";

interface PublicSettingsContextValue {
  /** The storefront settings, or null while loading / on failure. */
  settings: PublicSettings | null;
  loading: boolean;
}

const PublicSettingsContext = createContext<PublicSettingsContextValue>({
  settings: null,
  loading: true,
});

export function PublicSettingsProvider({
  children,
  initialSettings = null,
}: {
  children: React.ReactNode;
  initialSettings?: PublicSettings | null;
}) {
  const [settings, setSettings] = useState<PublicSettings | null>(initialSettings);
  const [loading, setLoading] = useState(initialSettings === null);

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
    if (initialSettings === null) loadSettings();
  }, [initialSettings, loadSettings]);

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

  return (
    <PublicSettingsContext.Provider value={{ settings, loading }}>
      {children}
    </PublicSettingsContext.Provider>
  );
}

export function usePublicSettings() {
  return useContext(PublicSettingsContext);
}
