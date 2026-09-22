import { api } from "@/api/client";

export const vendorService = {
  getVendors: () => api.get("/api/v1/vendors"),
  getVendorById: (id) => api.get(`/api/v1/vendors/${id}`),
  getPayouts: (vendorId) => api.get(`/api/v1/vendor/payouts${vendorId ? `?vendorId=${vendorId}` : ""}`),
  requestPayout: (payload) => api.post("/api/v1/vendor/payouts", payload),
  getVendorStats: (vendorId) => api.get(`/api/v1/vendor/stats${vendorId ? `?vendorId=${vendorId}` : ""}`),
};
