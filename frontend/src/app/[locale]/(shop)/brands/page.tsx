import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";
import { getBrands, getCategories, getProducts } from "@/shared/lib/server/catalog";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { localizedName } from "@/shared/lib/localizedName";
import {
	BrandsDirectory,
	type BrandDirectoryItem,
} from "@/modules/catalog/components/brands/brandsDirectory";
import type { Brand, Category, ProductDetailsEntity } from "@/modules/catalog/types";

type BrandsPageProps = {
	params: Promise<{ locale: string }>;
};

export async function generateMetadata(): Promise<Metadata> {
	const t = await getTranslations("meta.shop.brands");
	return { title: t("title"), description: t("description") };
}

function brandSlug(brand: { id: string; name: string }): string {
	return brand.name.toLowerCase().trim().replace(/\s+/g, "-") || brand.id;
}

/** Category label for a product, preferring the localized catalog name. */
function categoryLabels(categories: Category[], locale: string): Map<string, string> {
	const labels = new Map<string, string>();
	for (const category of categories) {
		const label = localizedName(category.name, locale, category.name);
		labels.set(category.name.toLowerCase(), label);
		if (category.slug) labels.set(category.slug.toLowerCase(), label);
		if (category.id) labels.set(String(category.id).toLowerCase(), label);
	}
	return labels;
}

function brandCategories(
	products: ProductDetailsEntity[],
	labels: Map<string, string>,
): Map<string, Set<string>> {
	const byBrand = new Map<string, Set<string>>();
	for (const product of products) {
		const brand =
			typeof product.brand === "string" ? product.brand : product.brand?.name;
		if (!brand || !product.category) continue;
		const label =
			labels.get(String(product.categoryId ?? "").toLowerCase()) ??
			labels.get(product.category.toLowerCase()) ??
			product.category;
		const key = brand.toLowerCase();
		const set = byBrand.get(key) ?? new Set<string>();
		set.add(label);
		byBrand.set(key, set);
	}
	return byBrand;
}

function toDirectoryItem(
	brand: Brand,
	locale: string,
	byBrand: Map<string, Set<string>>,
): BrandDirectoryItem {
	const name = localizedName(brand.name, locale, brand.name);
	const categories = [...(byBrand.get(brand.name.toLowerCase()) ?? [])].sort((a, b) =>
		a.localeCompare(b),
	);
	return {
		id: String(brand.id),
		name,
		slug: brandSlug({ id: String(brand.id), name: brand.name }),
		logoUrl: brand.logoUrl || brand.logo
			? normalizeMediaUrl(brand.logoUrl || brand.logo)
			: undefined,
		productCount: brand.productCount,
		categories,
	};
}

export default async function BrandsPage({ params }: BrandsPageProps) {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "shop" });
	// Category chips derive from the first 500 products; enough for a directory
	// view without paging the whole catalog.
	const [brands, categories, products] = await Promise.all([
		getBrands(100),
		getCategories(),
		getProducts({ pageSize: 500 }),
	]);

	const labels = categoryLabels(categories, locale);
	const byBrand = brandCategories(products, labels);
	const items = brands.map((brand) => toDirectoryItem(brand, locale, byBrand));

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "ItemList",
		itemListElement: items.map((brand, index) => ({
			"@type": "ListItem",
			position: index + 1,
			url: `/brands/${brand.slug}`,
			name: brand.name,
		})),
	};

	return (
		<div className="w-full py-4 sm:py-6">
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<h1 className="text-3xl font-bold text-foreground md:text-4xl">{t("brands.title")}</h1>
			<p className="mt-2 text-sm text-muted-foreground">{t("brands.description")}</p>

			{brands.length === 0 ? (
				<div className="mt-8 rounded-lg border border-border bg-background px-4 py-16 text-center">
					<p className="text-sm font-medium text-foreground">{t("brands.emptyTitle")}</p>
					<p className="mt-1 text-sm text-muted-foreground">{t("brands.emptyDescription")}</p>
				</div>
			) : (
				<BrandsDirectory brands={items} />
			)}
		</div>
	);
}
