"use client";

import { CustomButton } from "@/components/ui/button";
import PaymentMethods from "@/features/cart/components/ui/paymentMethods";
import { useTranslations } from "next-intl";
import { useCart } from "@/hooks/useCart";
import { useRouter } from "@/i18n/navigation";
import { Truck, Sparkles, ShieldCheck } from "lucide-react";

export default function PaymentBox() {
  const t = useTranslations("cart");
  const { cartItems, appliedCoupon } = useCart();
  const router = useRouter();

  if (!cartItems || cartItems.length === 0) {
    return null;
  }

  const subtotal = cartItems.reduce(
    (acc, item) =>
      acc +
      (parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "") || "0") *
        (item.quantity || 1)),
    0
  );

  // Discount calculation
  let discount = 0;
  if (appliedCoupon?.discountPercent) {
    discount = (subtotal * appliedCoupon.discountPercent) / 100;
  } else if (appliedCoupon?.fixedDiscount) {
    discount = appliedCoupon.fixedDiscount;
  }

  const shippingThreshold = 150;
  const isFreeShipping = subtotal >= shippingThreshold || appliedCoupon?.freeShipping;
  const shippingCost = isFreeShipping || subtotal === 0 ? 0 : 9.99;
  const remainingForFreeShipping = Math.max(0, shippingThreshold - subtotal);
  const shippingProgress = Math.min(100, (subtotal / shippingThreshold) * 100);

  const total = Math.max(0, subtotal - discount + shippingCost);

  return (
    <div className="border border-border p-6 rounded-2xl bg-card shadow-xs flex flex-col">
      <h3 className="mb-4 font-semibold text-lg text-foreground">
        {t("orderSummary.title", { defaultMessage: "Order Summary" })}
      </h3>

      {/* Free shipping progress bar */}
      {subtotal > 0 && (
        <div className="mb-6 p-3.5 rounded-xl bg-secondary/40 border border-border/60">
          <div className="flex items-center justify-between text-xs font-medium mb-2">
            <span className="flex items-center gap-1.5 text-foreground">
              <Truck className="size-3.5 text-accent-brand" />
              {isFreeShipping
                ? "You unlocked Free Shipping!"
                : `Add $${remainingForFreeShipping.toFixed(2)} for Free Shipping`}
            </span>
            <span className="text-muted-foreground">{Math.round(shippingProgress)}%</span>
          </div>
          <div className="w-full h-1.5 bg-border rounded-full overflow-hidden">
            <div
              className="h-full bg-accent-brand rounded-full transition-all duration-500 ease-out"
              style={{ width: `${isFreeShipping ? 100 : shippingProgress}%` }}
            />
          </div>
        </div>
      )}

      <div className="flex flex-col gap-3 mb-6">
        <p className="flex justify-between text-sm">
          <span className="text-muted-foreground">
            {t("orderSummary.subtotal", { defaultMessage: "Subtotal" })}
          </span>
          <span className="text-foreground font-medium">${subtotal.toFixed(2)}</span>
        </p>

        {discount > 0 && (
          <p className="flex justify-between text-sm text-emerald-600 dark:text-emerald-400 font-medium">
            <span className="flex items-center gap-1">
              <Sparkles className="size-3.5" />
              Discount ({appliedCoupon?.code})
            </span>
            <span>-${discount.toFixed(2)}</span>
          </p>
        )}

        <p className="flex justify-between text-sm">
          <span className="text-muted-foreground">
            {t("orderSummary.shipping", { defaultMessage: "Shipping" })}
          </span>
          <span className="text-foreground font-medium">
            {isFreeShipping ? (
              <span className="text-emerald-600 dark:text-emerald-400 font-medium">FREE</span>
            ) : (
              `$${shippingCost.toFixed(2)}`
            )}
          </span>
        </p>
        <p className="flex justify-between text-sm">
          <span className="text-muted-foreground">
            {t("orderSummary.tax", { defaultMessage: "Estimated Tax" })}
          </span>
          <span className="text-foreground font-medium">$0.00</span>
        </p>
      </div>

      <hr className="border-border/50" />

      <div className="mt-5 mb-6">
        <p className="flex justify-between items-center text-lg">
          <span className="text-foreground font-semibold">
            {t("orderSummary.total", { defaultMessage: "Total" })}
          </span>
          <span className="text-foreground font-bold text-2xl">${total.toFixed(2)}</span>
        </p>
      </div>

      <div className="w-full">
        <CustomButton
          text={`${t("orderSummary.checkout", { defaultMessage: "Proceed to Checkout" })} (${cartItems?.length || 0})`}
          className="py-3 text-base shadow-sm w-full"
          disabled={!cartItems || cartItems.length === 0}
          onClick={() => router.push("/checkout")}
        />
      </div>

      <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground mt-4">
        <ShieldCheck className="size-3.5 text-emerald-600" />
        <span>Secure 256-Bit SSL Encrypted Checkout</span>
      </div>

      <div className="mt-6 pt-4 border-t border-border/50">
        <PaymentMethods />
      </div>
    </div>
  );
}
