"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { settingsService, type PublicSettings } from "@/services/settings";

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

  useEffect(() => {
    let isMounted = true;
    settingsService
      .getPublicSettings()
      .then((res) => {
        if (isMounted) setSettings(res?.data ?? null);
      })
      .catch(() => {
        // Storefront must keep working if settings cannot be loaded; the UI
        // simply falls back to its built-in defaults.
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, []);

  return (
    <PublicSettingsContext.Provider value={{ settings, loading }}>
      {children}
    </PublicSettingsContext.Provider>
  );
}

export function usePublicSettings() {
  return useContext(PublicSettingsContext);
}
