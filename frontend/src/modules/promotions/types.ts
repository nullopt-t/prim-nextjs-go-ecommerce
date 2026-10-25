/**
 * Promotion & Coupon domain models matching backend catalog/coupon handlers.
 */

export interface Promotion {
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
  status: "active" | "disabled" | "expired" | string;
  usageCount: number;
  createdAt: string;
  updatedAt: string;
}

export type Coupon = Promotion;

export interface CreatePromotionPayload {
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
