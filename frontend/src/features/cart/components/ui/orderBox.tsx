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

  const [isUpdating, setIsUpdating] = useState(false);

  // Maximum allowed is availableStock from inventory (or fallback 99 if unbounded)
  const maxStock = orderDetails.availableStock !== undefined 
    ? Math.max(1, orderDetails.availableStock) 
    : 99;
  const isOutOfStock = orderDetails.inStock === false || (orderDetails.availableStock !== undefined && orderDetails.availableStock <= 0);

  const handleQuantityChange = async (newQty: number) => {
    if (newQty <= 0) {
      handleRemove();
      return;
    }

    if (orderDetails.availableStock !== undefined && newQty > orderDetails.availableStock) {
      toast.error(`Only ${orderDetails.availableStock} units available in stock`);
      return;
    }

    setIsUpdating(true);
    const res = await updateCartItem(orderDetails.id, newQty);
    setIsUpdating(false);
    if (!res.success) {
      toast.error(res.error || "Failed to update quantity");
    }
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
      <div className="w-20 sm:w-24 shrink-0 rounded-xl overflow-hidden aspect-square border border-border/50 bg-secondary/20 relative">
        <LazyImage
          src={imgSrc}
          alt={productName}
          containerClassName="w-full h-full"
          imageClassName="rounded-xl object-cover object-center w-full h-full"
        />
        {isOutOfStock && (
          <div className="absolute inset-0 bg-background/80 backdrop-blur-2xs flex items-center justify-center p-1 text-center">
            <span className="text-[10px] font-bold text-destructive leading-tight uppercase">
              Out of stock
            </span>
          </div>
        )}
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

            {/* Stock indicator badge if low or maxed out */}
            {orderDetails.availableStock !== undefined && orderDetails.availableStock > 0 && orderDetails.availableStock <= 5 && (
              <span className="text-[11px] text-amber-600 dark:text-amber-400 font-medium mt-1">
                Only {orderDetails.availableStock} left in stock
              </span>
            )}
          </div>
          <p className="text-foreground font-semibold text-base shrink-0">
            {orderDetails.productPrice}
          </p>
        </div>

        <div className="flex justify-between items-center mt-3 pt-3 border-t border-border/40">
          <button
            type="button"
            onClick={handleRemove}
            disabled={isDeleting || isUpdating}
            className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground hover:text-destructive transition-colors cursor-pointer"
          >
            <Trash2 className="size-3.5" />
            <span>{t("product.remove", { defaultMessage: "Remove" })}</span>
          </button>
          <div className="w-24 sm:w-28 h-8 sm:h-9">
            <QuantitySelector
              value={orderDetails.quantity || 1}
              max={maxStock}
              disabled={isDeleting || isUpdating || isOutOfStock}
              onChange={handleQuantityChange}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
