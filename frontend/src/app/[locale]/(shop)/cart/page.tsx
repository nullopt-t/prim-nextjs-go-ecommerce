import CartLayout from "@/features/cart/components/layout/cartLayout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.shop.cart");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function CartPage() {
  return <CartLayout />;
}
