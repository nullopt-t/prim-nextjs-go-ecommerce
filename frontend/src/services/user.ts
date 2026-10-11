import { api } from "@/api/client";

export const userService = {
  // Profile — uses auth endpoints
  getProfile: () => api.get("/api/v1/auth/me"),
  /** @todo Not yet implemented on the backend */
  updateProfile: (_payload: Record<string, unknown>) =>
    Promise.reject(new Error("Profile update is not yet available.")),

  // Orders
  getOrders: () => api.get("/api/v1/orders"),
  getOrderById: (id: string) => api.get(`/api/v1/orders/${id}`),

  // Reviews
  getMyReviews: () => api.get("/api/v1/reviews/me"),
  /** Alias kept for legacy hook compatibility */
  getReviews: () => api.get("/api/v1/reviews/me"),
  deleteReview: (id: string) => api.delete(`/api/v1/reviews/${id}`),
  updateReview: (
    id: string,
    payload: { rating?: number; title?: string; body?: string }
  ) => api.patch(`/api/v1/reviews/${id}`, payload),
  /** @todo Not yet implemented on the backend */
  markReviewHelpful: (_id: string) =>
    Promise.reject(new Error("Helpful votes are not yet available.")),

  // Sessions
  getSessions: () => api.get("/api/v1/auth/sessions"),
  deleteSession: (id: string) => api.delete(`/api/v1/auth/sessions/${id}`),
  deleteAllSessions: () => api.delete("/api/v1/auth/sessions"),

  // Addresses
  getAddresses: () => api.get("/api/v1/user/addresses"),
  getAddressById: (id: string) => api.get(`/api/v1/user/addresses/${id}`),
  createAddress: (payload: Record<string, unknown>) =>
    api.post("/api/v1/user/addresses", payload),
  updateAddress: (id: string, payload: Record<string, unknown>) =>
    api.put(`/api/v1/user/addresses/${id}`, payload),
  setDefaultAddress: (id: string) =>
    api.patch(`/api/v1/user/addresses/${id}/default`),
  deleteAddress: (id: string) => api.delete(`/api/v1/user/addresses/${id}`),

  // Payment methods — @todo Not yet implemented on the backend
  getPaymentMethods: () =>
    Promise.reject(new Error("Payment method management is not yet available.")),
  createPaymentMethod: (_payload: Record<string, unknown>) =>
    Promise.reject(new Error("Payment method management is not yet available.")),
  deletePaymentMethod: (_id: string) =>
    Promise.reject(new Error("Payment method management is not yet available.")),
  setDefaultPaymentMethod: (_id: string) =>
    Promise.reject(new Error("Payment method management is not yet available.")),
};
