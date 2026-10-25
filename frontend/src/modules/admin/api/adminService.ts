import { api } from "@/shared/api/client";
import type { CursorMeta } from "@/shared/lib/pagination";

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

export interface ListUsersParams {
  cursor?: string;
  pageSize?: number;
  search?: string;
  role?: string;
  status?: string;
}

export interface ListUsersResponse {
  data: AdminUser[];
  meta: CursorMeta;
}

export interface UpdateRolePayload {
  role: "customer" | "vendor" | "support" | "admin" | "superadmin";
}

export interface UpdateStatusPayload {
  status: "active" | "inactive" | "suspended" | "deleted";
  suspendedUntil?: string | null;
}

export type SystemServiceStatus = "healthy" | "degraded" | "unhealthy";

export interface SystemResourceMetric {
  /** Stable identifier the UI maps to a localized label (e.g. "memory"). */
  key: string;
  /** Formatted, already-human-readable reading (e.g. "1.2 MB"). */
  value: string;
}

export interface SystemServiceHealth {
  key: "postgres" | "redis" | "kafka" | "minio" | string;
  status: SystemServiceStatus;
  latencyMs: number;
  resources?: SystemResourceMetric[];
  error?: string;
}

export interface SystemHealthReport {
  status: SystemServiceStatus;
  checkedAt: string;
  services: SystemServiceHealth[];
}

export interface AuditLogEntry {
  id: string;
  actorId: string | null;
  actorRole: string;
  action: string;
  entityType: string;
  entityId: string;
  method: string;
  path: string;
  statusCode: number;
  ipAddress: string;
  userAgent: string;
  before?: Record<string, unknown> | null;
  after?: Record<string, unknown> | null;
  metadata?: Record<string, unknown> | null;
  createdAt: string;
}

export interface StaffSession {
  id: string;
  userId: string;
  email: string;
  fullName?: string | null;
  role: AdminUser["role"];
  userAgent: string;
  ipAddress: string;
  createdAt: string;
  lastActive: string;
  current: boolean;
}

export interface ListAuditParams {
  cursor?: string;
  pageSize?: number;
  action?: string;
  entityType?: string;
  entityId?: string;
  actorId?: string;
  from?: string;
  to?: string;
}

export interface SalesPoint {
  period: string;
  revenue: number;
  orders: number;
}

export interface SalesTotals {
  revenue: number;
  orders: number;
  averageOrderValue: number;
  cancelledOrders: number;
}

export interface SalesReport {
  from: string | null;
  to: string | null;
  groupBy: "day" | "month";
  totals: SalesTotals;
  series: SalesPoint[];
}

export type ReportExportType = "orders" | "payments" | "customers" | "inventory";

export interface ReportQuery {
  dateFrom?: string;
  dateTo?: string;
  status?: string;
}

export const adminService = {
  // Users Management
  getUsers: async (params?: ListUsersParams): Promise<ListUsersResponse> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
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

  // Staff session oversight (superadmin)
  getStaffSessions: async (role?: string): Promise<{ data: StaffSession[] }> => {
    const qs = role ? `?role=${encodeURIComponent(role)}` : "";
    return api.get<{ data: StaffSession[] }>(`/api/v1/admin/sessions${qs}`);
  },

  revokeStaffSession: async (sessionId: string): Promise<{ message: string }> => {
    return api.delete<{ message: string }>(`/api/v1/admin/sessions/${encodeURIComponent(sessionId)}`);
  },

  deleteUser: async (id: string): Promise<{ message: string }> => {
    return api.delete<{ message: string }>(`/api/v1/admin/users/${id}`);
  },

  // Orders Management
  getOrders: async (params?: { cursor?: string; pageSize?: number; customerId?: string; status?: string; dateFrom?: string; dateTo?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.customerId) query.set("customerId", params.customerId);
    if (params?.status && params.status !== "all") query.set("status", params.status);
    if (params?.dateFrom) query.set("dateFrom", params.dateFrom);
    if (params?.dateTo) query.set("dateTo", params.dateTo);

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

  updateOrderTracking: async (id: string, trackingNumber: string): Promise<{ message: string }> => {
    return api.patch<{ message: string }>(`/api/v1/admin/orders/${id}/tracking`, { trackingNumber });
  },

  // Dashboard Overview
  getDashboard: async (): Promise<any> => {
    return api.get(`/api/v1/admin/dashboard`);
  },

  // Promotions
  getPromotions: async (params?: { cursor?: string; pageSize?: number; status?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.status && params.status !== "all") query.set("status", params.status);
    const qs = query.toString();
    return api.get(`/api/v1/admin/promotions${qs ? `?${qs}` : ""}`);
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

  getPromotionRedemptions: async (id: string, params?: { cursor?: string; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/promotions/${id}/redemptions${qs ? `?${qs}` : ""}`);
  },

  // Payments
  getPayments: async (params?: { cursor?: string; pageSize?: number; status?: string; dateFrom?: string; dateTo?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.status && params.status !== "all") query.set("status", params.status);
    if (params?.dateFrom) query.set("dateFrom", params.dateFrom);
    if (params?.dateTo) query.set("dateTo", params.dateTo);
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
  getMedia: async (params?: { cursor?: string; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/media${qs ? `?${qs}` : ""}`);
  },

  uploadMedia: async (file: File): Promise<any> => {
    const formData = new FormData();
    formData.append("file", file);
    return api.upload(`/api/v1/admin/media`, formData);
  },

  deleteMedia: async (id: string): Promise<{ message: string }> => {
    return api.delete<{ message: string }>(`/api/v1/admin/media/${id}`);
  },

  // Attributes
  getAttributes: async (): Promise<any> => {
    return api.get(`/api/v1/admin/attributes`);
  },

  createAttribute: async (payload: any): Promise<any> => {
    return api.post(`/api/v1/admin/attributes`, payload);
  },

  updateAttribute: async (id: string, payload: any): Promise<any> => {
    return api.patch(`/api/v1/admin/attributes/${id}`, payload);
  },

  deleteAttribute: async (id: string): Promise<any> => {
    return api.delete(`/api/v1/admin/attributes/${id}`);
  },

  // Inventory
  getInventory: async (params?: { cursor?: string; pageSize?: number }): Promise<any> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    const qs = query.toString();
    return api.get(`/api/v1/admin/inventory${qs ? `?${qs}` : ""}`);
  },

  getInventoryLedgers: async (params?: number | { cursor?: string; pageSize?: number; variantId?: string; productId?: string }): Promise<any> => {
    const query = new URLSearchParams();
    if (typeof params === "number") {
      query.set("pageSize", String(params));
    } else if (params) {
      if (params.cursor) query.set("cursor", params.cursor);
      if (params.pageSize) query.set("pageSize", String(params.pageSize));
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

  // System
  getSystemHealth: async (): Promise<{ data: SystemHealthReport }> => {
    return api.get(`/api/v1/admin/system/health`);
  },

  // Audit trail (superadmin only)
  getAuditLogs: async (params?: ListAuditParams): Promise<{ data: AuditLogEntry[]; meta: any }> => {
    const query = new URLSearchParams();
    if (params?.cursor) query.set("cursor", params.cursor);
    if (params?.pageSize) query.set("pageSize", String(params.pageSize));
    if (params?.action) query.set("action", params.action);
    if (params?.entityType) query.set("entityType", params.entityType);
    if (params?.entityId) query.set("entityId", params.entityId);
    if (params?.actorId) query.set("actorId", params.actorId);
    if (params?.from) query.set("from", params.from);
    if (params?.to) query.set("to", params.to);
    const qs = query.toString();
    return api.get(`/api/v1/admin/audit${qs ? `?${qs}` : ""}`);
  },

  // Reports & exports
  getSalesReport: async (params?: ReportQuery & { groupBy?: "day" | "month" }): Promise<{ data: SalesReport }> => {
    const query = new URLSearchParams();
    if (params?.dateFrom) query.set("dateFrom", params.dateFrom);
    if (params?.dateTo) query.set("dateTo", params.dateTo);
    if (params?.groupBy) query.set("groupBy", params.groupBy);
    const qs = query.toString();
    return api.get(`/api/v1/admin/reports/sales${qs ? `?${qs}` : ""}`);
  },

  exportReport: async (type: ReportExportType, params?: ReportQuery): Promise<void> => {
    const query = new URLSearchParams();
    query.set("type", type);
    if (params?.dateFrom) query.set("dateFrom", params.dateFrom);
    if (params?.dateTo) query.set("dateTo", params.dateTo);
    if (params?.status) query.set("status", params.status);
    const stamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-");
    await api.download(`/api/v1/admin/reports/export?${query.toString()}`, `${type}-${stamp}.csv`);
  },

  // Settings
  getSettings: async (): Promise<{ data: Record<string, any> }> => {
    return api.get(`/api/v1/admin/settings`);
  },

  updateSettings: async (items: Record<string, any>): Promise<{ message: string }> => {
    return api.put<{ message: string }>(`/api/v1/admin/settings`, { items });
  },
};
