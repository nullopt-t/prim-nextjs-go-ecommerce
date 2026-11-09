import type { MetadataRoute } from "next";
import { getProducts } from "@/shared/lib/server/catalog";
import type { ProductDetailsEntity } from "@/modules/catalog/types";
import { routing } from "@/shared/i18n/routing";

function siteUrl(): string {
	return (process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000").replace(/\/$/, "");
}

const STATIC_ROUTES = [
	"",
	"/products",
	"/categories",
	"/brands",
	"/about",
	"/faqs",
	"/contact",
	"/shipping",
	"/returns",
	"/careers",
	"/press",
];

/** Fetch every product so the sitemap stays complete past a single page. */
async function getAllProducts(): Promise<ProductDetailsEntity[]> {
	const products: ProductDetailsEntity[] = [];
	let page = 1;
	const pageSize = 200;
	for (;;) {
		const batch = await getProducts({ page, pageSize }).catch(() => []);
		products.push(...batch);
		if (batch.length < pageSize) break;
		page += 1;
	}
	return products;
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
	const base = siteUrl();
	const products = await getAllProducts();
	const now = new Date();

	const staticEntries: MetadataRoute.Sitemap = routing.locales.flatMap((locale) =>
		STATIC_ROUTES.map((route) => ({
			url: `${base}/${locale}${route}`,
			lastModified: now,
		})),
	);

	const productEntries: MetadataRoute.Sitemap = routing.locales.flatMap((locale) =>
		products
			.filter((product) => product.slug)
			.map((product) => ({
				url: `${base}/${locale}/products/${product.slug}`,
				lastModified: now,
			})),
	);

	return [...staticEntries, ...productEntries];
}
