"use client";

import { Title } from "@/components/ui/title";
import { ProductsGrid } from "@/components/ui/productsGrid";
import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";
import { useWishlist } from "@/hooks/useWishlist";
import { useCartContext } from "@/context/CartContext";
import { toast } from "sonner";
import { normalizeProduct } from "@/context/CatalogContext";

const fallbackWishlistItems = [
  {
    id: "wishlist-1",
    img: "/placeholder-product.png",
    product: {
      en: "Sony WH-1000XM5",
      ar: "سوني WH-1000XM5",
    },
    stars: 5,
    reviews: 256,
    price: 399,
    oldPrice: 450,
    discountPercentage: "11%",
  },
  {
    id: "wishlist-2",
    img: "/placeholder-product.png",
    product: {
      en: "Apple Watch Ultra 2",
      ar: "آبل واتش الترا 2",
    },
    stars: 4.8,
    reviews: 180,
    price: 799,
    oldPrice: 899,
    discountPercentage: "15%",
  },
];

export default function WishlistContent() {
  const t = useTranslations("wishlist");
  const { wishlistItems, loading } = useWishlist();
  const { addToCart, openCart } = useCartContext();

  const rawList = Array.isArray(wishlistItems)
    ? wishlistItems
    : Array.isArray(wishlistItems?.data)
    ? wishlistItems.data
    : null;

  const items = (rawList && rawList.length > 0)
    ? rawList.map(normalizeProduct)
    : fallbackWishlistItems;

  const handleAddAllToCart = async () => {
    if (!items || items.length === 0) return;
    for (const item of items) {
      await addToCart({
        id: String(item.id),
        productName: item.product,
        productPrice: `$${item.price}`,
        img: item.img || "/placeholder-product.png",
        quantity: 1,
      });
    }
    toast.success("Added all wishlist items to cart!");
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <Title
          title={t("title")}
          subtitle={t("savedItems", { count: items.length })}
          className="mb-0"
        />
        {items.length > 0 && (
          <div className="w-full sm:w-44 h-10 bg-primary text-primary-foreground hover:opacity-90 font-medium rounded-md">
            <CustomButton text={t("addToCart")} onClick={handleAddAllToCart} />
          </div>
        )}
      </div>

      {loading && !rawList ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="aspect-square bg-secondary/50 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : (
        <ProductsGrid
          products={items}
          className="grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
          isWishlist={true}
        />
      )}
    </div>
  );
}
