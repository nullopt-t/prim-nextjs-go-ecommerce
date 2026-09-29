import { api } from "@/api/client";

export interface AddressPayload {
  name: string;
  street: string;
  city: string;
  state?: string;
  zip: string;
  country: string;
  isDefault?: boolean;
}

export interface PaymentMethodPayload {
  cardNumber: string;
  cardHolder: string;
  expiryMonth: string;
  expiryYear: string;
  cvv: string;
  isDefault?: boolean;
}

export const userService = {
  getProfile: () => api.get("/api/v1/auth/me"),
  updateProfile: (payload: any) => api.patch("/api/v1/auth/me", payload),

  getAddresses: () => api.get("/api/v1/user/addresses"),
  createAddress: (payload: AddressPayload) => api.post("/api/v1/user/addresses", payload),
  updateAddress: (id: string, payload: Partial<AddressPayload>) =>
    api.patch(`/api/v1/user/addresses/${id}`, payload),
  deleteAddress: (id: string) => api.delete(`/api/v1/user/addresses/${id}`),

  getPaymentMethods: () => api.get("/api/v1/user/payment-methods"),
  createPaymentMethod: (payload: PaymentMethodPayload) =>
    api.post("/api/v1/user/payment-methods", payload),
  deletePaymentMethod: (id: string) => api.delete(`/api/v1/user/payment-methods/${id}`),
  setDefaultPaymentMethod: (id: string) =>
    api.patch(`/api/v1/user/payment-methods/${id}/default`),

  getReviews: () => api.get("/api/v1/user/reviews"),
  markReviewHelpful: (id: string) => api.post(`/api/v1/user/reviews/${id}/helpful`),
  deleteReview: (id: string) => api.delete(`/api/v1/user/reviews/${id}`),
};
