"use client";

import { useState } from "react";
import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";
import { useCart } from "@/hooks/useCart";
import { Tag, Check, X } from "lucide-react";
import { toast } from "sonner";

export default function Coupon() {
  const t = useTranslations("cart");
  const { appliedCoupon, applyCoupon, removeCoupon } = useCart();
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);

  const handleApply = async (couponCodeToApply?: string) => {
    const targetCode = (couponCodeToApply || code).trim();
    if (!targetCode) {
      toast.error(t("coupon.errorEmpty", { defaultMessage: "Please enter a coupon code" }));
      return;
    }

    setLoading(true);
    const res = await applyCoupon(targetCode);
    setLoading(false);

    if (res.success) {
      toast.success(
        res.coupon?.message || `Coupon ${targetCode.toUpperCase()} applied!`
      );
      setCode("");
    } else {
      toast.error(res.error || t("coupon.invalid", { defaultMessage: "Failed to apply coupon" }));
    }
  };

  return (
    <div className="flex flex-col gap-3">
      {appliedCoupon ? (
        <div className="flex items-center justify-between p-3.5 bg-emerald-500/10 border border-emerald-500/30 rounded-xl text-sm">
          <div className="flex items-center gap-2.5 text-emerald-700 dark:text-emerald-400">
            <div className="p-1 rounded-full bg-emerald-500/20">
              <Check className="size-3.5" />
            </div>
            <span className="font-semibold tracking-wide">{appliedCoupon.code}</span>
            <span className="text-xs opacity-90">
              ({appliedCoupon.discountPercent
                ? `${appliedCoupon.discountPercent}% OFF`
                : appliedCoupon.freeShipping
                ? "Free Shipping"
                : "Applied"})
            </span>
          </div>
          <button
            type="button"
            onClick={() => {
              removeCoupon();
              toast.info(t("coupon.removed", { defaultMessage: "Coupon removed" }));
            }}
            className="flex items-center gap-1 text-xs text-muted-foreground hover:text-destructive transition-colors px-2 py-1 rounded-md hover:bg-background cursor-pointer"
          >
            <X className="size-3.5" />
            <span>{t("coupon.remove", { defaultMessage: "Remove" })}</span>
          </button>
        </div>
      ) : (
        <>
          <div className="flex gap-4">
            <div className="flex-1 bg-card border border-border focus-within:border-accent-brand rounded-xl transition-colors overflow-hidden flex items-center px-4 h-12 shadow-xs">
              <Tag className="size-4 text-muted-foreground ltr:mr-2.5 rtl:ml-2.5 shrink-0" />
              <input
                type="text"
                placeholder={t("coupon.placeholder", { defaultMessage: "Enter promo code" })}
                value={code}
                onChange={(e) => setCode(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleApply();
                  }
                }}
                className="w-full h-full bg-transparent outline-none text-sm text-foreground placeholder:text-muted-foreground uppercase"
              />
            </div>
            <div className="w-28 md:w-36 h-12 shadow-xs">
              <CustomButton
                text={loading ? "Applying..." : t("coupon.apply", { defaultMessage: "Apply" })}
                onClick={() => handleApply()}
                disabled={loading || !code.trim()}
              />
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground flex-wrap">
            <span>{t("coupon.available", { defaultMessage: "Available codes:" })}</span>
            <button
              onClick={() => handleApply("SAVE10")}
              type="button"
              className="px-2 py-0.5 rounded-md bg-secondary hover:bg-accent-brand/10 hover:text-accent-brand font-mono font-medium transition-colors cursor-pointer border border-border/50"
            >
              SAVE10 (10% OFF)
            </button>
            <button
              onClick={() => handleApply("WELCOME15")}
              type="button"
              className="px-2 py-0.5 rounded-md bg-secondary hover:bg-accent-brand/10 hover:text-accent-brand font-mono font-medium transition-colors cursor-pointer border border-border/50"
            >
              WELCOME15 (15% OFF)
            </button>
          </div>
        </>
      )}
    </div>
  );
}
