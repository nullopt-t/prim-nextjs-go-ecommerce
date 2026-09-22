"use client";

import Image, { StaticImageData } from "next/image";
import { Stars } from "@/components/ui/stars";
import { CustomButton } from "@/components/ui/button";
import { Heart } from "lucide-react";
import { useTranslations, useLocale } from "next-intl";

export interface CardDetails {
  id: string | number;
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

  const productName =
    locale === "en" ? cardDetails.product.en : cardDetails.product.ar;

  return (
    <div className="shadow-xs hover:shadow-md cursor-pointer transition-all duration-300 hover:border-accent-brand/80 border border-border rounded-2xl overflow-hidden bg-card flex flex-col justify-between group">
      <div>
        <div className="relative aspect-square w-full bg-secondary/50 overflow-hidden">
          {typeof cardDetails.img === "string" ? (
            <img
              src={cardDetails.img}
              alt={productName}
              className="object-center object-cover w-full h-full"
            />
          ) : (
            <Image
              src={cardDetails.img}
              alt={productName}
              fill
              className="object-cover object-center"
            />
          )}
          <button
            type="button"
            aria-label="Wishlist"
            className={`absolute flex items-center justify-center top-2.5 right-2.5 rounded-full w-8 h-8 transition-colors ${
              isWishlist
                ? "bg-accent-brand text-white hover:bg-background hover:text-accent-brand"
                : "bg-background text-foreground hover:bg-accent-brand hover:text-white"
            }`}
          >
            <Heart className="size-4" />
          </button>
        </div>

        <div className="p-2.5">
          <p className="font-medium mb-1 text-card-foreground line-clamp-1">
            {productName}
          </p>
          <div className="flex items-center gap-2.5 mb-2.5">
            <Stars starsNum={cardDetails.stars} />
            <span className="text-muted-foreground text-xs">
              ({cardDetails.reviews})
            </span>
          </div>
          <div className="flex flex-wrap gap-2 items-center">
            <span className="font-medium text-title-sm text-card-foreground">
              {tCommon("currency")}&nbsp;{cardDetails.price}
            </span>
            {cardDetails.oldPrice && (
              <del className="text-muted-foreground text-xs">
                {tCommon("currency")}&nbsp;{cardDetails.oldPrice}
              </del>
            )}
            {cardDetails.discountPercentage && (
              <span className="bg-[#d4183d] text-white rounded px-1 text-xs">
                {cardDetails.discountPercentage}
              </span>
            )}
          </div>
        </div>
      </div>

      <div>
        <div className="h-10 mx-3 mb-3 text-primary-foreground bg-primary rounded-xl overflow-hidden hover:opacity-90 transition-opacity">
          <CustomButton
            text={tHome("product.addToCart")}
            onClick={onAddToCart}
            className="rounded-xl border-none font-semibold text-xs tracking-wide"
          />
        </div>
        {isWishlist && (
          <button
            type="button"
            className="w-full text-center text-accent-brand font-medium mb-2.5 hover:underline"
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
