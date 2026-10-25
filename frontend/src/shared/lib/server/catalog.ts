import { normalizeProduct } from "@/modules/catalog/model/normalizeProduct";

export const CATALOG_REVALIDATE_SECONDS = 300;
export const PRODUCTS_TAG = "catalog:products";
export const CATEGORIES_TAG = "catalog:categories";
export const BRANDS_TAG = "catalog:brands";

function getApiBaseUrl() {
	return (process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "").replace(/\/$/, "");
}

function productTag(slug: string) {
	return `product:${slug}`;
}

async function fetchJson(path: string, tags: string[] = []) {
	const baseUrl = getApiBaseUrl();
	if (!baseUrl) return null;

	try {
		const response = await fetch(`${baseUrl}${path}`, {
			next: { revalidate: CATALOG_REVALIDATE_SECONDS, tags },
			headers: { Accept: "application/json" },
		});
		if (!response.ok) return null;
		const payload = await response.json();
		return payload?.data ?? null;
	} catch {
		return null;
	}
}

function asList(payload: any): any[] {
	if (Array.isArray(payload)) return payload;
	return Array.isArray(payload?.data) ? payload.data : [];
}

/**
 * Server-side product lookup for the product detail route. Returns the
 * normalized product so the page and its metadata share one payload.
 */
export async function getProductBySlug(slug: string) {
	const raw = await fetchJson(`/api/v1/products/${encodeURIComponent(slug)}`, [
		PRODUCTS_TAG,
		productTag(slug),
	]);
	return raw ? normalizeProduct(raw) : null;
}

export interface CatalogQuery {
	pageSize?: number;
	category?: string;
	search?: string;
}

export async function getProducts(query: CatalogQuery = {}) {
	const params = new URLSearchParams();
	params.set("pageSize", String(query.pageSize ?? 100));
	if (query.category && query.category.toLowerCase() !== "all") {
		params.set("category", query.category);
	}
	if (query.search) params.set("search", query.search);

	const raw = await fetchJson(`/api/v1/products?${params.toString()}`, [PRODUCTS_TAG]);
	return asList(raw).map(normalizeProduct);
}

export async function getCategories(pageSize = 50) {
	const raw = await fetchJson(`/api/v1/categories?pageSize=${pageSize}`, [CATEGORIES_TAG]);
	return asList(raw);
}

export async function getBrands(pageSize = 50) {
	const raw = await fetchJson(`/api/v1/brands?pageSize=${pageSize}`, [BRANDS_TAG]);
	return asList(raw);
}

export interface CatalogSnapshot {
	products: ReturnType<typeof normalizeProduct>[];
	categories: any[];
	brands: any[];
}

/** One cached catalog read shared by the storefront shell and its pages. */
export async function getCatalogSnapshot(pageSize = 100): Promise<CatalogSnapshot> {
	const [products, categories, brands] = await Promise.all([
		getProducts({ pageSize }),
		getCategories(),
		getBrands(),
	]);
	return { products, categories, brands };
}

/**
 * Navigation and filtering only need the category/brand facets. Use this for
 * storefront routes that do not render a product feed so the RSC payload does
 * not carry up to 100 serialized products.
 */
export async function getCatalogFacets(): Promise<
	Pick<CatalogSnapshot, "categories" | "brands">
> {
	const [categories, brands] = await Promise.all([getCategories(), getBrands()]);
	return { categories, brands };
}

