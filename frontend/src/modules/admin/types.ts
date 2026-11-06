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

export interface UpdateRolePayload {
  role: "customer" | "vendor" | "support" | "admin" | "superadmin";
}

export interface UpdateStatusPayload {
  status: "active" | "inactive" | "suspended" | "deleted";
  suspendedUntil?: string | null;
}

export type SystemServiceStatus = "healthy" | "degraded" | "unhealthy";

export interface SystemResourceMetric {
  key: string;
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

export interface AuditLog {
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

export type AuditLogEntry = AuditLog;

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
  points: SalesPoint[];
  totals: SalesTotals;
}

export interface InventoryStock {
  variantId: string;
  onHandQuantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  isInStock: boolean;
}

export interface InventoryLedger {
  id: string;
  variantId: string;
  productId?: string;
  productName?: string;
  variantTitle?: string;
  sku?: string;
  quantity: number;
  reason: string;
  createdAt: string;
}

export interface AdminMedia {
  id: string;
  filename?: string;
  objectKey?: string;
  publicUrl: string;
  url?: string;
  bucket?: string;
  size?: number;
  fileSize?: number;
  contentType?: string;
  createdAt?: string;
}

export interface AdminAttribute {
  id: string;
  name: string;
  type: string;
  key?: string;
  label?: string;
  categoryId?: string;
  valueType?: string;
  values?: string[];
  allowedValues?: string[];
  isFilterable?: boolean;
  sortOrder?: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface AdminPayment {
  id: string;
  orderId: string;
  customerEmail?: string;
  amount: number;
  currency: string;
  status: string;
  paymentMethod?: string;
  provider?: string;
  fee?: number | string;
  createdAt: string;
}

export interface DailySales {
  date: string;
  sales: number;
  orders: number;
}

export interface AdminVendor {
  id: string;
  userId: string;
  storeName: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  commissionRateBps: number;
  status: "pending" | "active" | "suspended" | "rejected" | string;
  statusReason?: string;
  reviewedBy?: string;
  reviewedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface AdminPayout {
  id: string;
  vendorId: string;
  amount: number;
  currency: string;
  status: "requested" | "approved" | "processing" | "paid" | "rejected" | string;
  notes?: string;
  snapshotAvailableBalance: number;
  createdAt: string;
  processedAt?: string;
}

export interface AdminDashboardOverview {
  totalRevenue: number;
  totalOrders: number;
  totalCustomers: number;
  totalProducts?: number;
  averageOrderValue?: number;
  pendingOrders?: number;
  processingOrders?: number;
  shippedOrders?: number;
  deliveredOrders?: number;
  canceledOrders?: number;
  recentSales?: DailySales[];
  activeProducts?: number;
  recentOrders?: any[];
  topProducts?: any[];
}
