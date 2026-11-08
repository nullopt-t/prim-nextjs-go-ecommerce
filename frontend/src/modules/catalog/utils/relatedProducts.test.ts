import { describe, expect, it } from "vitest";
import { selectCartRecommendations, selectRelatedProducts } from "./relatedProducts";

const TARGET = {
	id: "p1",
	category: "Audio",
	brand: "Sennheiser",
	tags: ["headphones"],
};

const CATALOG = [
	{ id: "p2", category: "Audio", brand: "Other", tags: ["speakers"] },
	{ id: "p3", category: "Audio", brand: "Sennheiser", tags: ["headphones", "wireless"] },
	{ id: "p4", category: "Toys", brand: "Sennheiser", tags: ["headphones"] },
	{ id: "p5", category: "Toys", brand: "Other", tags: ["toys"] },
];

describe("selectRelatedProducts", () => {
	it("ranks same-tag + same-category + same-brand first and excludes self", () => {
		const related = selectRelatedProducts(TARGET, CATALOG, ["p1"], 5);
		expect(related.map((p) => p.id)).toEqual(["p3", "p4", "p2"]);
	});

	it("drops unrelated products", () => {
		const related = selectRelatedProducts(TARGET, CATALOG, [], 5);
		expect(related.map((p) => p.id)).not.toContain("p5");
	});
});

describe("selectCartRecommendations", () => {
	it("matches same category/brand as cart items and excludes the cart", () => {
		const cart = [{ id: "p1", category: "Audio", brand: "Sennheiser", tags: [] }];
		const picks = selectCartRecommendations(cart, CATALOG, 6);
		expect(picks.map((p) => p.id)).toEqual(["p3", "p2", "p4"]);
	});

	it("returns nothing when the cart is empty", () => {
		expect(selectCartRecommendations([], CATALOG)).toEqual([]);
	});
});
