import type { MetadataRoute } from "next";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import sitemap from "./sitemap";

const fetchMock = vi.fn();
const originalSiteUrl = process.env.NEXT_PUBLIC_SITE_URL;
const originalApiUrl = process.env.NEXT_PUBLIC_API_URL;

beforeEach(() => {
	fetchMock.mockReset();
	(globalThis as unknown as { fetch?: typeof fetch }).fetch = fetchMock;
	process.env.NEXT_PUBLIC_SITE_URL = "https://example.com";
	process.env.NEXT_PUBLIC_API_URL = "http://api:8081";
});

afterEach(() => {
	if (originalSiteUrl === undefined) delete process.env.NEXT_PUBLIC_SITE_URL;
	else process.env.NEXT_PUBLIC_SITE_URL = originalSiteUrl;
	if (originalApiUrl === undefined) delete process.env.NEXT_PUBLIC_API_URL;
	else process.env.NEXT_PUBLIC_API_URL = originalApiUrl;
});

function respondWith(products: unknown[]) {
	fetchMock.mockResolvedValueOnce({
		ok: true,
		status: 200,
		json: async () => ({ data: products }),
	} as Response);
}

describe("sitemap", () => {
	it("lists static routes and products for every locale", async () => {
		respondWith([{ id: "p1", slug: "shoe" }]);

		const urls = (await sitemap()).map((entry: MetadataRoute.Sitemap[number]) => entry.url);

		expect(urls).toContain("https://example.com/en");
		expect(urls).toContain("https://example.com/en/products/shoe");
		expect(urls).toContain("https://example.com/ar");
		expect(urls).toContain("https://example.com/ar/products/shoe");
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it("paginates products beyond the first page instead of capping at one page", async () => {
		const fullPage = Array.from({ length: 200 }, (_, i) => ({
			id: `p${i}`,
			slug: `shoe-${i}`,
		}));
		const lastPage = [{ id: "last", slug: "last-one" }];
		fetchMock.mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => ({ data: fullPage }),
		} as Response);
		fetchMock.mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => ({ data: lastPage }),
		} as Response);

		const urls = (await sitemap()).map((entry: MetadataRoute.Sitemap[number]) => entry.url);

		expect(urls).toContain("https://example.com/en/products/last-one");
		expect(urls).toContain("https://example.com/ar/products/last-one");
		expect(fetchMock).toHaveBeenCalledTimes(2);
	});
});
