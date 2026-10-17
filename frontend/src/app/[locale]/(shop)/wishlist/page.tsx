import { WishlistContent } from "@/features/wishlist";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.shop.wishlist");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function WishlistPage() {
  return <WishlistContent />;
}
