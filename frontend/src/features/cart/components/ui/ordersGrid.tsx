"use client";

import OrderBox from "@/features/cart/components/ui/orderBox";
import { useCart } from "@/hooks/useCart";
import { ShoppingBag, ArrowRight } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";

export default function OrdersGrid() {
  const { cartItems: products, loading, errorMsg } = useCart();
  const t = useTranslations("cart");

  if (loading) {
    return (
      <div className="flex flex-col gap-4">
        {[...Array(2)].map((_, i) => (
          <div key={i} className="h-28 bg-secondary/40 rounded-2xl animate-pulse" />
        ))}
      </div>
    );
  }

  if (errorMsg) {
    return <div className="text-center text-destructive py-16">{errorMsg}</div>;
  }

  if (!products || products.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 px-6 bg-card border border-border rounded-2xl text-center">
        <div className="w-16 h-16 rounded-2xl bg-secondary flex items-center justify-center text-muted-foreground mb-4">
          <ShoppingBag className="size-8 stroke-[1.5]" />
        </div>
        <h3 className="text-xl font-bold text-foreground mb-2">
          {t("empty.title", { defaultMessage: "Your cart is empty" })}
        </h3>
        <p className="text-muted-foreground text-sm max-w-sm mb-6">
          {t("empty.description", {
            defaultMessage:
              "Looks like you haven't added anything to your cart yet. Discover trending products and premium gear.",
          })}
        </p>
        <Link
          href="/products"
          className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-primary text-primary-foreground font-medium text-sm hover:bg-primary/90 transition-colors shadow-sm"
        >
          <span>{t("empty.action", { defaultMessage: "Start Shopping" })}</span>
          <ArrowRight className="size-4 rtl:rotate-180" />
        </Link>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4">
      {products.map((prod) => (
        <OrderBox key={prod.id} orderDetails={prod} />
      ))}
    </div>
  );
}
