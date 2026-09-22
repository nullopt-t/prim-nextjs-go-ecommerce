"use client";

import { QuantitySelector } from "@/components/ui/quantitySelector";
import { useTranslations } from "next-intl";
import { Trash2 } from "lucide-react";
import { LazyImage } from "@/components/ui/lazyImage";
import { CartItemData } from "@/features/cart/types";
import { useCart } from "@/hooks/useCart";
import { toast } from "sonner";
import { useState } from "react";

export default function OrderBox({ orderDetails }: { orderDetails: CartItemData }) {
  const t = useTranslations("cart");
  const { updateCartItem, removeFromCart } = useCart();
  const [isDeleting, setIsDeleting] = useState(false);

  const productName = typeof orderDetails.productName === "object"
    ? (orderDetails.productName?.en || orderDetails.productName?.ar || "Product")
    : (orderDetails.productName || "Product");

  const imgSrc = orderDetails.img || "/placeholder-product.png";

  const handleQuantityChange = async (newQty: number) => {
    if (newQty <= 0) {
      handleRemove();
      return;
    }
    await updateCartItem(orderDetails.id, newQty);
  };

  const handleRemove = async () => {
    setIsDeleting(true);
    const res = await removeFromCart(orderDetails.id);
    setIsDeleting(false);
    if (res.success) {
      toast.success(`${productName} removed from cart`);
    } else {
      toast.error(res.error || "Failed to remove item");
    }
  };

  return (
    <div className={`flex gap-4 border border-border rounded-2xl p-4 bg-card hover:border-accent-brand/50 transition-all ${isDeleting ? "opacity-50 pointer-events-none" : ""}`}>
      <div className="w-20 sm:w-24 shrink-0 rounded-xl overflow-hidden aspect-square border border-border/50 bg-secondary/20">
        <LazyImage
          src={imgSrc}
          alt={productName}
          containerClassName="w-full h-full"
          imageClassName="rounded-xl object-cover object-center w-full h-full"
        />
      </div>

      <div className="flex flex-col flex-1 min-w-0 justify-between">
        <div className="flex justify-between items-start gap-2">
          <div className="flex flex-col min-w-0">
            <p className="font-semibold text-foreground text-base truncate">
              {productName}
            </p>
            <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
              <span>{orderDetails.productBrand || "PRIM"}</span>
              {orderDetails.color && (
                <>
                  <span>•</span>
                  <span>{orderDetails.color}</span>
                </>
              )}
            </div>
          </div>
          <p className="text-foreground font-semibold text-base shrink-0">
            {orderDetails.productPrice}
          </p>
        </div>

        <div className="flex justify-between items-center mt-3 pt-3 border-t border-border/40">
          <button
            type="button"
            onClick={handleRemove}
            disabled={isDeleting}
            className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-destructive transition-colors cursor-pointer"
          >
            <Trash2 className="size-3.5" />
            <span>{t("product.remove", { defaultMessage: "Remove" })}</span>
          </button>
          <div className="w-28 h-9">
            <QuantitySelector
              initialValue={orderDetails.quantity || 1}
              onChange={handleQuantityChange}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
