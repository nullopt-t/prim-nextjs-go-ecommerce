import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { ProductDetails } from "@/modules/catalog/components/ProductDetails";
import { getProductBySlug, getReviewsByProduct } from "@/shared/lib/server/catalog";
import { getPublicSettings } from "@/shared/lib/server/publicSettings";
import { routing } from "@/shared/i18n/routing";
import type { ProductReview, Review } from "@/modules/catalog/types";
import { mapReview } from "@/modules/catalog/utils/mapReview";

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

	const [rawReviews, tProducts] = await Promise.all([
		getReviewsByProduct(String(product.id)),
		getTranslations({ locale, namespace: "products" }),
	]);

	const initialReviews: ProductReview[] = rawReviews.map((r: Review) =>
		mapReview(r, {
			locale,
			authorLabel: tProducts("details.customer"),
			titleLabel: tProducts("details.customerReview"),
			dateLabel: tProducts("details.recently"),
		}),
	);

	const price = Number(product.price);
	const brandName =
		typeof product.brand === "string"
			? product.brand
			: typeof product.brand?.name === "string"
				? product.brand.name
				: undefined;
	// The product's SKU comes from its default variant; a configurable product
	// has no product-level code of its own.
	const sku =
		product.sku ??
		product.variants?.find((variant) => variant.isDefault)?.sku ??
		product.variants?.[0]?.sku;
	// Availability follows the variants: a product whose default variant sold
	// out but whose other variants remain is still purchasable.
	const variants = product.variants ?? [];
	const anyVariantInStock = variants.some(
		(variant) => variant.stock?.isInStock === true || (variant.stock?.availableQuantity ?? 0) > 0,
	);
	const inStock = variants.length > 0 ? anyVariantInStock : (product.inStock ?? true);

	// Per-review markup enables review rich snippets; the first five match the
	// order shown on the page. datePublished is the raw ISO timestamp.
	const reviewJsonLd = rawReviews.slice(0, 5).map((r: Review) => ({
		"@type": "Review",
		name: r.title || undefined,
		reviewBody: r.body || undefined,
		datePublished: r.createdAt || undefined,
		author: {
			"@type": "Person",
			name: r.user?.name || tProducts("details.customer"),
		},
		reviewRating: {
			"@type": "Rating",
			ratingValue: Number(r.rating) || 0,
			bestRating: 5,
			worstRating: 1,
		},
	}));

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "Product",
		name: localizedTitle(product, locale, id),
		description:
			typeof product.description === "string" ? product.description : undefined,
		image: product.img ? [product.img] : undefined,
		sku: sku || undefined,
		brand: brandName ? { "@type": "Brand", name: brandName } : undefined,
		aggregateRating:
			Number(product.stars) > 0 && Number(product.reviews) > 0
				? {
						"@type": "AggregateRating",
						ratingValue: Number(product.stars),
						reviewCount: Number(product.reviews),
					}
				: undefined,
		review: reviewJsonLd.length > 0 ? reviewJsonLd : undefined,
		offers:
			price > 0
				? {
						"@type": "Offer",
						price,
						priceCurrency: settings?.currency || product.currency || undefined,
						itemCondition: "https://schema.org/NewCondition",
						availability: inStock
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
			<ProductDetails initialProduct={product} initialReviews={initialReviews} />
		</>
	);
}
