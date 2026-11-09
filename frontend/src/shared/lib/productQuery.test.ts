import { describe, expect, it } from "vitest";
import { buildProductQuery } from "@/shared/lib/productQuery";

describe("buildProductQuery", () => {
	it("converts major-unit prices to backend minor units", () => {
		const q = buildProductQuery({ minPrice: 100, maxPrice: 250.5 });
		expect(q.get("min_price")).toBe("10000");
		expect(q.get("max_price")).toBe("25050");
	});

	it("translates URL sort values to backend sort syntax", () => {
		expect(buildProductQuery({ sort: "price:asc" }).get("sort")).toBe("price");
		expect(buildProductQuery({ sort: "price:desc" }).get("sort")).toBe("-price");
		expect(buildProductQuery({ sort: "-title" }).get("sort")).toBe("-title");
	});

	it("sends one repeated attribute param per facet", () => {
		const q = buildProductQuery({ attributes: { color: "black", size: "XL" } });
		expect(q.getAll("attribute")).toEqual(["color:black", "size:XL"]);
	});

	it("skips empty and 'all' category values", () => {
		expect(buildProductQuery({ category: "all" }).has("category")).toBe(false);
		expect(buildProductQuery({ category: "ALL" }).has("category")).toBe(false);
		expect(buildProductQuery({ category: "audio" }).get("category")).toBe("audio");
	});

	it("only sends inStock when true", () => {
		expect(buildProductQuery({ inStock: true }).get("inStock")).toBe("true");
		expect(buildProductQuery({ inStock: false }).has("inStock")).toBe(false);
	});
});
