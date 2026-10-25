import ProductsLayout from "@/modules/catalog/components/layout/productsLayout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { getCatalogSnapshot } from "@/shared/lib/server/catalog";

export async function generateMetadata({
  searchParams,
}: {
  searchParams: Promise<{ category?: string; q?: string }>;
}): Promise<Metadata> {
  const t = await getTranslations("meta.shop.products");
  const { category, q } = await searchParams;
  if (q) {
    return { title: `${q} | ${t("title")}`, description: t("description") };
  }
  if (category) {
    return {
      title: `${category.replace(/-/g, " ")} | ${t("title")}`,
      description: t("description"),
    };
  }
  return { title: t("title"), description: t("description") };
}

function matchesCategory(product: any, category?: string) {
  if (!category || category.toLowerCase() === "all") return true;
  const needle = category.toLowerCase();
  const name = (product.category || "").toLowerCase().replace(/\s+/g, "-");
  return (
    String(product.categoryId || "").toLowerCase() === needle ||
    name === needle ||
    name.replace(/-/g, " ") === needle.replace(/-/g, " ")
  );
}

function matchesSearch(product: any, query?: string) {
  if (!query) return true;
  const needle = query.toLowerCase();
  const title = (product.title || product.product?.en || product.product?.ar || "").toLowerCase();
  const brand = (typeof product.brand === "string" ? product.brand : product.brand?.name || "").toLowerCase();
  const description = (product.description || "").toLowerCase();
  return title.includes(needle) || brand.includes(needle) || description.includes(needle);
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ category?: string; q?: string }>;
}) {
  const { category, q } = await searchParams;
  const { products } = await getCatalogSnapshot();
  const filtered = products.filter(
    (product) => matchesCategory(product, category) && matchesSearch(product, q),
  );

  return <ProductsLayout category={category} products={filtered} />;
}
