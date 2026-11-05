"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { Vendor } from "@/modules/vendor/types";

export type { Vendor } from "@/modules/vendor/types";

interface VendorContextType {
  vendor: Vendor | null;
  refreshVendor: () => Promise<void>;
  loading: boolean;
  /** 404 from /api/v1/vendor/profile means there is no merchant profile yet. */
  notRegistered: boolean;
  /** Vendor profile load failed before the server could answer. */
  loadError: Error | null;
}

const VendorContext = createContext<VendorContextType>({
  vendor: null,
  refreshVendor: async () => {},
  loading: true,
  notRegistered: false,
  loadError: null,
});

const PROFILE_ENDPOINT = "/api/v1/vendor/profile";

function isVendorNotFound(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error ?? "");
  return /vendor not found|HTTP Error 404/i.test(message);
}

export function VendorProvider({ children }: { children: React.ReactNode }) {
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [loading, setLoading] = useState(true);
  const [notRegistered, setNotRegistered] = useState(false);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const refreshVendor = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const response = await vendorService.getProfile();
      setVendor(response.data);
      setNotRegistered(false);
    } catch (error) {
      if (isVendorNotFound(error)) {
        setVendor(null);
        setNotRegistered(true);
      } else {
        setLoadError(error instanceof Error ? error : new Error(String(error)));
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshVendor();
  }, [refreshVendor]);

  const value = useMemo(
    () => ({ vendor, refreshVendor, loading, notRegistered, loadError }),
    [vendor, refreshVendor, loading, notRegistered, loadError],
  );

  return <VendorContext.Provider value={value}>{children}</VendorContext.Provider>;
}

export function useVendorContext() {
  return useContext(VendorContext);
}

export { PROFILE_ENDPOINT };
