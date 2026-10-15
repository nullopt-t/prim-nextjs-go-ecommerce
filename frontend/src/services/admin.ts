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
};
