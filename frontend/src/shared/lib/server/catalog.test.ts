import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import {
	getProducts,
	getProductBySlug,
	getCategories,
	getBrands,
	getCatalogSnapshot,
	getCatalogFacets,
	CATALOG_REVALIDATE_SECONDS,
	PRODUCTS_TAG,
	CATEGORIES_TAG,
	BRANDS_TAG,
} from "@/shared/lib/server/catalog";

const fetchMock = vi.fn();

beforeEach(() => {
	process.env.NEXT_PUBLIC_API_URL = "http://api:8081";
	fetchMock.mockReset();
	(globalThis as unknown as { fetch?: typeof fetch }).fetch = fetchMock;
});

afterEach(() => {
	delete process.env.NEXT_PUBLIC_API_URL;
});

function respondWith(data: unknown, status = 200) {
	fetchMock.mockResolvedValueOnce({
		ok: status >= 200 && status < 300,
		status,
		json: async () => data,
	} as Response);
}

describe("server catalog fetchers", () => {
	it("getProducts builds the list request and normalizes products", async () => {
		respondWith({ data: [{ id: "p1", slug: "shoe", title: "Shoe", price: "120.00" }] });

		const products = await getProducts({ category: "audio", search: "zzz", pageSize: 24 });

		expect(products).toEqual([expect.objectContaining({ id: "p1", title: "Shoe" })]);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe("http://api:8081/api/v1/products?pageSize=24&category=audio&search=zzz");
		expect(init).toMatchObject({
			next: { revalidate: 300, tags: ["catalog:products"] },
			headers: { Accept: "application/json" },
		});
	});

	it("getProducts serializes attribute facets as repeated key:value params", async () => {
		respondWith({ data: [] });

		await getProducts({ pageSize: 24, attributes: { color: "black", size: "XL" } });

		const [url] = fetchMock.mock.calls[0];
		expect(url).toBe(
			"http://api:8081/api/v1/products?pageSize=24&attribute=color%3Ablack&attribute=size%3AXL",
		);
	});

	it("getProducts sends prices in minor units and backend sort syntax", async () => {
		respondWith({ data: [] });

		await getProducts({ pageSize: 24, minPrice: 100, maxPrice: 250, sort: "price:desc", inStock: true });

		const [url] = fetchMock.mock.calls[0];
		expect(url).toBe(
			"http://api:8081/api/v1/products?pageSize=24&inStock=true&min_price=10000&max_price=25000&sort=-price",
		);
	});

	it("getProducts omits the attribute param when no facets are selected", async () => {
		respondWith({ data: [] });

		await getProducts({ attributes: {} });

		const [url] = fetchMock.mock.calls[0];
		expect(url).not.toContain("attribute=");
	});

	it("getProducts defaults pageSize to 100", async () => {
		respondWith({ data: [] });
		await getProducts();
		const [url] = fetchMock.mock.calls[0];
		expect(url).toBe("http://api:8081/api/v1/products?pageSize=100");
	});

	it("getProductBySlug fetches a single product with its tag", async () => {
		respondWith({ data: { id: "p1", slug: "shoe", title: "Shoe", variants: [] } });

		const product = await getProductBySlug("shoe");

		expect(product?.id).toBe("p1");
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe("http://api:8081/api/v1/products/shoe");
		expect(init).toMatchObject({ next: { tags: ["catalog:products", "product:shoe"] } });
	});

	it("getProductBySlug returns null when nothing is returned", async () => {
		respondWith(null);
		expect(await getProductBySlug("missing")).toBeNull();
	});

	it("getCategories fetches the category list", async () => {
		respondWith({ data: [{ id: "c1", name: "Shoes" }] });
		const categories = await getCategories();
		expect(categories).toEqual([{ id: "c1", name: "Shoes" }]);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe("http://api:8081/api/v1/categories?pageSize=50");
		expect(init).toMatchObject({ next: { tags: ["catalog:categories"] } });
	});

	it("getBrands fetches the brand list", async () => {
		respondWith({ data: [{ id: "b1", name: "Acme" }] });
		const brands = await getBrands();
		expect(brands).toEqual([{ id: "b1", name: "Acme" }]);
		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe("http://api:8081/api/v1/brands?pageSize=50");
		expect(init).toMatchObject({ next: { tags: ["catalog:brands"] } });
	});

	it("getCatalogSnapshot combines products, categories and brands", async () => {
		respondWith({ data: [{ id: "p1", slug: "s", title: "S", variants: [] }] });
		respondWith({ data: [{ id: "c1", name: "C" }] });
		respondWith({ data: [{ id: "b1", name: "B" }] });

		const snapshot = await getCatalogSnapshot();
		expect(snapshot.products).toHaveLength(1);
		expect(snapshot.categories).toHaveLength(1);
		expect(snapshot.brands).toHaveLength(1);
	});

	it("getCatalogFacets returns only categories and brands", async () => {
		respondWith({ data: [{ id: "c1", name: "C" }] });
		respondWith({ data: [{ id: "b1", name: "B" }] });

		const facets = await getCatalogFacets();
		expect(facetCategories(facets)).toEqual([{ id: "c1", name: "C" }]);
		expect(facetBrands(facets)).toEqual([{ id: "b1", name: "B" }]);
		expect("products" in facets).toBe(false);
	});

	it("returns empty results when the API responds unsuccessfully", async () => {
		fetchMock.mockResolvedValueOnce({ ok: false, status: 500 } as Response);
		expect(await getProducts()).toEqual([]);
	});

	it("uses tag constants", () => {
		expect(CATALOG_REVALIDATE_SECONDS).toBe(300);
		expect(PRODUCTS_TAG).toBe("catalog:products");
		expect(CATEGORIES_TAG).toBe("catalog:categories");
		expect(BRANDS_TAG).toBe("catalog:brands");
	});
});

function facetCategories(facets: { categories: unknown[]; brands: unknown[] }) {
	return facets.categories;
}
function facetBrands(facets: { categories: unknown[]; brands: unknown[] }) {
	return facets.brands;
}
