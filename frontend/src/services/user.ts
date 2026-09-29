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

  // Addresses — @todo Not yet implemented on the backend
  getAddresses: () => api.get("/api/v1/addresses"),
  createAddress: (payload: Record<string, unknown>) =>
    api.post("/api/v1/addresses", payload),
  updateAddress: (id: string, payload: Record<string, unknown>) =>
    api.patch(`/api/v1/addresses/${id}`, payload),
  deleteAddress: (id: string) => api.delete(`/api/v1/addresses/${id}`),

  // Payment methods — @todo Not yet implemented on the backend
  getPaymentMethods: () => api.get("/api/v1/payment-methods"),
  createPaymentMethod: (payload: Record<string, unknown>) =>
    api.post("/api/v1/payment-methods", payload),
  deletePaymentMethod: (id: string) =>
    api.delete(`/api/v1/payment-methods/${id}`),
  setDefaultPaymentMethod: (id: string) =>
    api.patch(`/api/v1/payment-methods/${id}/default`),
};
