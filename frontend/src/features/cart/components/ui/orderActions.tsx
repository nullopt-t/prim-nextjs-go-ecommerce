"use client";

import { QuantitySelector } from "@/components/ui/quantitySelector";
import { useTranslations } from "next-intl";

interface OrderActionsProps {
  price?: string | number;
  onRemove?: () => void;
  onQuantityChange?: (quantity: number) => void;
}

export default function OrderActions({
  price = "$45",
  onRemove,
  onQuantityChange,
}: OrderActionsProps) {
  const t = useTranslations("cart");

  return (
    <div className="flex flex-col gap-2.5 items-end justify-between">
      <QuantitySelector onChange={onQuantityChange} />
      <p className="text-foreground font-medium text-txt-sm md:text-txt-md lg:text-txt-lg">
        {price}
      </p>
      <button
        type="button"
        onClick={onRemove}
        className="text-accent-brand hover:underline text-txt-sm cursor-pointer"
      >
        {t("product.remove")}
      </button>
    </div>
  );
}
