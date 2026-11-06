import { CartContextItem, CouponData } from "@/modules/cart/model/cartContext";

export interface CartCalculation {
  subtotal: number;
  discount: number;
  shippingCost: number;
  isFreeShipping: boolean;
  remainingForFreeShipping: number;
  shippingProgress: number;
  total: number;
  /** False when the shipping rules are unknown (settings unavailable). */
  shippingKnown: boolean;
}

/**
 * Pure calculation function for cart totals, shipping rules, and discounts.
 *
 * Shipping rules always come from public settings: pass the configured
 * free-shipping threshold (or null when settings are unavailable) and the
 * flat standard rate. When the threshold is unknown the shipping cost is
 * reported as unknown (`shippingKnown = false`) instead of inventing a price.
 */
export function calculateCartTotals(
  items: CartContextItem[] = [],
  appliedCoupon: CouponData | null = null,
  shippingThreshold: number | null,
  standardShippingCost: number,
): CartCalculation {
  const subtotal = items.reduce((acc, item) => {
    const rawPrice = typeof item.productPrice === "number"
      ? item.productPrice
      : parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "")) || 0;
    return acc + rawPrice * (item.quantity || 1);
  }, 0);

  // Calculate discount from coupon
  let discount = 0;
  if (appliedCoupon) {
    if (appliedCoupon.discountAmount && appliedCoupon.discountAmount > 0) {
      discount = appliedCoupon.discountAmount;
    } else if (appliedCoupon.discountPercent) {
      discount = (subtotal * appliedCoupon.discountPercent) / 100;
    } else if (appliedCoupon.fixedDiscount) {
      discount = appliedCoupon.fixedDiscount;
    } else if (appliedCoupon.discountType === "percentage" && appliedCoupon.discountValue) {
      discount = (subtotal * appliedCoupon.discountValue) / 100;
    } else if (appliedCoupon.discountType === "fixed_amount" && appliedCoupon.discountValue) {
      discount = appliedCoupon.discountValue / 100;
    }
  }

  const shippingKnown = shippingThreshold !== null;
  const isFreeShipping = shippingKnown && (subtotal >= (shippingThreshold as number) || Boolean(appliedCoupon?.freeShipping));
  const shippingCost = !shippingKnown || isFreeShipping || subtotal === 0 ? 0 : standardShippingCost;
  const remainingForFreeShipping = shippingKnown ? Math.max(0, (shippingThreshold as number) - subtotal) : 0;
  const shippingProgress = shippingKnown && shippingThreshold ? Math.min(100, (subtotal / shippingThreshold) * 100) : 0;

  const total = Math.max(0, subtotal - discount + shippingCost);

  return {
    subtotal,
    discount,
    shippingCost,
    isFreeShipping,
    remainingForFreeShipping,
    shippingProgress,
    total,
    shippingKnown,
  };
}
