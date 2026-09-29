import { CartContextItem, CouponData } from "@/context/CartContext";

export interface CartCalculation {
  subtotal: number;
  discount: number;
  shippingCost: number;
  isFreeShipping: boolean;
  remainingForFreeShipping: number;
  shippingProgress: number;
  total: number;
}

export const SHIPPING_THRESHOLD = 150;
export const STANDARD_SHIPPING_COST = 9.99;

/**
 * Pure calculation function for cart totals, shipping rules, and discounts.
 */
export function calculateCartTotals(
  items: CartContextItem[] = [],
  appliedCoupon: CouponData | null = null,
  shippingThreshold: number = SHIPPING_THRESHOLD,
  standardShippingCost: number = STANDARD_SHIPPING_COST
): CartCalculation {
  const subtotal = items.reduce((acc, item) => {
    const rawPrice = typeof item.productPrice === "number"
      ? item.productPrice
      : parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "")) || 0;
    return acc + rawPrice * (item.quantity || 1);
  }, 0);

  let discount = 0;
  if (appliedCoupon?.discountPercent) {
    discount = (subtotal * appliedCoupon.discountPercent) / 100;
  } else if (appliedCoupon?.fixedDiscount) {
    discount = appliedCoupon.fixedDiscount;
  }

  const isFreeShipping = subtotal >= shippingThreshold || Boolean(appliedCoupon?.freeShipping);
  const shippingCost = isFreeShipping || subtotal === 0 ? 0 : standardShippingCost;
  const remainingForFreeShipping = Math.max(0, shippingThreshold - subtotal);
  const shippingProgress = Math.min(100, (subtotal / shippingThreshold) * 100);

  const total = Math.max(0, subtotal - discount + shippingCost);

  return {
    subtotal,
    discount,
    shippingCost,
    isFreeShipping,
    remainingForFreeShipping,
    shippingProgress,
    total,
  };
}
