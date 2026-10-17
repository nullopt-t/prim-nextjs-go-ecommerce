"use client";

import { QuantitySelector } from "@/components/ui/quantitySelector";
import { useTranslations } from "next-intl";
import { Trash2 } from "lucide-react";
import { LazyImage } from "@/components/ui/lazyImage";
import { CartItemData } from "@/features/cart/types";
import { useCart } from "@/hooks/useCart";
import { Link } from "@/i18n/navigation";
import { notify, apiErrorMessage } from "@/lib/notify";
import { useState } from "react";

export default function OrderBox({ orderDetails }: { orderDetails: CartItemData }) {
  const t = useTranslations("cart");
  const { updateCartItem, removeFromCart, closeCart } = useCart();
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
      notify.error(t("toast.unitsAvailable", { count: orderDetails.availableStock }));
      return;
    }

    setIsUpdating(true);
    const res = await updateCartItem(orderDetails.id, newQty);
    setIsUpdating(false);
    if (!res.success) {
      notify.error(apiErrorMessage(res, t("toast.updateQuantityFailed")));
    }
  };

  const handleRemove = async () => {
    setIsDeleting(true);
    const res = await removeFromCart(orderDetails.id);
    setIsDeleting(false);
    if (res.success) {
      notify.success(t("toast.itemRemoved", { name: productName }));
    } else {
      notify.error(apiErrorMessage(res, t("toast.removeFailed")));
    }
  };

  const productSlug = orderDetails.productSlug || orderDetails.productId || orderDetails.id;
  const productHref = productSlug
    ? `/products/${productSlug}?${new URLSearchParams({
        ...(orderDetails.variantId ? { variant: orderDetails.variantId } : {}),
        ...(orderDetails.color ? { color: orderDetails.color } : {}),
      }).toString()}`
    : "/products";

  return (
    <div
      className={`group flex items-center gap-3 p-3 sm:p-3.5 rounded-xl bg-secondary/30 border border-border/50 hover:bg-secondary/50 transition-colors shadow-2xs ${
        isDeleting ? "opacity-50 pointer-events-none" : ""
      }`}
    >
      {/* Product Thumbnail */}
      <Link
        href={productHref}
        onClick={closeCart}
        className="size-16 sm:size-18 rounded-lg bg-background border border-border/60 overflow-hidden shrink-0 flex items-center justify-center p-1.5 relative group/img cursor-pointer shadow-2xs"
        aria-label={`View ${productName}`}
      >
        <LazyImage
          src={imgSrc}
          alt={productName}
          containerClassName="w-full h-full"
          imageClassName="rounded-md object-contain size-full group-hover/img:scale-105 transition-transform duration-200"
        />
        {isOutOfStock && (
          <div className="absolute inset-0 bg-background/80 backdrop-blur-2xs flex items-center justify-center p-1 text-center">
            <span className="text-[9px] font-bold text-destructive leading-tight uppercase">
              {t("outOfStock")}
            </span>
          </div>
        )}
      </Link>

      {/* Middle: Details & Controls */}
      <div className="flex-1 min-w-0 flex flex-col justify-between self-stretch py-0.5">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <Link
              href={productHref}
              onClick={closeCart}
              className="text-xs sm:text-sm font-semibold text-foreground line-clamp-1 hover:text-accent-brand transition-colors cursor-pointer"
            >
              {productName}
            </Link>
            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground mt-0.5">
              <span>{orderDetails.productBrand || "PRIM"}</span>
              {orderDetails.color && (
                <>
                  <span>•</span>
                  <span>{orderDetails.color}</span>
                </>
              )}
            </div>

            {/* Stock indicator badge if low */}
            {orderDetails.availableStock !== undefined && orderDetails.availableStock > 0 && orderDetails.availableStock <= 5 && (
              <span className="text-[10px] text-amber-600 dark:text-amber-400 font-medium inline-block mt-0.5">
                Only {orderDetails.availableStock} left in stock
              </span>
            )}
          </div>

          {/* Price */}
          <div className="text-xs sm:text-sm font-bold text-foreground shrink-0 text-right rtl:text-left">
            {orderDetails.productPrice}
          </div>
        </div>

        {/* Bottom Actions: Remove & Stepper */}
        <div className="flex items-center justify-between gap-2 mt-2 pt-1.5 border-t border-border/30">
          <button
            type="button"
            onClick={handleRemove}
            disabled={isDeleting || isUpdating}
            className="flex items-center gap-1 text-[11px] font-medium text-muted-foreground hover:text-destructive transition-colors cursor-pointer disabled:opacity-50"
          >
            <Trash2 className="size-3" />
            <span>{t("product.remove", { defaultMessage: "Remove" })}</span>
          </button>

          <div className="w-22 sm:w-24 h-7 sm:h-7.5">
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
