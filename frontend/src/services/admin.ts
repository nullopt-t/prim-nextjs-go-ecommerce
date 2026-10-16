import { api } from "@/api/client";

export interface AdminUser {
  id: string;
  email: string;
  fullName?: string | null;
  role: "customer" | "vendor" | "support" | "admin" | "superadmin";
  isEmailVerified: boolean;
  status: "active" | "inactive" | "suspended" | "deleted";
  suspendedUntil?: string | null;
  lockedUntil?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PaginationMeta {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasPrevious: boolean;
  hasNext: boolean;
}

export interface ListUsersParams {
  page?: number;
  pageSize?: number;
  search?: string;
  role?: string;
  status?: string;
}

export interface ListUsersResponse {
  data: AdminUser[];
  meta: PaginationMeta;
}

export interface UpdateRolePayload {
  role: "customer" | "vendor" | "support" | "admin" | "superadmin";
}

export interface UpdateStatusPayload {
  status: "active" | "inactive" | "suspended" | "deleted";
  suspendedUntil?: string | null;
}

export const adminService = {
  // Users Management
  getUsers: async (params?: ListUsersParams): Promise<ListUsersResponse> => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.search) query.set("search", params.search);
    if (params?.role && params.role !== "all") query.set("role", params.role);
    if (params?.status && params.status !== "all") query.set("status", params.status);

    const qs = query.toString();
    const endpoint = `/api/v1/admin/users${qs ? `?${qs}` : ""}`;
    return api.get<ListUsersResponse>(endpoint);
  },

  getUserById: async (id: string): Promise<{ data: AdminUser }> => {
    return api.get<{ data: AdminUser }>(`/api/v1/admin/users/${id}`);
  },

  updateUserRole: async (id: string, payload: UpdateRolePayload): Promise<{ message: string }> => {
    return api.patch<{ message: string }>(`/api/v1/admin/users/${id}/role`, payload);
  },

  updateUserStatus: async (id: string, payload: UpdateStatusPayload): Promise<{ message: string }> => {
    return api.patch<{ message: string }>(`/api/v1/admin/users/${id}/status`, payload);
  },

  revokeUserSessions: async (id: string): Promise<{ message: string }> => {
    return api.post<{ message: string }>(`/api/v1/admin/users/${id}/revoke-sessions`, {});
  },

  // Orders Management
  getOrders: async (params?: { page?: number; pageSize?: number; customerId?: string; status?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.customerId) query.set("customerId", params.customerId);
    if (params?.status && params.status !== "all") query.set("status", params.status);

    const qs = query.toString();
    const endpoint = `/api/v1/admin/orders${qs ? `?${qs}` : ""}`;
    return api.get(endpoint);
  },

  getOrderById: async (id: string): Promise<any> => {
    return api.get(`/api/v1/admin/orders/${id}`);
  },

  updateOrderStatus: async (id: string, status: string, notes?: string): Promise<{ message: string }> => {
    return api.patch<{ message: string }>(`/api/v1/admin/orders/${id}/status`, { status, notes });
  },

  // Dashboard Overview
  getDashboard: async (): Promise<any> => {
    return api.get(`/api/v1/admin/dashboard`);
  },

  // Promotions
  getPromotions: async (): Promise<any> => {
    return api.get(`/api/v1/admin/promotions`);
  },

  createPromotion: async (payload: any): Promise<any> => {
    return api.post(`/api/v1/admin/promotions`, payload);
  },

  updatePromotionStatus: async (id: string, status: string): Promise<any> => {
    return api.patch(`/api/v1/admin/promotions/${id}/status`, { status });
  },

  deletePromotion: async (id: string): Promise<any> => {
    return api.delete(`/api/v1/admin/promotions/${id}`);
  },

  // Payments
  getPayments: async (params?: { page?: number; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/payments${qs ? `?${qs}` : ""}`);
  },

  refundPayment: async (id: string): Promise<any> => {
    return api.post(`/api/v1/admin/payments/${id}/refund`, {});
  },

  updatePaymentStatus: async (id: string, status: string): Promise<any> => {
    return api.patch(`/api/v1/admin/payments/${id}/status`, { status });
  },

  // Media Library
  getMedia: async (params?: { page?: number; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/media${qs ? `?${qs}` : ""}`);
  },

  uploadMedia: async (file: File): Promise<any> => {
    const formData = new FormData();
    formData.append("file", file);
    return api.upload(`/api/v1/admin/media`, formData);
  },

  // Attributes
  getAttributes: async (): Promise<any> => {
    return api.get(`/api/v1/admin/attributes`);
  },

  createAttribute: async (payload: any): Promise<any> => {
    return api.post(`/api/v1/admin/attributes`, payload);
  },

  deleteAttribute: async (id: string): Promise<any> => {
    return api.delete(`/api/v1/admin/attributes/${id}`);
  },

  // Inventory
  getInventory: async (params?: { page?: number; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.page) query.set("page", String(params.page));
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/inventory${qs ? `?${qs}` : ""}`);
  },

  getInventoryLedgers: async (params?: number | { limit?: number; variantId?: string; productId?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (typeof params === "number") {
      query.set("limit", String(params));
    } else if (params) {
      if (params.limit) query.set("limit", String(params.limit));
      if (params.variantId) query.set("variantId", params.variantId);
      if (params.productId) query.set("productId", params.productId);
    }
    const qs = query.toString();
    return api.get(`/api/v1/admin/inventory/ledgers${qs ? `?${qs}` : ""}`);
  },

  adjustInventory: async (variantId: string, quantity: number, reason: string): Promise<any> => {
    return api.post(`/api/v1/admin/inventory/variants/${variantId}/adjust`, { quantity, reason });
  },

  adjustProductInventory: async (productId: string, quantity: number, reason: string): Promise<any> => {
    return api.post(`/api/v1/admin/inventory/products/${productId}/adjust`, { quantity, reason });
  },

  // Settings
  getSettings: async (): Promise<{ data: Record<string, any> }> => {
    return api.get(`/api/v1/admin/settings`);
  },

  updateSettings: async (items: Record<string, any>): Promise<{ message: string }> => {
    return api.put<{ message: string }>(`/api/v1/admin/settings`, { items });
  },
};
