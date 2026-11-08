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
  const round2 = (n: number) => Math.round(n * 100) / 100;
  const subtotal = round2(items.reduce((acc, item) => {
    const rawPrice = Number.isFinite(item.unitPrice)
      ? (item.unitPrice as number)
      : typeof item.productPrice === "number" && Number.isFinite(item.productPrice)
        ? item.productPrice
        : parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "")) || 0;
    const qty = Number.isFinite(item.quantity) ? item.quantity : 1;
    return acc + rawPrice * qty;
  }, 0));

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

  // Backend caps the discount at the gross total; mirror that here so a fixed
  // coupon applied to a since-shrunk cart cannot display a discount > subtotal.
  discount = Math.min(round2(discount), subtotal);

  const shippingKnown = shippingThreshold !== null;
  // Free shipping is threshold-only: the backend waives shipping solely on the
  // gross-total threshold and has no free-shipping coupon benefit, so a
  // coupon flag must never waive it here (display would drift from charge).
  const isFreeShipping = shippingKnown && subtotal >= (shippingThreshold as number);
  // Settings are major-unit floats; the backend rounds them to minor units
  // (math.Round(v*100)), so round here too or 9.999 displays/charges as 9.99.
  const shippingCost = !shippingKnown || isFreeShipping || subtotal === 0 ? 0 : round2(standardShippingCost);
  const remainingForFreeShipping = shippingKnown ? Math.max(0, (shippingThreshold as number) - subtotal) : 0;
  const shippingProgress = shippingKnown && shippingThreshold !== null && shippingThreshold > 0
    ? Math.min(100, (subtotal / shippingThreshold) * 100)
    : shippingKnown && shippingThreshold === 0
      ? 100
      : 0;

  const total = round2(Math.max(0, subtotal - discount + shippingCost));

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
