import { Checkout } from "@/features/checkout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.shop.checkout");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default function CheckoutPage() {
  return <Checkout />;
}
