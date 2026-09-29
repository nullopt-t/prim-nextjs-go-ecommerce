"use client";

import { X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useCart } from "@/hooks/useCart";
import OrdersGrid from "@/features/cart/components/ui/ordersGrid";
import PaymentBox from "@/features/cart/components/ui/paymentBox";
import { createPortal } from "react-dom";
import { useEffect, useState } from "react";

export function CartSidebar() {
  const t = useTranslations("cart");
  const { cartItems, isCartOpen, closeCart } = useCart();
  const [mounted, setMounted] = useState(false);
  const hasItems = cartItems && cartItems.length > 0;

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (isCartOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "unset";
    }

    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isCartOpen]);

  if (!mounted) return null;

  return createPortal(
    <>
      {/* Backdrop */}
      {isCartOpen && (
        <div
          className="fixed inset-0 h-[100dvh] z-[100] bg-black/60 backdrop-blur-xs transition-opacity duration-300"
          onClick={closeCart}
        />
      )}

      {/* Sidebar Drawer */}
      <div
        className={`fixed top-0 bottom-0 h-[100dvh] z-[100] w-full sm:w-[420px] max-w-full bg-background border-l border-border shadow-2xl transition-transform duration-300 ease-out flex flex-col ltr:right-0 rtl:left-0 rtl:border-r rtl:border-l-0 ${
          isCartOpen
            ? "translate-x-0"
            : "ltr:translate-x-full rtl:-translate-x-full"
        }`}
      >
        <div className="flex items-center justify-between p-4 sm:p-5 border-b border-border bg-card/60 backdrop-blur-sm">
          <div className="flex items-baseline gap-2">
            <h2 className="text-xl font-bold text-foreground">
              {t("cart.title", { defaultMessage: "Your Cart" })}
            </h2>
            {cartItems && cartItems.length > 0 && (
              <span className="text-xs text-muted-foreground font-medium">
                ({cartItems.length})
              </span>
            )}
          </div>
          <button
            type="button"
            onClick={closeCart}
            aria-label="Close cart"
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-secondary rounded-full transition-colors cursor-pointer"
          >
            <X className="size-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 sm:p-5 flex flex-col gap-4">
          <OrdersGrid />
        </div>

        {hasItems && (
          <div className="p-3.5 sm:p-4 border-t border-border/80 bg-background/95 backdrop-blur-md shadow-lg">
            <PaymentBox compact={true} />
          </div>
        )}
      </div>
    </>,
    document.body
  );
}

export default CartSidebar;
