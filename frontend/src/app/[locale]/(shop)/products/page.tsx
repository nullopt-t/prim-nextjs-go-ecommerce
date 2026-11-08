import ProductsLayout from "@/modules/catalog/components/layout/productsLayout";
import { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { getCategories, getProductsWithMeta } from "@/shared/lib/server/catalog";

type ProductsSearchParams = {
  category?: string;
  q?: string;
  brand?: string;
  inStock?: string;
  minPrice?: string;
  maxPrice?: string;
  sort?: string;
  page?: string;
  /** Comma-separated "key:value" facet pairs, e.g. color:black,size:XL. */
  attribute?: string;
};

function parseNumber(value?: string): number | undefined {
  if (!value) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : undefined;
}

function parsePage(value?: string): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 1 ? Math.floor(parsed) : 1;
}

function parseAttributes(value?: string): Record<string, string> | undefined {
  if (!value) return undefined;
  const out: Record<string, string> = {};
  for (const part of value.split(",")) {
    const colon = part.indexOf(":");
    if (colon <= 0 || colon >= part.length - 1) continue;
    const k = part.slice(0, colon).trim();
    const v = part.slice(colon + 1).trim();
    if (k && v) out[k] = v;
  }
  return Object.keys(out).length > 0 ? out : undefined;
}

function resolveCategoryName(
	name: string | { en?: string; ar?: string } | undefined,
	locale: string,
	fallback: string,
): string {
	if (typeof name === "string") return name.trim() ? name : fallback;
	if (!name) return fallback;
	return (locale === "ar" ? name.ar || name.en : name.en || name.ar) || fallback;
}

export async function generateMetadata({
  searchParams,
}: {
  searchParams: Promise<ProductsSearchParams>;
}): Promise<Metadata> {
  const t = await getTranslations("meta.shop.products");
  const { category, q } = await searchParams;
  if (q) {
    // Search pages are thin, query-specific surfaces: keep them out of the
    // index so crawlers follow the curated category/product pages instead.
    return { title: `${q} | ${t("title")}`, description: t("description"), robots: { index: false } };
  }
  if (category && category.toLowerCase() !== "all") {
    const categories = await getCategories();
    const matched = categories.find(
      (c) => String(c.id) === category || (c as { slug?: string }).slug === category,
    );
    return {
      title: `${resolveCategoryName(matched?.name, "en", category.replace(/-/g, " "))} | ${t("title")}`,
      description: t("description"),
    };
  }
  return { title: t("title"), description: t("description") };
}

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<ProductsSearchParams>;
}) {
  const params = await searchParams;
  const page = parsePage(params.page);
  const { products, totalItems, totalPages } = await getProductsWithMeta({
    category: params.category && params.category.toLowerCase() !== "all" ? params.category : undefined,
    search: params.q?.trim() ? params.q.trim() : undefined,
    brand: params.brand?.trim() ? params.brand.trim() : undefined,
    inStock: params.inStock === "true",
    minPrice: parseNumber(params.minPrice),
    maxPrice: parseNumber(params.maxPrice),
    sort: params.sort?.trim() ? params.sort.trim() : undefined,
    attributes: parseAttributes(params.attribute),
    page,
    pageSize: 24,
  });

  const itemListJsonLd = {
    "@context": "https://schema.org",
    "@type": "ItemList",
    itemListElement: products.slice(0, 20).map((product, index) => ({
      "@type": "ListItem",
      position: index + 1,
      url: `/products/${product.slug || product.id}`,
      name:
        (typeof product.product === "object"
          ? product.product?.en || product.product?.ar
          : product.product) ||
        product.title ||
        "",
    })),
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(itemListJsonLd) }} />
      <ProductsLayout
        category={params.category}
        products={products}
        page={page}
        totalItems={totalItems}
        totalPages={totalPages}
      />
    </>
  );
}
