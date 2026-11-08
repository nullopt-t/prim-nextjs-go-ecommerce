"use client";

import { X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useCart } from "@/modules/cart/model/useCart";
import OrdersGrid from "@/modules/cart/components/ui/ordersGrid";
import PaymentBox from "@/modules/cart/components/ui/paymentBox";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { ProductsCard } from "@/shared/ui/productsCard";
import type { CardDetails } from "@/shared/ui/productsCard";
import {
  selectCartRecommendations,
  type RelatedCandidate,
} from "@/modules/catalog/utils/relatedProducts";
import { createPortal } from "react-dom";
import { useEffect, useMemo, useState } from "react";

export function CartSidebar() {
  const t = useTranslations("cart");
  const tUi = useTranslations("components");
  const { cartItems, isCartOpen, closeCart } = useCart();
  const { products } = useCatalogContext();
  const [mounted, setMounted] = useState(false);
  const hasItems = cartItems && cartItems.length > 0;

  const recommendations = useMemo(() => {
    if (!hasItems || !products?.length) return [];
    const byId = new Map(products.map((p) => [String(p.id), p]));
    const cartAffinity = (cartItems ?? [])
      .map((item) => byId.get(String(item.productId ?? item.id)))
      .filter(Boolean) as unknown as RelatedCandidate[];
    return selectCartRecommendations(
      cartAffinity,
      products as unknown as RelatedCandidate[],
      6,
    );
  }, [cartItems, hasItems, products]);

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

  useEffect(() => {
    if (!isCartOpen) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") closeCart();
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isCartOpen, closeCart]);

  if (!mounted) return null;

  return createPortal(
    <>
      {/* Backdrop */}
      {isCartOpen && (
        <button
          type="button"
          aria-label={tUi("cart.closeCart")}
          className="fixed inset-0 h-[100dvh] z-[100] cursor-default bg-black/60 backdrop-blur-xs transition-opacity duration-300"
          onClick={closeCart}
        />
      )}

      {/* Sidebar Drawer */}
      <div
        role="dialog"
        aria-modal={isCartOpen}
        aria-label={t("cart.title", { defaultMessage: "Your Cart" })}
        aria-hidden={!isCartOpen}
        className={`fixed top-0 bottom-0 h-[100dvh] z-[100] w-full sm:w-[420px] max-w-full bg-background border-s border-border shadow-2xl transition-transform duration-300 ease-out flex flex-col end-0 ${
          isCartOpen
            ? "translate-x-0"
            : "translate-x-full rtl:-translate-x-full invisible"
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
            aria-label={tUi("cart.closeCart")}
            className="flex size-11 shrink-0 items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary rounded-full transition-colors cursor-pointer"
          >
            <X className="size-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 sm:p-5 flex flex-col gap-4">
          <OrdersGrid />

          {recommendations.length > 0 && (
            <section aria-label={t("youMayAlsoLike")} className="flex flex-col gap-2.5 border-t border-border/70 pt-4">
              <h3 className="text-sm font-bold text-foreground">{t("youMayAlsoLike")}</h3>
              <div className="flex gap-3 overflow-x-auto pb-1 scrollbar-none snap-x">
                {recommendations.map((item) => (
                  <div key={String(item.id)} className="w-36 sm:w-40 shrink-0 snap-start">
                    <ProductsCard
                      cardDetails={item as unknown as CardDetails}
                    />
                  </div>
                ))}
              </div>
            </section>
          )}
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
