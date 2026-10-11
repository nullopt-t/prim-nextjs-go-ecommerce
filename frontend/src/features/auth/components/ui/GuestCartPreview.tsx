"use client";

import { useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { useCart } from "@/hooks/useCart";
import { LazyImage } from "@/components/ui/lazyImage";
import { ShoppingBag, Sparkles, Trash2, Plus, Minus, Loader2 } from "lucide-react";
import { Toggle } from "@/components/ui/Toggle";
import { toast } from "sonner";

interface GuestCartPreviewProps {
  mergeCart: boolean;
  onToggleMerge: (enabled: boolean) => void;
}

export function GuestCartPreview({ mergeCart, onToggleMerge }: GuestCartPreviewProps) {
  const t = useTranslations("auth.guestCart");
  const locale = useLocale();
  const { cartItems, cartSummary, updateCartItem, removeFromCart } = useCart();
  const [busyItemId, setBusyItemId] = useState<string | null>(null);

  if (!cartItems || cartItems.length === 0) {
    return null;
  }

  const handleUpdateQty = async (itemId: string, currentQty: number, delta: number, availableStock?: number) => {
    const newQty = currentQty + delta;
    if (newQty <= 0) {
      handleRemove(itemId);
      return;
    }

    if (availableStock !== undefined && newQty > availableStock) {
      toast.error(t("unitsAvailable", { count: availableStock }));
      return;
    }

    try {
      setBusyItemId(itemId);
      const res = await updateCartItem(itemId, newQty);
      if (!res.success) {
        toast.error(res.error || "Failed to update item");
      }
    } finally {
      setBusyItemId(null);
    }
  };

  const handleRemove = async (itemId: string) => {
    try {
      setBusyItemId(itemId);
      const res = await removeFromCart(itemId);
      if (res.success) {
        toast.success(t("removed"));
      } else {
        toast.error(res.error || "Failed to remove item");
      }
    } finally {
      setBusyItemId(null);
    }
  };

  return (
    <div className="w-full flex flex-col gap-4 ltr:md:border-l rtl:md:border-r border-border/70 ltr:md:pl-8 rtl:md:pr-8">
      {/* Header */}
      <div className="flex items-center justify-between pb-3 border-b border-border/60">
        <div className="flex items-center gap-3">
          <div className="size-9 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center shrink-0">
            <ShoppingBag className="size-4.5" />
          </div>
          <div className="flex flex-col">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-sm text-foreground">
                {t("title")}
              </h3>
              <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-accent-brand/10 text-accent-brand">
                {cartItems.reduce((acc, item) => acc + (item.quantity || 1), 0)}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
          </div>
        </div>
      </div>

      {/* Cart Items List */}
      <div className="flex flex-col gap-2 max-h-[380px] overflow-y-auto ltr:pr-1 rtl:pl-1">
        {cartItems.map((item) => {
          const name =
            typeof item.productName === "object"
              ? locale === "ar"
                ? item.productName.ar || item.productName.en
                : item.productName.en || item.productName.ar
              : item.productName;

          const price =
            typeof item.productPrice === "number"
              ? item.productPrice
              : parseFloat(String(item.productPrice || "").replace(/[^0-9.-]+/g, "")) || 0;

          const itemTotal = price * (item.quantity || 1);
          const isBusy = busyItemId === item.id;

          return (
            <div
              key={item.id}
              className="flex items-center gap-2.5 p-2 rounded-xl bg-secondary/30 border border-border/40 hover:bg-secondary/50 transition-colors"
            >
              {/* Product Thumbnail */}
              <div className="size-11 rounded-lg bg-background border border-border/60 overflow-hidden shrink-0 flex items-center justify-center p-1">
                <LazyImage
                  src={item.img || "/placeholder-product.png"}
                  alt={name || "Product"}
                  width={44}
                  height={44}
                  className="object-contain size-full"
                />
              </div>

              {/* Title & Brand/Color in one tight block */}
              <div className="flex-1 min-w-0">
                <h4 className="text-xs font-semibold text-foreground line-clamp-1 break-words">
                  {name}
                </h4>
                <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground mt-0.5">
                  {item.productBrand && <span className="truncate max-w-[80px]">{item.productBrand}</span>}
                  {item.color && (
                    <>
                      <span>•</span>
                      <span>{item.color}</span>
                    </>
                  )}
                </div>
              </div>

              {/* Compact Quantity Stepper */}
              <div className="flex items-center bg-background border border-border/70 rounded-md p-0.5 shadow-2xs shrink-0">
                <button
                  type="button"
                  onClick={() => handleUpdateQty(item.id, item.quantity || 1, -1, item.availableStock)}
                  disabled={isBusy || (item.quantity || 1) <= 1}
                  className="size-5 inline-flex items-center justify-center rounded text-foreground hover:bg-secondary transition-colors cursor-pointer disabled:opacity-25 disabled:cursor-not-allowed"
                  aria-label="Decrease quantity"
                >
                  <Minus className="size-2.5" />
                </button>

                <span className="min-w-5 text-center text-[11px] font-semibold text-foreground px-0.5">
                  {isBusy ? <Loader2 className="size-2.5 animate-spin mx-auto text-accent-brand" /> : item.quantity || 1}
                </span>

                <button
                  type="button"
                  onClick={() => handleUpdateQty(item.id, item.quantity || 1, 1, item.availableStock)}
                  disabled={isBusy || (item.availableStock !== undefined && (item.quantity || 1) >= item.availableStock)}
                  className="size-5 inline-flex items-center justify-center rounded text-foreground hover:bg-secondary transition-colors cursor-pointer disabled:opacity-25 disabled:cursor-not-allowed"
                  aria-label="Increase quantity"
                >
                  <Plus className="size-2.5" />
                </button>
              </div>

              {/* Item Price */}
              <div className="min-w-[56px] text-xs font-bold text-foreground text-right rtl:text-left shrink-0">
                ${itemTotal.toFixed(2)}
              </div>

              {/* Quick Remove Button */}
              <button
                type="button"
                onClick={() => handleRemove(item.id)}
                disabled={isBusy}
                className="p-1 rounded-md text-muted-foreground/70 hover:text-destructive hover:bg-destructive/10 transition-colors cursor-pointer disabled:opacity-30 shrink-0"
                aria-label={t("remove")}
                title={t("remove")}
              >
                <Trash2 className="size-3.5" />
              </button>
            </div>
          );
        })}
      </div>

      {/* Subtotal & Switch */}
      <div className="pt-3 border-t border-border/60 flex flex-col gap-3.5">
        <div className="flex items-center justify-between text-xs px-1">
          <span className="text-muted-foreground font-medium">{t("total")}</span>
          <span className="text-sm font-bold text-foreground">
            ${cartSummary.total > 0 ? cartSummary.total.toFixed(2) : cartSummary.subtotal.toFixed(2)}
          </span>
        </div>

        {/* Merge Switch Card */}
        <div className="flex items-center justify-between gap-3 p-3.5 rounded-xl bg-secondary/50 border border-border">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
              <Sparkles className="size-3.5 text-accent-brand shrink-0" />
              <span>{t("mergeLabel")}</span>
            </div>
            <p className="text-[11px] text-muted-foreground mt-0.5 leading-tight">
              {t("mergeDescription")}
            </p>
          </div>
          <div className="shrink-0">
            <Toggle
              isEnabled={mergeCart}
              onChange={(e) => onToggleMerge(e.target.checked)}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
