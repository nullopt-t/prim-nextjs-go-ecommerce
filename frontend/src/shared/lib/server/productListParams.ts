import type { CatalogQuery } from "@/shared/lib/server/catalog";

/**
 * URL params for the product listing surfaces (`/products` and `/search`).
 * Both pages parse the same shape so a shared link behaves identically on
 * either route. Prices stay in major units here; `buildProductQuery`
 * converts them to the backend's minor units at request time.
 */
export type ProductListSearchParams = {
	category?: string;
	q?: string;
	brand?: string;
	inStock?: string;
	minPrice?: string;
	maxPrice?: string;
	sort?: string;
	page?: string;
	/** Comma-separated or repeated "key:value" facet pairs, e.g. color:black,size:XL. */
	attribute?: string | string[];
};

function parseNumber(value?: string): number | undefined {
	if (!value) return undefined;
	const parsed = Number(value);
	return Number.isFinite(parsed) && parsed >= 0 ? parsed : undefined;
}

export function parsePage(value?: string): number {
	const parsed = Number(value);
	return Number.isFinite(parsed) && parsed >= 1 ? Math.floor(parsed) : 1;
}

export function parseAttributes(value?: string | string[]): Record<string, string> | undefined {
	if (!value) return undefined;
	const out: Record<string, string> = {};
	const list = Array.isArray(value) ? value : [value];
	for (const item of list) {
		for (const part of item.split(",")) {
			const colon = part.indexOf(":");
			if (colon <= 0 || colon >= part.length - 1) continue;
			const k = part.slice(0, colon).trim();
			const v = part.slice(colon + 1).trim();
			if (k && v) out[k] = v;
		}
	}
	return Object.keys(out).length > 0 ? out : undefined;
}

/** Parses raw URL search params into a backend catalog query. */
export function toCatalogQuery(params: ProductListSearchParams, pageSize: number): CatalogQuery {
	return {
		category: params.category && params.category.toLowerCase() !== "all" ? params.category : undefined,
		search: params.q?.trim() ? params.q.trim() : undefined,
		brand: params.brand?.trim() ? params.brand.trim() : undefined,
		inStock: params.inStock === "true",
		minPrice: parseNumber(params.minPrice),
		maxPrice: parseNumber(params.maxPrice),
		sort: params.sort?.trim() ? params.sort.trim() : undefined,
		attributes: parseAttributes(params.attribute),
		page: parsePage(params.page),
		pageSize,
	};
}
