"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { ApiError, SELLER_NOT_AUTHORIZED_EVENT } from "@/shared/api/client";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { sellerService } from "@/modules/seller/api/sellerService";
import type { Seller, SellerStatus } from "@/modules/seller/types";

export type { Seller } from "@/modules/seller/types";

export type SellerState =
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

interface SellerContextType {
  seller: Seller | null;
  state: SellerState;
  refreshSeller: (options?: RefreshOptions) => Promise<void>;
  loading: boolean;
  /** 404 from /api/v1/seller/profile means there is no merchant profile yet. */
  notRegistered: boolean;
  /** Seller profile load failed before the server could answer. */
  loadError: Error | null;
}

const SellerContext = createContext<SellerContextType>({
  seller: null,
  state: "loading",
  refreshSeller: async () => {},
  loading: true,
  notRegistered: false,
  loadError: null,
});

const PROFILE_ENDPOINT = "/api/v1/seller/profile";

const KNOWN_STATUSES = new Set<SellerStatus>([
  "pending",
  "active",
  "suspended",
  "rejected",
]);

function isSellerNotFound(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.status === 404 || error.code === "SELLER_NOT_FOUND")
  );
}

function stateForStatus(status: string): SellerState {
  return KNOWN_STATUSES.has(status as SellerStatus)
    ? (status as SellerState)
    : "error";
}

export function SellerProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuthContext();
  const [seller, setSeller] = useState<Seller | null>(null);
  const [state, setState] = useState<SellerState>("loading");
  const [loading, setLoading] = useState(true);
  const [notRegistered, setNotRegistered] = useState(false);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const refreshSeller = useCallback(
    async (options?: RefreshOptions) => {
      if (!user) {
        setSeller(null);
        setNotRegistered(false);
        setLoadError(null);
        setState("unauthenticated");
        setLoading(false);
        return;
      }

      if (!options?.silent) setLoading(true);
      setLoadError(null);
      try {
        const response = await sellerService.getProfile();
        setSeller(response.data);
        setNotRegistered(false);
        setState(stateForStatus(response.data.status));
      } catch (error) {
        if (isSellerNotFound(error)) {
          setSeller(null);
          setNotRegistered(true);
          setState("unregistered");
        } else {
          setSeller(null);
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
    void refreshSeller();
  }, [refreshSeller]);

  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible" && user) {
        void refreshSeller({ silent: true });
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", onVisibilityChange);
  }, [refreshSeller, user]);

  useEffect(() => {
    const onUnauthorized = () => void refreshSeller({ silent: true });
    window.addEventListener(SELLER_NOT_AUTHORIZED_EVENT, onUnauthorized);
    return () =>
      window.removeEventListener(SELLER_NOT_AUTHORIZED_EVENT, onUnauthorized);
  }, [refreshSeller]);

  const value = useMemo(
    () => ({ seller, state, refreshSeller, loading, notRegistered, loadError }),
    [seller, state, refreshSeller, loading, notRegistered, loadError],
  );

  return <SellerContext.Provider value={value}>{children}</SellerContext.Provider>;
}

export function useSellerContext() {
  return useContext(SellerContext);
}

export { PROFILE_ENDPOINT };
