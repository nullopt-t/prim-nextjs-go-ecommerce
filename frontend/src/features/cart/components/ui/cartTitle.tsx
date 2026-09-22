"use client";

import { useTranslations } from "next-intl";
import { useCart } from "@/hooks/useCart";

export default function CartTitle() {
  const t = useTranslations("cart");
  const { cartItems } = useCart();
  const count = cartItems?.length || 0;

  return (
    <div className="flex flex-col gap-1">
      <h1 className="font-bold text-foreground text-3xl md:text-4xl tracking-tight">
        {t("cart.title", { defaultMessage: "Your Cart" })}
      </h1>
      {count > 0 && (
        <p className="text-muted-foreground text-sm sm:text-base">
          {count === 1
            ? t("cart.item", { count: 1, defaultMessage: "1 item in your cart" })
            : t("cart.items", { count, defaultMessage: `${count} items in your cart` })}
        </p>
      )}
    </div>
  );
}
