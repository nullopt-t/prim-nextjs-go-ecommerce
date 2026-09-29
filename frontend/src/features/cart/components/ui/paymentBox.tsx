"use client";

import { CustomButton } from "@/components/ui/button";
import PaymentMethods from "@/features/cart/components/ui/paymentMethods";
import { useTranslations } from "next-intl";
import { useCart } from "@/hooks/useCart";
import { useRouter } from "@/i18n/navigation";
import { Truck, Sparkles, ShieldCheck } from "lucide-react";

import { calculateCartTotals } from "@/features/cart/domain/cartCalculations";

export default function PaymentBox({ compact = false }: { compact?: boolean }) {
  const t = useTranslations("cart");
  const { cartItems, appliedCoupon, closeCart } = useCart();
  const router = useRouter();

  if (!cartItems || cartItems.length === 0) {
    return null;
  }

  const {
    subtotal,
    discount,
    shippingCost,
    isFreeShipping,
    remainingForFreeShipping,
    shippingProgress,
    total,
  } = calculateCartTotals(cartItems, appliedCoupon);

  if (compact) {
    return (
      <div className="flex flex-col gap-3">
        {/* Compact Free Shipping Notice */}
        {subtotal > 0 && !isFreeShipping && (
          <div className="flex items-center justify-between text-xs text-muted-foreground bg-secondary/50 px-3 py-1.5 rounded-lg border border-border/40">
            <span className="flex items-center gap-1.5 text-foreground truncate">
              <Truck className="size-3 text-accent-brand shrink-0" />
              <span>Add <strong className="text-accent-brand">${remainingForFreeShipping.toFixed(2)}</strong> for free shipping</span>
            </span>
            <span className="text-[11px] font-semibold">{Math.round(shippingProgress)}%</span>
          </div>
        )}
        {subtotal > 0 && isFreeShipping && (
          <div className="flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-lg border border-emerald-500/20 font-medium">
            <Truck className="size-3 shrink-0" />
            <span>You unlocked Free Shipping!</span>
          </div>
        )}

        {/* Minimal Subtotal & Total Line */}
        <div className="flex items-baseline justify-between pt-1">
          <div className="flex flex-col">
            <span className="text-xs text-muted-foreground">
              {t("orderSummary.total", { defaultMessage: "Total" })}
            </span>
            <span className="text-xl font-bold text-foreground tracking-tight">
              ${total.toFixed(2)}
            </span>
          </div>

          <div className="text-end text-xs text-muted-foreground">
            {discount > 0 ? (
              <span className="text-emerald-600 dark:text-emerald-400 font-medium block">
                -${discount.toFixed(2)} discount
              </span>
            ) : null}
            <span>
              {isFreeShipping ? "Free Shipping" : `+$${shippingCost.toFixed(2)} shipping`}
            </span>
          </div>
        </div>

        {/* Action Button */}
        <CustomButton
          text={`${t("orderSummary.checkout", { defaultMessage: "Proceed to Checkout" })}`}
          className="py-3 text-sm font-semibold w-full shadow-xs rounded-xl"
          disabled={!cartItems || cartItems.length === 0}
          onClick={() => {
            closeCart();
            router.push("/checkout");
          }}
        />

        {/* Micro Security Trust text */}
        <div className="flex items-center justify-center gap-1.5 text-[11px] text-muted-foreground/80">
          <ShieldCheck className="size-3 text-emerald-600" />
          <span>Encrypted checkout</span>
        </div>
      </div>
    );
  }

  return (
    <div className="border border-border rounded-2xl bg-card shadow-xs flex flex-col p-6">
      <h3 className="font-semibold text-foreground text-lg mb-4">
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
          onClick={() => {
            closeCart();
            router.push("/checkout");
          }}
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
