"use client";

import { Title } from "@/components/ui/title";
import { ProductsGrid } from "@/components/ui/productsGrid";
import { CustomButton } from "@/components/ui/button";
import { useTranslations, useLocale } from "next-intl";
import { useWishlist } from "@/hooks/useWishlist";
import { useAuthContext } from "@/context/AuthContext";
import { useCartContext } from "@/context/CartContext";
import { toast } from "sonner";
import { normalizeProduct } from "@/context/CatalogContext";
import { Heart, Trash2, ArrowRight, ShoppingBag } from "lucide-react";
import { Link } from "@/i18n/navigation";

export default function WishlistContent() {
  const t = useTranslations("wishlist");
  const locale = useLocale();
  const { isAuthenticated } = useAuthContext();
  const { wishlistItems, loading, clearWishlist, removeFromWishlist, removeWishlistItem } = useWishlist();
  const { addToCart } = useCartContext();

  const items = Array.isArray(wishlistItems)
    ? wishlistItems.map(normalizeProduct).filter(Boolean)
    : [];

  const handleAddAllToCart = async () => {
    if (!items || items.length === 0) return;
    let addedCount = 0;
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
      toast.success(
        locale === "ar"
          ? `تمت إضافة ${addedCount} منتجات إلى سلة التسوق!`
          : `Added ${addedCount} items to your cart!`
      );
    }
  };

  const handleRemove = async (item: any) => {
    const productName = typeof item.product === "object" ? (item.product.en || item.product.ar) : item.title;
    if (item.wishlistItemId) {
      await removeWishlistItem(item.wishlistItemId, productName);
    } else {
      await removeFromWishlist(String(item.productId || item.id), productName);
    }
  };

  return (
    <div className="w-full">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <Title
          title={t("title")}
          subtitle={t("savedItems", { count: items.length })}
          className="mb-0"
        />
        {items.length > 0 && (
          <div className="flex items-center gap-3">
            <button
              onClick={() => clearWishlist()}
              className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
            >
              <Trash2 className="size-4" />
              <span>{locale === "ar" ? "مسح الكل" : "Clear All"}</span>
            </button>
            <div className="w-full sm:w-44 h-10 bg-primary text-primary-foreground hover:opacity-90 font-medium rounded-md">
              <CustomButton text={t("addToCart")} onClick={handleAddAllToCart} />
            </div>
          </div>
        )}
      </div>

      {loading && items.length === 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="aspect-square bg-secondary/50 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 px-4 text-center border border-dashed border-border rounded-2xl bg-card/50">
          <div className="size-16 rounded-full bg-accent-brand/10 flex items-center justify-center text-accent-brand mb-4">
            <Heart className="size-8 stroke-[1.5]" />
          </div>
          <h3 className="text-lg font-semibold text-foreground mb-1">
            {locale === "ar" ? "قائمة رغباتك فارغة" : "Your wishlist is empty"}
          </h3>
          <p className="text-sm text-muted-foreground max-w-sm mb-6">
            {!isAuthenticated
              ? locale === "ar"
                ? "قم بتسجيل الدخول لحفظ منتجاتك المفضلة ومزامنتها عبر جميع أجهزتك."
                : "Sign in to save your favorite products and sync them across all your devices."
              : locale === "ar"
              ? "استكشف مجموعتنا واحفظ المنتجات التي تعجبك بالضغط على أيقونة القلب."
              : "Explore our catalog and save the items you love by clicking the heart icon."}
          </p>
          <Link
            href={!isAuthenticated ? "/sign-in" : "/products"}
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-accent-brand text-white text-sm font-medium hover:opacity-90 transition-opacity"
          >
            <span>{!isAuthenticated ? (locale === "ar" ? "تسجيل الدخول" : "Sign In") : (locale === "ar" ? "تصفح المنتجات" : "Explore Products")}</span>
            <ArrowRight className="size-4 rtl:rotate-180" />
          </Link>
        </div>
      ) : (
        <ProductsGrid
          products={items}
          className="grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
          isWishlist={true}
          onRemoveItem={handleRemove}
        />
      )}
    </div>
  );
}
