"use client";

import React, { useState } from "react";
import Image, { StaticImageData } from "next/image";
import { Link } from "@/i18n/navigation";
import { Stars } from "@/components/ui/stars";
import { Heart, ShoppingBag, Check } from "lucide-react";
import { useTranslations, useLocale } from "next-intl";
import { useCartContext } from "@/context/CartContext";
import { toast } from "sonner";

export interface CardDetails {
  id: string | number;
  slug?: string;
  img: string | StaticImageData;
  product: {
    en: string;
    ar: string;
  };
  stars: number | string;
  reviews: number | string;
  price: number | string;
  oldPrice?: number | string;
  discountPercentage?: string;
  variants?: any[];
  /** Fallback variant ID used when no variants array is present */
  defaultVariantId?: string;
  brand?: string;
  inStock?: boolean;
}

interface ProductsCardProps {
  cardDetails: CardDetails;
  isWishlist?: boolean;
  onRemove?: () => void;
  onAddToCart?: () => void;
}

export function ProductsCard({
  cardDetails,
  isWishlist = false,
  onRemove,
  onAddToCart,
}: ProductsCardProps) {
  const tHome = useTranslations("home");
  const tCommon = useTranslations("common");
  const locale = useLocale();
  const { addToCart } = useCartContext();

  const [isAdding, setIsAdding] = useState(false);
  const [justAdded, setJustAdded] = useState(false);
  const [isLiked, setIsLiked] = useState(isWishlist);

  const productName =
    typeof cardDetails.product === "object"
      ? (locale === "ar" ? cardDetails.product.ar || cardDetails.product.en : cardDetails.product.en || cardDetails.product.ar)
      : String(cardDetails.product || "Product");

  const productTarget = cardDetails.slug || String(cardDetails.id);

  const handleCartClick = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (onAddToCart) {
      onAddToCart();
      return;
    }

    if (isAdding) return;
    setIsAdding(true);

    try {
      const defaultVariant = (cardDetails.variants && cardDetails.variants[0]) || null;
      const payload = {
        id: String(cardDetails.id),
        slug: cardDetails.slug || String(cardDetails.id),
        variantId: defaultVariant?.id || cardDetails.defaultVariantId,
        productName: cardDetails.product,
        productPrice: typeof cardDetails.price === "number" ? `$${cardDetails.price.toFixed(2)}` : String(cardDetails.price),
        img: typeof cardDetails.img === "string" ? cardDetails.img : "/placeholder-product.png",
        quantity: 1,
      };

      const res = await addToCart(payload);
      if (res.success) {
        setJustAdded(true);
        toast.success(
          locale === "ar"
            ? `تمت إضافة "${productName}" إلى سلة التسوق!`
            : `Added "${productName}" to your cart!`
        );
        setTimeout(() => setJustAdded(false), 2000);
      } else {
        toast.error(res.error || (locale === "ar" ? "فشل إضافة المنتج إلى السلة" : "Failed to add to cart"));
      }
    } catch (err: any) {
      toast.error(err?.message || "Failed to add to cart");
    } finally {
      setIsAdding(false);
    }
  };

  const handleWishlistClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (isWishlist && onRemove) {
      onRemove();
      return;
    }

    const nextState = !isLiked;
    setIsLiked(nextState);
    if (nextState) {
      toast.success(locale === "ar" ? `تمت الإضافة للمفضلة` : `Saved to wishlist!`);
    } else {
      toast.info(locale === "ar" ? `تمت الإزالة من المفضلة` : `Removed from wishlist`);
    }
  };

  return (
    <div className="group relative flex flex-col justify-between h-full rounded-2xl border border-border bg-card overflow-hidden shadow-2xs hover:shadow-md hover:border-accent-brand/60 transition-all duration-200">
      <div>
        {/* Thumbnail & Badges */}
        <div className="relative aspect-square w-full bg-secondary/40 overflow-hidden">
          <Link
            href={`/products/${productTarget}`}
            className="block w-full h-full"
          >
            {typeof cardDetails.img === "string" ? (
              <img
                src={cardDetails.img}
                alt={productName}
                className="object-cover object-center w-full h-full group-hover:scale-105 transition-transform duration-300"
              />
            ) : (
              <Image
                src={cardDetails.img}
                alt={productName}
                fill
                className="object-cover object-center group-hover:scale-105 transition-transform duration-300"
              />
            )}
          </Link>

          {/* Discount Badge */}
          {cardDetails.discountPercentage && (
            <span className="absolute top-2.5 ltr:left-2.5 rtl:right-2.5 bg-[#d4183d] text-white rounded-md px-1.5 py-0.5 text-[10px] font-bold shadow-xs">
              {cardDetails.discountPercentage}
            </span>
          )}

          {/* Wishlist Button */}
          <button
            type="button"
            onClick={handleWishlistClick}
            aria-label="Wishlist"
            className={`absolute top-2.5 ltr:right-2.5 rtl:left-2.5 flex items-center justify-center rounded-full size-7.5 transition-all shadow-xs cursor-pointer ${
              isLiked
                ? "bg-accent-brand text-white hover:bg-accent-brand/90 scale-105"
                : "bg-background/90 text-foreground hover:bg-accent-brand hover:text-white backdrop-blur-xs"
            }`}
          >
            <Heart className={`size-3.5 ${isLiked ? "fill-current" : ""}`} />
          </button>
        </div>

        {/* Content details */}
        <div className="p-2.5 sm:p-3 text-left rtl:text-right">
          <Link
            href={`/products/${productTarget}`}
            className="block group/title"
          >
            <h4 className="font-semibold text-xs sm:text-sm text-foreground group-hover/title:text-accent-brand transition-colors line-clamp-1">
              {productName}
            </h4>
          </Link>

          {/* Ratings */}
          <div className="flex items-center gap-1.5 mt-1 mb-2">
            <Stars starsNum={cardDetails.stars} />
            <span className="text-muted-foreground text-[10px] sm:text-[11px]">
              ({cardDetails.reviews})
            </span>
          </div>

          {/* Pricing */}
          <div className="flex flex-wrap items-baseline gap-1.5">
            <span className="font-bold text-sm sm:text-base text-foreground">
              {tCommon("currency")}&nbsp;
              {typeof cardDetails.price === "number"
                ? cardDetails.price.toFixed(2)
                : cardDetails.price}
            </span>
            {cardDetails.oldPrice && (
              <del className="text-muted-foreground text-[11px]">
                {tCommon("currency")}&nbsp;
                {typeof cardDetails.oldPrice === "number"
                  ? cardDetails.oldPrice.toFixed(2)
                  : cardDetails.oldPrice}
              </del>
            )}
          </div>
        </div>
      </div>

      {/* Action Footer */}
      <div className="p-2.5 sm:p-3 pt-0">
        <button
          type="button"
          onClick={handleCartClick}
          disabled={isAdding}
          className={`w-full h-8 sm:h-9 rounded-xl font-medium text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer shadow-xs disabled:opacity-60 ${
            justAdded
              ? "bg-emerald-600 text-white"
              : "bg-primary text-primary-foreground hover:opacity-90 active:scale-[0.98]"
          }`}
        >
          {justAdded ? (
            <>
              <Check className="size-3.5" />
              <span>Added</span>
            </>
          ) : (
            <>
              <ShoppingBag className="size-3.5" />
              <span>{isAdding ? "Adding..." : tHome("product.addToCart")}</span>
            </>
          )}
        </button>

        {isWishlist && (
          <button
            type="button"
            className="w-full text-center text-xs text-accent-brand font-medium mt-2 hover:underline cursor-pointer"
            onClick={onRemove}
          >
            Remove
          </button>
        )}
      </div>
    </div>
  );
}

export default ProductsCard;
