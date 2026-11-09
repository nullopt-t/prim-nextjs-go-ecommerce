/**
 * Shared serialization of storefront product query params for the backend
 * `/api/v1/products` contract. Used by both the server fetchers and the
 * client catalog service so the two can never drift apart:
 * - prices arrive in major units (what the shopper types) and the backend
 *   expects minor units (cents/piasters);
 * - the URL sort syntax (`price:asc`/`price:desc`) is translated to the
 *   backend's (`price`/`-price`);
 * - attribute facets are sent as repeated `?attribute=k:v` params, which the
 *   backend binds to a slice (a comma-joined single param would corrupt
 *   values containing commas or colons).
 */
export interface ProductQuery {
	pageSize?: number;
	page?: number;
	cursor?: string;
	category?: string;
	search?: string;
	brand?: string;
	inStock?: boolean;
	minPrice?: number;
	maxPrice?: number;
	sort?: string;
	attributes?: Record<string, string>;
}

function backendSort(sort: string): string {
	if (sort === "price:asc") return "price";
	if (sort === "price:desc") return "-price";
	return sort;
}

export function buildProductQuery(query: ProductQuery): URLSearchParams {
	const params = new URLSearchParams();
	if (query.pageSize) params.set("pageSize", String(query.pageSize));
	if (query.page) params.set("page", String(query.page));
	if (query.cursor) params.set("cursor", query.cursor);
	if (query.category && query.category.toLowerCase() !== "all") {
		params.set("category", query.category);
	}
	if (query.search) params.set("search", query.search);
	if (query.brand) params.set("brand", query.brand);
	if (query.inStock) params.set("inStock", "true");
	if (query.minPrice !== undefined) params.set("min_price", String(Math.round(query.minPrice * 100)));
	if (query.maxPrice !== undefined) params.set("max_price", String(Math.round(query.maxPrice * 100)));
	if (query.sort) params.set("sort", backendSort(query.sort));
	if (query.attributes) {
		for (const [k, v] of Object.entries(query.attributes)) {
			if (k.trim() === "" || v.trim() === "") continue;
			params.append("attribute", `${k.trim()}:${v.trim()}`);
		}
	}
	return params;
}
