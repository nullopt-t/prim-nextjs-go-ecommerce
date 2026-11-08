"use client";

import React, { useState } from "react";
import Image, { StaticImageData } from "next/image";
import { Link } from "@/shared/i18n/navigation";
import { Stars } from "@/shared/ui/stars";
import { Heart, ShoppingBag, Check } from "lucide-react";
import { useTranslations, useLocale } from "next-intl";
import { useCartContext } from "@/modules/cart/model/cartContext";
import { useWishlist } from "@/modules/wishlist/model/useWishlist";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { CatalogImage } from "@/shared/ui/catalogImage";
import { usePriceFormatter } from "@/modules/settings/model/usePriceFormatter";

export interface CardDetails {
  id: string | number;
  productId?: string;
  wishlistItemId?: string;
  variantId?: string;
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
  variants?: unknown[];
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
  const tCard = useTranslations("components");
  const locale = useLocale();
  const formatPrice = usePriceFormatter();
  const { addToCart } = useCartContext();
  const { toggleWishlist } = useWishlist();

  const [isAdding, setIsAdding] = useState(false);
  const [justAdded, setJustAdded] = useState(false);
  const targetProductId = String(cardDetails.productId || cardDetails.id);
  const productName =
    typeof cardDetails.product === "object"
      ? (locale === "ar" ? cardDetails.product.ar || cardDetails.product.en : cardDetails.product.en || cardDetails.product.ar)
      : String(cardDetails.product || "—");

  const baseProductTarget = cardDetails.slug || String(cardDetails.id);
  const productHref = cardDetails.variantId
    ? `/products/${baseProductTarget}?variant=${encodeURIComponent(cardDetails.variantId)}`
    : `/products/${baseProductTarget}`;

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
      const variantList = (cardDetails.variants ?? []) as Array<{ id?: string; isDefault?: boolean }>;
      const defaultVariant = variantList.find((v) => v.isDefault) ?? variantList[0] ?? null;
      const payload = {
        id: String(cardDetails.id),
        slug: cardDetails.slug || String(cardDetails.id),
        variantId: defaultVariant?.id || cardDetails.defaultVariantId,
        productName: cardDetails.product,
        productPrice: formatPrice(cardDetails.price),
        img: typeof cardDetails.img === "string" ? normalizeMediaUrl(cardDetails.img) : "/placeholder-product.png",
        quantity: 1,
      };

      const res = await addToCart(payload);
      if (res.success) {
        setJustAdded(true);
        notify.success(tCard("productsCard.addedToCart", { productName }));
        setTimeout(() => setJustAdded(false), 2000);
      } else {
        notify.error(apiErrorMessage(res, tCard("productsCard.addToCartFailed")));
      }
    } catch (err) {
      notify.error(apiErrorMessage(err, tCard("productsCard.addToCartFailed")));
    } finally {
      setIsAdding(false);
    }
  };

  const handleWishlistClick = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (isWishlist && onRemove) {
      onRemove();
      return;
    }

    await toggleWishlist(targetProductId, productName);
  };

  return (
    <div className="group relative flex flex-col justify-between h-full rounded-2xl border border-border bg-card overflow-hidden shadow-2xs hover:shadow-md hover:border-accent-brand/60 transition-all duration-200">
      <div>
        {/* Thumbnail & Badges */}
        <div className="relative aspect-square w-full bg-secondary/40 overflow-hidden">
          <Link
            href={productHref}
            className="block w-full h-full"
          >
            {typeof cardDetails.img === "string" ? (
              <CatalogImage
                src={cardDetails.img}
                alt={productName}
                fill
                className="object-cover object-center group-hover:scale-105 transition-transform duration-300"
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
            <span className="absolute top-2.5 start-2.5 bg-[#d4183d] text-white rounded-md px-1.5 py-0.5 text-[11px] font-bold shadow-xs">
              {cardDetails.discountPercentage}
            </span>
          )}

          {/* Wishlist Button (Only displayed on wishlist page for removal if needed, hidden on general listings) */}
          {isWishlist && (
            <button
              type="button"
              onClick={handleWishlistClick}
              aria-label={tCommon("header.wishlist")}
              className="absolute top-2.5 end-2.5 flex size-11 items-center justify-center rounded-full transition-all shadow-xs cursor-pointer bg-accent-brand text-white hover:bg-accent-brand/90 scale-105"
            >
              <Heart className="size-5 fill-current" />
            </button>
          )}
        </div>

        {/* Content details */}
        <div className="p-2.5 sm:p-3 text-start">
          <Link
            href={productHref}
            className="block group/title"
          >
            <h4 dir="auto" className="font-semibold text-xs sm:text-sm text-foreground group-hover/title:text-accent-brand transition-colors line-clamp-1">
              {productName}
            </h4>
          </Link>

          {/* Ratings: wrap at 320px so the count never clips the stars */}
          <div className="flex flex-wrap items-center gap-1.5 mt-1 mb-2">
            <Stars starsNum={cardDetails.stars} />
            <span className="text-muted-foreground text-[10px] sm:text-[11px]">
              ({cardDetails.reviews})
            </span>
          </div>

          {/* Pricing */}
          <div className="flex flex-wrap items-baseline gap-1.5">
            <span className="font-bold text-sm sm:text-base text-foreground">
              {formatPrice(cardDetails.price)}
            </span>
            {cardDetails.oldPrice && (
              <del className="text-muted-foreground text-[11px]">
                {formatPrice(cardDetails.oldPrice)}
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
          className={`w-full min-h-11 rounded-xl font-medium text-xs flex items-center justify-center gap-1.5 transition-all cursor-pointer shadow-xs disabled:opacity-60 ${
            justAdded
              ? "bg-emerald-600 text-white"
              : "bg-primary text-primary-foreground hover:opacity-90 active:scale-[0.98]"
          }`}
        >
          {justAdded ? (
            <>
              <Check className="size-3.5" />
              <span>{tCard("productsCard.added")}</span>
            </>
          ) : (
            <>
              <ShoppingBag className="size-3.5" />
              <span>{isAdding ? tCard("productsCard.adding") : tHome("product.addToCart")}</span>
            </>
          )}
        </button>

        {isWishlist && (
          <button
            type="button"
            className="w-full text-center text-xs text-accent-brand font-medium mt-2 hover:underline cursor-pointer"
            onClick={onRemove}
          >
            {tCard("productsCard.remove")}
          </button>
        )}
      </div>
    </div>
  );
}

export default ProductsCard;
