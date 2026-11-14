import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { Categories } from "@/modules/catalog/components/categories";
import { getCategories, getProducts } from "@/shared/lib/server/catalog";
export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.categories");
	return { title: t("title"), description: t("description") };
}

function categoryName(name: string | { en?: string; ar?: string } | undefined, locale: string): string {
	if (typeof name === "string") return name;
	if (!name) return "";
	return (locale === "ar" ? name.ar || name.en : name.en || name.ar) || "";
}

export default async function CategoriesPage({
	params,
}: {
	params: Promise<{ locale: string }>;
}) {
	const { locale } = await params;
	const [categories, products] = await Promise.all([
		getCategories(),
		getProducts({ pageSize: 500 }),
	]);

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "ItemList",
		itemListElement: categories.map((category, index) => ({
			"@type": "ListItem",
			position: index + 1,
			url: `/${locale}/products?category=${encodeURIComponent(category.slug || String(category.id))}`,
			name: categoryName(category.name, locale),
		})),
	};

	return (
		<>
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<Categories categories={categories} products={products} />
		</>
	);
}
