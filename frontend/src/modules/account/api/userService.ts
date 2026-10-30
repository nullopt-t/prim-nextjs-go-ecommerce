import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type { Address, CreateAddressPayload, UpdateAddressPayload } from "@/modules/account/types";
import type { Session, User } from "@/modules/auth/types";
import type { Order } from "@/modules/orders/types";
import type { Review } from "@/modules/catalog/types";

export const userService = {
  // Profile — uses auth endpoints
  getProfile: () => api.get<ApiResponse<User>>("/api/v1/auth/me"),
  updateProfile: (payload: { fullName?: string; name?: string; phone?: string }) =>
    api.patch<ApiResponse<User>>("/api/v1/user/profile", payload),

  // Orders
  getOrders: () => api.get<ApiResponse<Order[]>>("/api/v1/orders"),
  getOrderById: (id: string) => api.get<ApiResponse<Order>>(`/api/v1/orders/${id}`),
  cancelOrder: (id: string) => api.post<ApiResponse<Order>>(`/api/v1/orders/${id}/cancel`),

  // Reviews
  getMyReviews: () => api.get<ApiResponse<Review[]>>("/api/v1/reviews/me"),
  /** Alias kept for legacy hook compatibility */
  getReviews: () => api.get<ApiResponse<Review[]>>("/api/v1/reviews/me"),
  deleteReview: (id: string) => api.delete<void>(`/api/v1/reviews/${id}`),
  updateReview: (
    id: string,
    payload: { rating?: number; title?: string; body?: string }
  ) => api.patch<ApiResponse<Review>, { rating?: number; title?: string; body?: string }>(`/api/v1/reviews/${id}`, payload),
  markReviewHelpful: (id: string) =>
    api.post<ApiResponse<{ voted: boolean; helpfulCount: number }>>(`/api/v1/reviews/${id}/helpful`),

  // Sessions
  getSessions: () => api.get<ApiResponse<Session[]>>("/api/v1/auth/sessions"),
  deleteSession: (id: string) => api.delete<void>(`/api/v1/auth/sessions/${id}`),
  deleteAllSessions: () => api.delete<void>("/api/v1/auth/sessions"),

  // Addresses
  getAddresses: () => api.get<ApiResponse<Address[]>>("/api/v1/user/addresses"),
  getAddressById: (id: string) => api.get<ApiResponse<Address>>(`/api/v1/user/addresses/${id}`),
  createAddress: (payload: CreateAddressPayload) =>
    api.post<ApiResponse<Address>, CreateAddressPayload>("/api/v1/user/addresses", payload),
  updateAddress: (id: string, payload: UpdateAddressPayload) =>
    api.put<ApiResponse<Address>, UpdateAddressPayload>(`/api/v1/user/addresses/${id}`, payload),
  setDefaultAddress: (id: string) =>
    api.patch<ApiResponse<Address>>(`/api/v1/user/addresses/${id}/default`),
  deleteAddress: (id: string) => api.delete<void>(`/api/v1/user/addresses/${id}`),

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
