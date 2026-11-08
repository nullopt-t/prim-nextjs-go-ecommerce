import { normalizeProduct, type RawProductPayload } from "@/modules/catalog/model/normalizeProduct";
import type { Brand, Category, ProductDetailsEntity } from "@/modules/catalog/types";
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
	const params = new URLSearchParams();
	params.set("pageSize", String(query.pageSize ?? 100));
	if (query.page) params.set("page", String(query.page));
	if (query.category && query.category.toLowerCase() !== "all") {
		params.set("category", query.category);
	}
	if (query.search) params.set("search", query.search);
	if (query.brand) params.set("brand", query.brand);
	if (query.inStock) params.set("inStock", "true");
	if (query.minPrice !== undefined) params.set("min_price", String(query.minPrice));
	if (query.maxPrice !== undefined) params.set("max_price", String(query.maxPrice));
	if (query.sort) params.set("sort", query.sort);
	// Backend binds a single `attribute` query param: comma-separated
	// "key:value" pairs (e.g. attribute=color:black,size:XL). Repeated
	// params would only bind the last value.
	if (query.attributes) {
		const pairs = Object.entries(query.attributes)
			.filter(([k, v]) => k.trim() !== "" && v.trim() !== "")
			.map(([k, v]) => `${k.trim()}:${v.trim()}`);
		if (pairs.length > 0) params.set("attribute", pairs.join(","));
	}
	return params;
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
