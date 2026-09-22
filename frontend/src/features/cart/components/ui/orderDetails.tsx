"use client";

import { useLocale } from "next-intl";
import { CartItemData } from "@/features/cart/types";

export default function OrderDetails({ details }: { details: CartItemData }) {
  const locale = useLocale();

  const displayName = typeof details.productName === "string"
    ? details.productName
    : (locale === "ar" ? (details.productName?.ar || details.productName?.en) : (details.productName?.en || details.productName?.ar)) || "Product";

  return (
    <div className="flex gap-4 items-center">
      <div className="rounded-md bg-secondary aspect-square w-20 md:w-24 shrink-0 flex items-center justify-center font-bold text-muted-foreground text-xs">
        PRIM
      </div>
      <div>
        <p className="font-medium text-foreground text-txt-sm md:text-txt-md mb-1">
          {displayName}
        </p>
        <p className="text-muted-foreground text-xs md:text-sm mb-2">
          {details.productBrand}
        </p>
        <p className="font-medium text-foreground text-txt-sm md:text-txt-md">
          {details.productPrice}
        </p>
      </div>
    </div>
  );
}
