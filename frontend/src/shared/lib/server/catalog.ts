import { normalizeProduct, type RawProductPayload } from "@/modules/catalog/model/normalizeProduct";
import type { Brand, Category, ProductDetailsEntity, Review } from "@/modules/catalog/types";
import { buildProductQuery } from "@/shared/lib/productQuery";
import { getServerApiBaseUrl } from "@/shared/lib/server/apiBaseUrl";

// Server-side catalog revalidation window. Caching is on by default; set
// CATALOG_REVALIDATE_SECONDS=0 locally to always read fresh data while
// developing (Next.js ISR treats 0 as "no cache").
function resolveRevalidateSeconds(): number {
	const raw = process.env.CATALOG_REVALIDATE_SECONDS?.trim();
	if (!raw) {
		return 300;
	}
	const parsed = Number(raw);
	return Number.isFinite(parsed) && parsed >= 0 ? parsed : 300;
}

export const CATALOG_REVALIDATE_SECONDS = resolveRevalidateSeconds();
export const PRODUCTS_TAG = "catalog:products";
export const CATEGORIES_TAG = "catalog:categories";
export const BRANDS_TAG = "catalog:brands";

function getApiBaseUrl() {
	return getServerApiBaseUrl() || "";
}

function productTag(slug: string) {
	return `product:${slug}`;
}

async function fetchJson<T = unknown>(path: string, tags: string[] = []): Promise<T | null> {
	const baseUrl = getApiBaseUrl();
	if (!baseUrl) return null;

	try {
		const response = await fetch(`${baseUrl}${path}`, {
			next: { revalidate: CATALOG_REVALIDATE_SECONDS, tags },
			headers: { Accept: "application/json" },
			signal: AbortSignal.timeout(3000),
		});
		if (!response.ok) return null;
		const payload = await response.json();
		return (payload?.data ?? null) as T | null;
	} catch {
		return null;
	}
}

async function fetchEnvelope<T = unknown>(path: string, tags: string[] = []): Promise<{ data: T | null; meta: Record<string, unknown> | null }> {
	const baseUrl = getApiBaseUrl();
	if (!baseUrl) return { data: null, meta: null };

	try {
		const response = await fetch(`${baseUrl}${path}`, {
			next: { revalidate: CATALOG_REVALIDATE_SECONDS, tags },
			headers: { Accept: "application/json" },
			signal: AbortSignal.timeout(3000),
		});
		if (!response.ok) return { data: null, meta: null };
		const payload = (await response.json()) as { data?: T | null; meta?: Record<string, unknown> };
		return { data: (payload?.data ?? null) as T | null, meta: payload?.meta ?? null };
	} catch {
		return { data: null, meta: null };
	}
}

function asList<T>(payload: unknown): T[] {
	if (Array.isArray(payload)) return payload as T[];
	if (payload && typeof payload === "object" && "data" in payload && Array.isArray((payload as { data: unknown }).data)) {
		return (payload as { data: T[] }).data;
	}
	return [];
}

/**
 * Server-side product lookup for the product detail route. Returns the
 * normalized product so the page and its metadata share one payload.
 */
export async function getProductBySlug(slug: string): Promise<ProductDetailsEntity | null> {
	const raw = await fetchJson(`/api/v1/products/${encodeURIComponent(slug)}`, [
		PRODUCTS_TAG,
		productTag(slug),
	]);
	return raw ? (normalizeProduct(raw as RawProductPayload) as ProductDetailsEntity | null) : null;
}

/**
 * Reviews fetched per page on the product detail route. The server renders
 * the first page into the HTML and the PDP appends further pages when the
 * shopper asks for more, so the visible list can reach every review.
 */
export const REVIEWS_PAGE_SIZE = 20;

/**
 * Server-side review list for the product detail route so reviews are part of
 * the initial HTML instead of appearing only after hydration.
 */
export async function getReviewsByProduct(
	productId: string,
	page = 1,
	pageSize = REVIEWS_PAGE_SIZE,
): Promise<Review[]> {
	const raw = await fetchJson<Review[]>(
		`/api/v1/reviews?productId=${encodeURIComponent(productId)}&page=${page}&pageSize=${pageSize}`,
		[`product-reviews:${productId}`],
	);
	return asList<Review>(raw);
}

export interface CatalogQuery {
	pageSize?: number;
	page?: number;
	category?: string;
	search?: string;
	brand?: string;
	inStock?: boolean;
	minPrice?: number;
	maxPrice?: number;
	sort?: string;
	/** Selected variant attribute facets, e.g. { color: "black", size: "XL" }. */
	attributes?: Record<string, string>;
}

function productParams(query: CatalogQuery): URLSearchParams {
	return buildProductQuery({ ...query, pageSize: query.pageSize ?? 100 });
}

export async function getProducts(query: CatalogQuery = {}): Promise<ProductDetailsEntity[]> {
	const raw = await fetchJson(`/api/v1/products?${productParams(query).toString()}`, [PRODUCTS_TAG]);
	return (asList<RawProductPayload>(raw)
		.map(normalizeProduct)
		.filter((p) => p !== null)) as unknown as ProductDetailsEntity[];
}

export interface ProductsWithMeta {
	products: ProductDetailsEntity[];
	page: number;
	pageSize: number;
	totalItems: number;
	totalPages: number;
}

export async function getProductsWithMeta(query: CatalogQuery = {}): Promise<ProductsWithMeta> {
	const pageSize = query.pageSize ?? 100;
	const page = query.page ?? 1;
	const { data, meta } = await fetchEnvelope<RawProductPayload[]>(
		`/api/v1/products?${productParams({ ...query, pageSize, page }).toString()}`,
		[PRODUCTS_TAG],
	);
	const products = (asList<RawProductPayload>(data ?? [])
		.map(normalizeProduct)
		.filter((p) => p !== null)) as unknown as ProductDetailsEntity[];

	const totalItems = typeof meta?.totalItems === "number" ? meta.totalItems : products.length;
	const totalPages =
		typeof meta?.totalPages === "number"
			? meta.totalPages
			: Math.max(1, Math.ceil(totalItems / pageSize));
	return { products, page, pageSize, totalItems, totalPages };
}

export async function getCategories(pageSize = 50): Promise<Category[]> {
	const raw = await fetchJson<Category[]>(`/api/v1/categories?pageSize=${pageSize}`, [CATEGORIES_TAG]);
	return asList<Category>(raw);
}

export async function getBrands(pageSize = 50): Promise<Brand[]> {
	const raw = await fetchJson<Brand[]>(`/api/v1/brands?pageSize=${pageSize}`, [BRANDS_TAG]);
	return asList<Brand>(raw);
}

export interface CatalogSnapshot {
	products: ProductDetailsEntity[];
	categories: Category[];
	brands: Brand[];
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
