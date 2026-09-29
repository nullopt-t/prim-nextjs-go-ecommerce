import { api } from "@/api/client";

export const userService = {
  // Profile — uses auth endpoints
  getProfile: () => api.get("/api/v1/auth/me"),

  // Orders
  getOrders: () => api.get("/api/v1/orders"),
  getOrderById: (id: string) => api.get(`/api/v1/orders/${id}`),

  // Reviews (correct path)
  getMyReviews: () => api.get("/api/v1/reviews/me"),
  deleteReview: (id: string) => api.delete(`/api/v1/reviews/${id}`),
  updateReview: (
    id: string,
    payload: { rating?: number; title?: string; body?: string }
  ) => api.patch(`/api/v1/reviews/${id}`, payload),

  // Sessions
  getSessions: () => api.get("/api/v1/auth/sessions"),
  deleteSession: (id: string) => api.delete(`/api/v1/auth/sessions/${id}`),
};
