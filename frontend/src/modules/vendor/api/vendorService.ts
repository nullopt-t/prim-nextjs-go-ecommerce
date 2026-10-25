import { api } from "@/shared/api/client";
import type {
  RequestPayoutPayload,
  Vendor,
  VendorPayout,
  VendorStats,
} from "@/modules/vendor/types";

export type { Vendor, VendorPayout, VendorStats, RequestPayoutPayload };

export const vendorService = {
  getVendors: () => api.get<any>("/api/v1/vendors"),
  getVendorById: (id: string) => api.get<any>(`/api/v1/vendors/${id}`),
  getPayouts: (vendorId?: string) =>
    api.get<any>(
      `/api/v1/vendor/payouts${vendorId ? `?vendorId=${vendorId}` : ""}`
    ),
  requestPayout: (payload: RequestPayoutPayload) =>
    api.post<any, RequestPayoutPayload>("/api/v1/vendor/payouts", payload),
  getVendorStats: (vendorId?: string) =>
    api.get<any>(`/api/v1/vendor/stats${vendorId ? `?vendorId=${vendorId}` : ""}`),
};
