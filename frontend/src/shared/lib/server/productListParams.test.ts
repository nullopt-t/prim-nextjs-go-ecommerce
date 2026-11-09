import { describe, expect, it } from "vitest";
import { parseAttributes, parsePage, toCatalogQuery } from "@/shared/lib/server/productListParams";

describe("toCatalogQuery", () => {
	it("maps URL params to a catalog query", () => {
		expect(
			toCatalogQuery(
				{ q: "  shoes ", brand: "Acme", inStock: "true", minPrice: "100", sort: "price:asc", page: "2" },
				24,
			),
		).toEqual({
			category: undefined,
			search: "shoes",
			brand: "Acme",
			inStock: true,
			minPrice: 100,
			maxPrice: undefined,
			sort: "price:asc",
			attributes: undefined,
			page: 2,
			pageSize: 24,
		});
	});

	it("drops invalid numbers and non-true inStock", () => {
		const q = toCatalogQuery({ minPrice: "-5", maxPrice: "abc", inStock: "false" }, 24);
		expect(q.minPrice).toBeUndefined();
		expect(q.maxPrice).toBeUndefined();
		expect(q.inStock).toBe(false);
	});

	it("treats category=all as no category", () => {
		expect(toCatalogQuery({ category: "all" }, 24).category).toBeUndefined();
		expect(toCatalogQuery({ category: "audio" }, 24).category).toBe("audio");
	});
});

describe("parseAttributes", () => {
	it("accepts repeated and comma-joined forms", () => {
		expect(parseAttributes(["color:red", "size:xl"])).toEqual({ color: "red", size: "xl" });
		expect(parseAttributes("color:red,size:xl")).toEqual({ color: "red", size: "xl" });
		expect(parseAttributes("color:")).toBeUndefined();
	});
});

describe("parsePage", () => {
	it("falls back to 1", () => {
		expect(parsePage(undefined)).toBe(1);
		expect(parsePage("0")).toBe(1);
		expect(parsePage("abc")).toBe(1);
		expect(parsePage("3.9")).toBe(3);
	});
});
