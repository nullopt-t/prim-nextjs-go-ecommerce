import ProductsLayout from "@/features/products/components/layout/productsLayout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta.shop.products");
  return {
    title: t("title"),
    description: t("description"),
  };
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ category?: string }>;
}) {
  const { category } = await searchParams;

  return <ProductsLayout category={category} />;
}
