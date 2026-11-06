import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import ProductsLayout from "@/modules/catalog/components/layout/productsLayout";
import { getProductsWithMeta } from "@/shared/lib/server/catalog";

type SearchPageProps = {
	searchParams: Promise<{ q?: string }>;
};

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.products");
	// Search is a thin, query-specific surface: keep it out of the index and
	// let the canonical category/product pages carry SEO weight.
	return { title: t("title"), description: t("description"), robots: { index: false } };
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
	const { q } = await searchParams;
	const query = q?.trim() ? q.trim() : undefined;

	const { products } = await getProductsWithMeta({
		search: query,
		pageSize: 100,
	});

	return <ProductsLayout products={products} />;
}
