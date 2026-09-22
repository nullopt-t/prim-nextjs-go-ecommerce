"use client";

import { Title } from "@/components/ui/title";
import { ProductsGrid } from "@/components/ui/productsGrid";
import { CustomButton } from "@/components/ui/button";
import { useTranslations } from "next-intl";

const defaultWishlistItems = [
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

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <Title
          title={t("title")}
          subtitle={t("savedItems", { count: defaultWishlistItems.length })}
          className="mb-0"
        />
        <div className="w-full sm:w-44 h-10 bg-primary text-primary-foreground hover:opacity-90 font-medium rounded-md">
          <CustomButton text={t("addToCart")} onClick={() => {}} />
        </div>
      </div>
      <ProductsGrid
        products={defaultWishlistItems}
        className="grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        isWishlist={true}
      />
    </div>
  );
}
