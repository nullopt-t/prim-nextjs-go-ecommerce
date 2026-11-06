"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { ApiError, VENDOR_NOT_AUTHORIZED_EVENT } from "@/shared/api/client";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type { Vendor, VendorStatus } from "@/modules/vendor/types";

export type { Vendor } from "@/modules/vendor/types";

export type VendorState =
  | "loading"
  | "unauthenticated"
  | "unregistered"
  | "pending"
  | "active"
  | "suspended"
  | "rejected"
  | "error";

interface RefreshOptions {
  /** Refresh without flashing the full-page loading state (tab focus). */
  silent?: boolean;
}

interface VendorContextType {
  vendor: Vendor | null;
  state: VendorState;
  refreshVendor: (options?: RefreshOptions) => Promise<void>;
  loading: boolean;
  /** 404 from /api/v1/vendor/profile means there is no merchant profile yet. */
  notRegistered: boolean;
  /** Vendor profile load failed before the server could answer. */
  loadError: Error | null;
}

const VendorContext = createContext<VendorContextType>({
  vendor: null,
  state: "loading",
  refreshVendor: async () => {},
  loading: true,
  notRegistered: false,
  loadError: null,
});

const PROFILE_ENDPOINT = "/api/v1/vendor/profile";

const KNOWN_STATUSES = new Set<VendorStatus>([
  "pending",
  "active",
  "suspended",
  "rejected",
]);

function isVendorNotFound(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.status === 404 || error.code === "VENDOR_NOT_FOUND")
  );
}

function stateForStatus(status: string): VendorState {
  return KNOWN_STATUSES.has(status as VendorStatus)
    ? (status as VendorState)
    : "error";
}

export function VendorProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuthContext();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [state, setState] = useState<VendorState>("loading");
  const [loading, setLoading] = useState(true);
  const [notRegistered, setNotRegistered] = useState(false);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const refreshVendor = useCallback(
    async (options?: RefreshOptions) => {
      if (!user) {
        setVendor(null);
        setNotRegistered(false);
        setLoadError(null);
        setState("unauthenticated");
        setLoading(false);
        return;
      }

      if (!options?.silent) setLoading(true);
      setLoadError(null);
      try {
        const response = await vendorService.getProfile();
        setVendor(response.data);
        setNotRegistered(false);
        setState(stateForStatus(response.data.status));
      } catch (error) {
        if (isVendorNotFound(error)) {
          setVendor(null);
          setNotRegistered(true);
          setState("unregistered");
        } else {
          setVendor(null);
          setNotRegistered(false);
          setLoadError(error instanceof Error ? error : new Error(String(error)));
          setState("error");
        }
      } finally {
        if (!options?.silent) setLoading(false);
      }
    },
    [user],
  );

  useEffect(() => {
    void refreshVendor();
  }, [refreshVendor]);

  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible" && user) {
        void refreshVendor({ silent: true });
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", onVisibilityChange);
  }, [refreshVendor, user]);

  useEffect(() => {
    const onUnauthorized = () => void refreshVendor({ silent: true });
    window.addEventListener(VENDOR_NOT_AUTHORIZED_EVENT, onUnauthorized);
    return () =>
      window.removeEventListener(VENDOR_NOT_AUTHORIZED_EVENT, onUnauthorized);
  }, [refreshVendor]);

  const value = useMemo(
    () => ({ vendor, state, refreshVendor, loading, notRegistered, loadError }),
    [vendor, state, refreshVendor, loading, notRegistered, loadError],
  );

  return <VendorContext.Provider value={value}>{children}</VendorContext.Provider>;
}

export function useVendorContext() {
  return useContext(VendorContext);
}

export { PROFILE_ENDPOINT };
