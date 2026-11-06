import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import ProductsLayout from "@/modules/catalog/components/layout/productsLayout";
import { getBrands, getProducts } from "@/shared/lib/server/catalog";
import { localizedName } from "@/shared/lib/localizedName";

type BrandDetailProps = {
	params: Promise<{ locale: string; slug: string }>;
};

function brandSlug(brand: { id: string; name: string }): string {
	return brand.name.toLowerCase().trim().replace(/\s+/g, "-") || brand.id;
}

async function findBrand(slug: string) {
	const brands = await getBrands(100);
	return (
		brands.find(
			(brand) => brand.id === slug || brandSlug(brand) === slug.toLowerCase(),
		) ?? null
	);
}

export async function generateMetadata({ params }: BrandDetailProps): Promise<Metadata> {
	const { locale, slug } = await params;
	const t = await getTranslations({ locale, namespace: "meta.shop.brands" });
	const brand = await findBrand(slug);
	if (!brand) return { title: t("title"), description: t("description") };
	return {
		title: `${localizedName(brand.name, locale)} | ${t("title")}`,
		description: t("description"),
	};
}

export default async function BrandDetailPage({ params }: BrandDetailProps) {
	const { locale, slug } = await params;
	const brand = await findBrand(slug);
	if (!brand) notFound();

	const all = await getProducts({ pageSize: 500 });
	const brandName = brand.name.toLowerCase();
	const products = all.filter((product) => {
		const productBrand =
			typeof product.brand === "string" ? product.brand : product.brand?.name;
		return (
			product.brand === brand.id ||
			(typeof productBrand === "string" && productBrand.toLowerCase() === brandName)
		);
	});

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "Brand",
		name: localizedName(brand.name, locale),
	};

	return (
		<>
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<ProductsLayout heading={localizedName(brand.name, locale)} products={products} />
		</>
	);
}
