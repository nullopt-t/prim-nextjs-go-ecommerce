import ProductsLayout from "@/features/products/components/layout/productsLayout";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Products Catalog | PRIM",
  description: "Browse our premium audio and gear catalog",
};

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ category?: string }>;
}) {
  const { category } = await searchParams;

  return <ProductsLayout category={category} />;
}
