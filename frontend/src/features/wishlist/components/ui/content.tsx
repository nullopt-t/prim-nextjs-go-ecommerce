"use client";

import { useState } from "react";
import { ProductsGrid } from "@/components/ui/productsGrid";
import { useTranslations, useLocale } from "next-intl";
import { useWishlist } from "@/hooks/useWishlist";
import { useAuthContext } from "@/context/AuthContext";
import { useCartContext } from "@/context/CartContext";
import { toast } from "sonner";
import { normalizeProduct } from "@/context/CatalogContext";
import { Link } from "@/i18n/navigation";

export default function WishlistContent() {
  const t = useTranslations("wishlist");
  const locale = useLocale();
  const { isAuthenticated } = useAuthContext();
  const { wishlistItems, loading, clearWishlist, removeFromWishlist, removeWishlistItem } = useWishlist();
  const { addToCart } = useCartContext();

  const [isAddingAll, setIsAddingAll] = useState(false);
  const [isClearing, setIsClearing] = useState(false);

  const items = Array.isArray(wishlistItems)
    ? wishlistItems.map(normalizeProduct).filter(Boolean)
    : [];

  const handleAddAllToCart = async () => {
    if (!items || items.length === 0 || isAddingAll) return;
    setIsAddingAll(true);
    let addedCount = 0;
    try {
      for (const item of items) {
        if (!item) continue;
        const res = await addToCart({
          id: String(item.productId || item.id),
          slug: item.slug,
          productName: item.product,
          productPrice: typeof item.price === "number" ? `$${item.price.toFixed(2)}` : String(item.price),
          img: typeof item.img === "string" ? item.img : "/placeholder-product.png",
          quantity: 1,
        });
        if (res.success) addedCount++;
      }
      if (addedCount > 0) {
        toast.success(t("toast.addedAll", { count: addedCount }));
      }
    } catch {
      toast.error(locale === "ar" ? "فشلت إضافة بعض العناصر إلى السلة" : "Failed to add some items");
    } finally {
      setIsAddingAll(false);
    }
  };

  const handleClearAll = async () => {
    if (items.length === 0 || isClearing) return;
    if (typeof window !== "undefined" && !window.confirm(t("clearConfirm"))) {
      return;
    }
    setIsClearing(true);
    try {
      await clearWishlist();
      toast.success(t("toast.cleared"));
    } finally {
      setIsClearing(false);
    }
  };

  const handleRemove = async (item: any) => {
    const productName = typeof item.product === "object" ? (item.product.ar || item.product.en) : item.title || "Product";
    try {
      if (item.wishlistItemId) {
        await removeWishlistItem(item.wishlistItemId, productName);
      } else {
        await removeFromWishlist(String(item.productId || item.id), item.variantId, productName);
      }
      toast.success(t("toast.itemRemoved", { name: productName }));
    } catch {
      // errors handled inside hook
    }
  };

  return (
    <div className="w-full max-w-7xl mx-auto flex flex-col gap-6 pb-12">
      {/* ── Header Toolbar ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/70 pb-5">
        <div>
          <h1 className="text-2xl font-bold text-foreground tracking-tight">
            {t("title")}
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
            {t("savedItems", { count: items.length })}
          </p>
        </div>

        <div className="flex items-center gap-2.5 shrink-0">
          <Link
            href="/products"
            className="px-3.5 py-1.5 rounded-lg text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-colors shadow-2xs"
          >
            {t("continueShopping")}
          </Link>

          {items.length > 0 && (
            <>
              <button
                type="button"
                onClick={handleClearAll}
                disabled={isClearing}
                className="px-3.5 py-1.5 rounded-lg text-xs font-medium text-destructive hover:bg-destructive/10 border border-destructive/20 transition-colors disabled:opacity-50 shadow-2xs"
              >
                {t("clearAll")}
              </button>

              <button
                type="button"
                onClick={handleAddAllToCart}
                disabled={isAddingAll}
                className="px-4 py-1.5 rounded-lg text-xs font-semibold bg-accent-brand text-white hover:opacity-90 transition-opacity disabled:opacity-60 shadow-2xs"
              >
                {isAddingAll ? t("addingAll") : t("addAllToCart")}
              </button>
            </>
          )}
        </div>
      </div>

      {/* ── Content States ── */}
      {loading && items.length === 0 ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3 sm:gap-4 md:gap-5">
          {Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              className="bg-card border border-border rounded-2xl p-4 flex flex-col gap-3 animate-pulse"
            >
              <div className="aspect-square bg-secondary/50 rounded-xl" />
              <div className="h-4 bg-secondary/60 rounded w-3/4" />
              <div className="h-4 bg-secondary/40 rounded w-1/3" />
            </div>
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="w-full bg-card border border-border rounded-2xl p-10 sm:p-16 text-center flex flex-col items-center justify-center gap-3 shadow-2xs">
          <div className="size-16 rounded-full bg-secondary/60 border border-border/80 flex items-center justify-center text-muted-foreground text-xl mb-1">
            <span className="opacity-70">♥</span>
          </div>
          <h2 className="text-lg font-bold text-foreground">
            {t("empty.title")}
          </h2>
          <p className="text-xs sm:text-sm text-muted-foreground max-w-md leading-relaxed">
            {!isAuthenticated
              ? t("empty.guestSubtitle")
              : t("empty.userSubtitle")}
          </p>

          <div className="flex items-center gap-3 mt-4">
            {!isAuthenticated ? (
              <Link
                href="/auth"
                className="px-4 py-2 rounded-lg bg-accent-brand text-white text-xs font-semibold hover:opacity-90 transition-opacity shadow-xs"
              >
                {t("empty.signIn")}
              </Link>
            ) : null}
            <Link
              href="/products"
              className="px-4 py-2 rounded-lg bg-secondary hover:bg-secondary/80 text-foreground border border-border text-xs font-semibold transition-colors shadow-xs"
            >
              {t("empty.browseCatalog")}
            </Link>
          </div>
        </div>
      ) : (
        <ProductsGrid
          products={items}
          className="grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
          isWishlist={true}
          onRemoveItem={handleRemove}
        />
      )}
    </div>
  );
}
