/**
 * Seller domain models matching the backend seller portal API.
 *
 * Money is handled in minor units, as returned by the API. The UI converts a
 * value to major units only when formatting it for display.
 */

export type SellerStatus = "pending" | "active" | "suspended" | "rejected" | string;
export type PayoutStatus = "requested" | "approved" | "processing" | "paid" | "rejected" | string;

export interface Seller {
  id: string;
  userId: string;
  storeName: string;
  slug: string;
  description?: string;
  logoUrl?: string;
  bannerUrl?: string;
  commissionRateBps: number;
  status: SellerStatus;
  statusReason?: string;
  reviewedBy?: string;
  reviewedAt?: string;
  bankAccountInfo?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface SellerPayout {
  id: string;
  sellerId: string;
  amount: number;
  currency: string;
  status: PayoutStatus;
  referenceNumber?: string;
  notes?: string;
  snapshotAvailableBalance: number;
  createdAt: string;
  processedAt?: string;
}

export interface SellerDashboard {
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

export interface RegisterSellerPayload {
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

export interface UpdateSellerPayload {
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

export interface SellerProduct {
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

export interface SellerProductListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  category?: string;
  brand?: string;
  cursor?: string;
}

export interface CreateSellerProductPayload {
  brandId?: string;
  categoryId: string;
  slug: string;
  title: string;
  description?: string;
  highlights?: string[];
  productType?: string;
  variationAxes?: string[];
}

export interface UpdateSellerProductPayload {
  brandId?: string;
  categoryId?: string;
  title?: string;
  description?: string;
  highlights?: string[];
  productType?: string;
  variationAxes?: string[];
}

export interface SellerProductResponse {
  id: string;
  sellerId: string;
  slug: string;
  title: string;
  status: ProductStatus;
  productType: ProductType;
  createdAt: string;
  updatedAt: string;
}

export interface SellerVariant {
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

export interface CreateSellerVariantPayload {
  sku?: string;
  title: string;
  price?: number;
  crossedOutPrice?: number;
  currency?: string;
  attributes?: Record<string, unknown>;
  isDefault?: boolean;
  thumbnailUrl?: string;
}

export interface UpdateSellerVariantPayload {
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

export interface SellerMediaUpload {
  url: string;
}

export interface SellerPromotion {
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
  sellerId?: string;
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateSellerPromotionPayload {
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

export interface SellerStock {
  variantId: string;
  onHandQuantity: number;
  reservedQuantity: number;
  availableQuantity: number;
  isInStock: boolean;
}

export type SellerOrderStatus =
  | "pending"
  | "paid"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled"
  | "refunded"
  | string;

export interface SellerOrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;
  lineTotal: number;
  sellerEarning: number;
  commissionRateBps: number;
  commissionAmount: number;
  title?: string;
}

/**
 * Merchant-facing order projection: the seller's own lines only, with frozen
 * money math. No customer PII is included by design.
 */
export interface SellerOrder {
  id: string;
  status: SellerOrderStatus;
  currency: string;
  items: SellerOrderItem[];
  couponCode?: string;
  discountAmount?: number;
  subtotal: number;
  earnings: number;
  createdAt: string;
  updatedAt: string;
}
