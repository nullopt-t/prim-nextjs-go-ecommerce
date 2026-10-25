import { api } from "@/shared/api/client";

export const vendorService = {
	getVendors: () => api.get("/api/v1/vendors"),
	getVendorById: (id: string) => api.get(`/api/v1/vendors/${id}`),
	getPayouts: (vendorId?: string) =>
		api.get(`/api/v1/vendor/payouts${vendorId ? `?vendorId=${vendorId}` : ""}`),
	requestPayout: (payload: Record<string, unknown>) =>
		api.post("/api/v1/vendor/payouts", payload),
	getVendorStats: (vendorId?: string) =>
		api.get(`/api/v1/vendor/stats${vendorId ? `?vendorId=${vendorId}` : ""}`),
};
