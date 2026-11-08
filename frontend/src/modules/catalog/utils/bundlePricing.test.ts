import { describe, expect, it } from "vitest";
import { computeBundleSavings, resolveItemPrice } from "./bundlePricing";

describe("resolveItemPrice", () => {
	it("prefers the default variant price", () => {
		expect(
			resolveItemPrice({
				price: 100,
				variants: [
					{ price: 90, isDefault: false },
					{ extractedPrice: 80, isDefault: true },
				],
			}),
		).toBe(80);
	});

	it("matches a specific variant by id or sku", () => {
		expect(
			resolveItemPrice(
				{
					price: 100,
					variants: [
						{ id: "var-1", price: 90, isDefault: true },
						{ id: "var-2", sku: "SKU-2", extractedPrice: 75, isDefault: false },
					],
				},
				"SKU-2",
			),
		).toBe(75);
	});

	it("falls back to the default variant when target variant is missing or not found", () => {
		expect(
			resolveItemPrice(
				{
					price: 100,
					variants: [
						{ id: "var-1", price: 90, isDefault: true },
						{ id: "var-2", price: 75, isDefault: false },
					],
				},
				"var-nonexistent",
			),
		).toBe(90);
	});

	it("falls back to the product price", () => {
		expect(resolveItemPrice({ extractedPrice: 42 })).toBe(42);
		expect(resolveItemPrice({})).toBe(0);
	});
});

describe("computeBundleSavings", () => {
	it("sums item prices weighted by quantity", () => {
		const result = computeBundleSavings(
			[
				{ slug: "a", quantity: 2, price: 50 },
				{ slug: "b", quantity: 1, price: 30 },
			],
			100,
		);
		expect(result).toEqual({ itemsTotal: 130, savings: 30, percentOff: 23 });
	});

	it("returns null when the kit costs as much or more than its parts", () => {
		expect(computeBundleSavings([{ slug: "a", quantity: 1, price: 10 }], 12)).toBeNull();
		expect(computeBundleSavings([], 5)).toBeNull();
		expect(computeBundleSavings([{ slug: "a", quantity: 1, price: 10 }], 0)).toBeNull();
	});
});
