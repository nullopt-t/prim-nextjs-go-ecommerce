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
  getAddresses: (_payload?: Record<string, unknown>) =>
    Promise.reject(new Error("Address management is not yet available.")),
  createAddress: (_payload: Record<string, unknown>) =>
    Promise.reject(new Error("Address management is not yet available.")),
  updateAddress: (_id: string, _payload: Record<string, unknown>) =>
    Promise.reject(new Error("Address management is not yet available.")),
  deleteAddress: (_id: string) =>
    Promise.reject(new Error("Address management is not yet available.")),

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
