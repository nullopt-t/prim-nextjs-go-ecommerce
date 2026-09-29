import { Check, Clock, Heart, Share2, ShoppingCart, Zap } from "lucide-react";
import { Stars } from "@/components/ui/stars";
import { CustomButton } from "@/components/ui";
import QuantitySelector from "@/components/ui/quantitySelector";
import { ProductColorOption } from "../../types";

interface ProductActionsProps {
  productCategory: string;
  productName: string;
  stars?: number;
  reviewsCount?: number;
  price: number;
  oldPrice?: number;
  currency: string;
  colors?: ProductColorOption[];
  selectedColor: number;
  onSelectColor: (idx: number) => void;
  availableStock: number;
  isVariantInStock: boolean;
  quantity: number;
  onQuantityChange: (val: number) => void;
  isAdding: boolean;
  isWishlisted: boolean;
  onAddToCart: () => void;
  onBuyNow: () => void;
  onToggleWishlist: () => void;
  onShare: () => void;
}

export function ProductActions({
  productCategory,
  productName,
  stars = 5,
  reviewsCount = 0,
  price,
  oldPrice,
  currency,
  colors = [],
  selectedColor,
  onSelectColor,
  availableStock,
  isVariantInStock,
  quantity,
  onQuantityChange,
  isAdding,
  isWishlisted,
  onAddToCart,
  onBuyNow,
  onToggleWishlist,
  onShare,
}: ProductActionsProps) {
  return (
    <div className="flex flex-col">
      <div className="flex flex-col gap-2 mb-4">
        <span className="text-sm font-medium text-accent-brand capitalize">
          {productCategory.replace("-", " ")}
        </span>
        <h1 className="text-3xl lg:text-4xl font-bold text-foreground">{productName}</h1>
      </div>

      <div className="flex items-center gap-4 mb-6">
        <div className="flex items-center gap-1.5">
          <Stars starsNum={stars} />
          <span className="text-sm font-medium text-foreground ml-1">{stars}</span>
        </div>
        <div className="w-1 h-1 rounded-full bg-border" />
        <span className="text-sm text-muted-foreground underline decoration-dashed cursor-pointer hover:text-foreground transition-colors">
          {reviewsCount} Reviews
        </span>
      </div>

      <div className="flex items-end gap-3 mb-6">
        <span className="text-4xl font-bold text-foreground">
          {currency}{price}
        </span>
        {oldPrice && (
          <span className="text-xl text-muted-foreground line-through mb-1">
            {currency}{oldPrice}
          </span>
        )}
      </div>

      <p className="text-muted-foreground text-base leading-relaxed mb-6">
        We've carefully designed the {productName} to be something you'll love using every single day. It's built to last, feels great to use, and fits right into your lifestyle without any fuss.
      </p>

      {/* Colors (Radio Group) */}
      {colors.length > 0 && (
        <fieldset className="flex flex-col gap-3 mb-8">
          <legend className="text-sm font-medium text-foreground">
            Color: <span className="text-muted-foreground ml-1">{colors[selectedColor]?.name || ""}</span>
          </legend>
          <div className="flex flex-wrap gap-2.5" role="radiogroup" aria-label="Select color">
            {colors.map((color, idx) => {
              const isSelected = selectedColor === idx;
              return (
                <button
                  key={idx}
                  type="button"
                  role="radio"
                  aria-checked={isSelected}
                  onClick={() => onSelectColor(idx)}
                  className={`px-4 py-2 rounded-xl text-sm font-medium border transition-all cursor-pointer ${
                    isSelected
                      ? "border-foreground bg-foreground text-background shadow-xs font-semibold"
                      : "border-border bg-card text-foreground hover:bg-secondary/70 hover:border-border"
                  }`}
                >
                  <span>{color.name}</span>
                </button>
              );
            })}
          </div>
        </fieldset>
      )}

      {/* Stock status */}
      <div className="flex items-center gap-2 mb-6">
        {isVariantInStock ? (
          <>
            <div className="flex items-center justify-center size-5 rounded-full bg-emerald-500/10 text-emerald-500">
              <Check className="size-3.5" />
            </div>
            <span className="text-sm font-medium text-emerald-600 dark:text-emerald-400">
              In Stock ({availableStock} available)
            </span>
          </>
        ) : (
          <>
            <div className="flex items-center justify-center size-5 rounded-full bg-destructive/10 text-destructive">
              <Clock className="size-3.5" />
            </div>
            <span className="text-sm font-medium text-destructive">
              Out of Stock
            </span>
          </>
        )}
      </div>

      {/* Purchase actions */}
      <div className="flex flex-col gap-3.5 mb-8">
        <div className="w-full h-12">
          <QuantitySelector
            value={Math.min(quantity, Math.max(1, availableStock))}
            max={Math.max(1, availableStock)}
            disabled={!isVariantInStock}
            onChange={onQuantityChange}
          />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <CustomButton
            text={isAdding ? "Adding..." : isVariantInStock ? "Add to Cart" : "Out of Stock"}
            icon={<ShoppingCart className="size-4" />}
            onClick={onAddToCart}
            className={`py-3.5 text-sm shadow-sm ${!isVariantInStock ? "opacity-50 cursor-not-allowed" : ""}`}
            disabled={isAdding || !isVariantInStock}
          />

          <button
            type="button"
            onClick={onBuyNow}
            disabled={isAdding || !isVariantInStock}
            className="py-3.5 px-4 rounded-xl bg-accent-brand hover:bg-accent-brand/90 text-white font-semibold text-sm transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          >
            <Zap className="size-4 fill-current" />
            <span>Buy Now</span>
          </button>
        </div>

        <div className="flex gap-3">
          <button
            type="button"
            onClick={onToggleWishlist}
            className={`flex-1 py-2.5 px-4 flex items-center justify-center gap-2 rounded-xl border transition-all text-xs font-medium shadow-sm cursor-pointer ${
              isWishlisted
                ? "border-red-500/40 bg-red-500/10 text-red-600"
                : "border-border bg-card hover:bg-secondary hover:text-foreground text-muted-foreground"
            }`}
          >
            <Heart className={`size-4 ${isWishlisted ? "fill-current text-red-600" : ""}`} />
            <span>{isWishlisted ? "Saved to Wishlist" : "Add to Wishlist"}</span>
          </button>

          <button
            type="button"
            onClick={onShare}
            className="py-2.5 px-4 flex items-center justify-center gap-2 rounded-xl border border-border bg-card hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors shadow-sm text-xs font-medium cursor-pointer"
          >
            <Share2 className="size-4" />
            <span>Share</span>
          </button>
        </div>
      </div>
    </div>
  );
}
