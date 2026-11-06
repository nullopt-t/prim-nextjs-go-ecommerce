/**
 * Vendor domain models matching the backend vendor portal API.
 *
 * Money is handled in minor units, as returned by the API. The UI converts a
 * value to major units only when formatting it for display.
 */

export type VendorStatus = "pending" | "active" | "suspended" | "rejected" | string;
export type PayoutStatus = "requested" | "approved" | "processing" | "paid" | "rejected" | string;

export interface Vendor {
  id: string;
  userId: string;
  storeName: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  commissionRateBps: number;
  status: VendorStatus;
  statusReason?: string;
  reviewedBy?: string;
  reviewedAt?: string;
  bankAccountInfo?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface VendorPayout {
  id: string;
  vendorId: string;
  amount: number;
  currency: string;
  status: PayoutStatus;
  referenceNumber?: string;
  notes?: string;
  snapshotAvailableBalance: number;
  createdAt: string;
  processedAt?: string;
}

export interface VendorDashboard {
  availableBalance: number;
  pendingEarnings: number;
  lifetimeEarnings: number;
  outstandingPayouts: number;
  grossRevenue: number;
  orderCount: number;
  productCount: number;
  publishedProductCount: number;
  draftProductCount: number;
  lowStockVariantCount: number;
}

export interface RegisterVendorPayload {
  storeName: string;
  slug: string;
  description?: string;
  logoObjectId?: string;
  bannerObjectId?: string;
}

export interface SlugAvailability {
  available: boolean;
  slug: string;
  reason?: string;
}

export interface UpdateVendorPayload {
  storeName?: string;
  slug?: string;
  description?: string;
  logoObjectId?: string;
  bannerObjectId?: string;
  bankAccountInfo?: Record<string, unknown>;
}

export interface RequestPayoutPayload {
  amount: number;
}

export type ProductStatus = "draft" | "published" | "archived" | string;
export type ProductType = "simple" | "variable" | string;

export interface VendorProduct {
  id: string;
  slug: string;
  title: string;
  status: ProductStatus;
  productType: ProductType;
  price?: number;
  currency?: string;
  sku?: string;
  stock: number;
  thumbnailUrl?: string;
  createdAt: string;
  updatedAt: string;
}

export interface VendorProductListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  category?: string;
  brand?: string;
  cursor?: string;
}

export interface CreateVendorProductPayload {
  brandId?: string;
  categoryId: string;
  slug: string;
  title: string;
  description?: string;
  highlights?: string[];
  productType?: string;
  variationAxes?: string[];
}

export interface UpdateVendorProductPayload {
  brandId?: string;
  categoryId?: string;
  title?: string;
  description?: string;
  highlights?: string[];
  productType?: string;
  variationAxes?: string[];
}

export interface VendorProductResponse {
  id: string;
  vendorId: string;
  slug: string;
  title: string;
  status: ProductStatus;
  productType: ProductType;
  createdAt: string;
  updatedAt: string;
}

export interface VendorVariant {
  id: string;
  productId: string;
  sku: string;
  title: string;
  price?: number;
  crossedOutPrice?: number;
  currency?: string;
  attributes?: Record<string, unknown>;
  isDefault: boolean;
}

export interface CreateVendorVariantPayload {
  sku?: string;
  title: string;
  price?: number;
  crossedOutPrice?: number;
  currency?: string;
  attributes?: Record<string, unknown>;
  isDefault?: boolean;
  thumbnailUrl?: string;
}

export interface UpdateVendorVariantPayload {
  sku?: string;
  title?: string;
  price?: number;
  crossedOutPrice?: number;
  currency?: string;
  attributes?: Record<string, unknown>;
  isDefault?: boolean;
  thumbnailUrl?: string;
}

export type StockReason = "restock" | "return" | "adjustment";

export interface VendorMediaUpload {
  url: string;
}

export interface VendorPromotion {
  id: string;
  code: string;
  discountType: "percentage" | "fixed_amount" | string;
  discountValue: number;
  currency: string;
  minCartAmount?: number;
  usageLimit?: number;
  perUserLimit?: number;
  startsAt?: string;
  expiresAt?: string;
  status: "active" | "disabled" | "archived" | string;
  vendorId?: string;
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateVendorPromotionPayload {
  code: string;
  discountType: "percentage" | "fixed_amount";
  discountValue: number;
  currency?: string;
  minCartAmount?: number;
  usageLimit?: number;
  perUserLimit?: number;
  startsAt?: string;
  expiresAt?: string;
}
export interface AdjustStockPayload {
  quantity: number;
  reason: StockReason;
}

export interface VendorStock {
  variantId: string;
  onHandQuantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  isInStock: boolean;
}

export type VendorOrderStatus =
  | "pending"
  | "paid"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled"
  | "refunded"
  | string;

export interface VendorOrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;
  lineTotal: number;
  vendorEarning: number;
  commissionRateBps: number;
  commissionAmount: number;
  title?: string;
}

/**
 * Merchant-facing order projection: the vendor's own lines only, with frozen
 * money math. No customer PII is included by design.
 */
export interface VendorOrder {
  id: string;
  status: VendorOrderStatus;
  currency: string;
  items: VendorOrderItem[];
  subtotal: number;
  earnings: number;
  createdAt: string;
  updatedAt: string;
}
