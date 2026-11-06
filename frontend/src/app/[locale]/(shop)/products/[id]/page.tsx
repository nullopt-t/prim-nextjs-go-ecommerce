import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { ProductDetails } from "@/modules/catalog/components/ProductDetails";
import { getProductBySlug } from "@/shared/lib/server/catalog";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";
import { routing } from "@/shared/i18n/routing";

type ProductPageProps = {
	params: Promise<{ locale: string; id: string }>;
};

function localizedTitle(product: any, locale: string, fallback: string) {
	const name = product?.product?.[locale] || product?.product?.en || product?.title;
	return typeof name === "string" && name.trim() ? name : fallback;
}

export async function generateMetadata({
	params,
}: ProductPageProps): Promise<Metadata> {
	const { locale, id } = await params;
	const t = await getTranslations({ locale, namespace: "meta.shop.products" });
	const product = await getProductBySlug(id);

	if (!product) {
		return { title: t("title"), description: t("description") };
	}

	const title = localizedTitle(product, locale, id);
	const description =
		typeof product.description === "string" && product.description.trim()
			? product.description.slice(0, 160)
			: t("description");
	const path = `/${locale}/products/${product.slug || id}`;

	return {
		title,
		description,
		alternates: {
			canonical: path,
			languages: Object.fromEntries(
				routing.locales.map((supported) => [
					supported,
					`/${supported}/products/${product.slug || id}`,
				])
			),
		},
		openGraph: {
			title,
			description,
			type: "website",
			url: path,
			images: product.img ? [{ url: product.img, alt: title }] : undefined,
		},
		twitter: {
			card: "summary_large_image",
			title,
			description,
			images: product.img ? [product.img] : undefined,
		},
	};
}

export default async function ProductDetailPage({ params }: ProductPageProps) {
	const { locale, id } = await params;
	const [product, settings] = await Promise.all([
		getProductBySlug(id),
		getPublicSettings(),
	]);

	if (!product) notFound();

	const price = Number(product.price);
	const brandName =
		typeof product.brand === "string"
			? product.brand
			: typeof product.brand?.name === "string"
				? product.brand.name
				: undefined;

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "Product",
		name: localizedTitle(product, locale, id),
		description:
			typeof product.description === "string" ? product.description : undefined,
		image: product.img ? [product.img] : undefined,
		sku: product.sku || undefined,
		brand: brandName ? { "@type": "Brand", name: brandName } : undefined,
		aggregateRating:
			Number(product.stars) > 0 && Number(product.reviews) > 0
				? {
						"@type": "AggregateRating",
						ratingValue: Number(product.stars),
						reviewCount: Number(product.reviews),
					}
				: undefined,
		offers:
			price > 0
				? {
						"@type": "Offer",
						price,
						priceCurrency: settings?.currency || product.currency || undefined,
						itemCondition: "https://schema.org/NewCondition",
						availability: product.inStock
							? "https://schema.org/InStock"
							: "https://schema.org/OutOfStock",
						seller: settings?.storeName
							? { "@type": "Organization", name: settings.storeName }
							: undefined,
					}
				: undefined,
	};

	const breadcrumbJsonLd = {
		"@context": "https://schema.org",
		"@type": "BreadcrumbList",
		itemListElement: [
			{ "@type": "ListItem", position: 1, name: "Home", item: `/${locale}` },
			{ "@type": "ListItem", position: 2, name: "Products", item: `/${locale}/products` },
			{ "@type": "ListItem", position: 3, name: localizedTitle(product, locale, id) },
		],
	};

	return (
		<>
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
			<script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }} />
			<ProductDetails initialProduct={product} />
		</>
	);
}
