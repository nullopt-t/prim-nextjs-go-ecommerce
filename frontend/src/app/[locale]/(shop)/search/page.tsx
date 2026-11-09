import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import ProductsLayout from "@/modules/catalog/components/layout/productsLayout";
import { getProductsWithMeta } from "@/shared/lib/server/catalog";
import {
	parsePage,
	toCatalogQuery,
	type ProductListSearchParams,
} from "@/shared/lib/server/productListParams";

type SearchPageProps = {
	searchParams: Promise<ProductListSearchParams>;
};

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.products");
	// Search is a thin, query-specific surface: keep it out of the index and
	// let the canonical category/product pages carry SEO weight.
	return { title: t("title"), description: t("description"), robots: { index: false } };
}

export default async function SearchPage({ searchParams }: SearchPageProps) {
	const params = await searchParams;
	const page = parsePage(params.page);

	// Same facet params as /products so shared links filter identically here.
	const { products, totalItems, totalPages } = await getProductsWithMeta(toCatalogQuery(params, 24));

	return (
		<ProductsLayout
			category={params.category}
			products={products}
			page={page}
			totalItems={totalItems}
			totalPages={totalPages}
		/>
	);
}
