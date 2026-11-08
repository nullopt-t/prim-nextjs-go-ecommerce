import { describe, expect, it } from "vitest";
import { normalizeTiers, resolveTier } from "./tierPricing";

describe("normalizeTiers", () => {
	it("drops invalid rows and sorts by threshold", () => {
		const tiers = normalizeTiers([
			{ minQuantity: 10, discountPercent: 15 },
			{ minQuantity: "nope", discountPercent: 5 },
			{ minQuantity: 3, discountPercent: 5 },
			{ minQuantity: 0, discountPercent: 50 },
			null,
			{ minQuantity: 5 },
		]);
		expect(tiers).toEqual([
			{ minQuantity: 3, discountPercent: 5 },
			{ minQuantity: 10, discountPercent: 15 },
		]);
	});

	it("returns [] for non-array input", () => {
		expect(normalizeTiers(undefined)).toEqual([]);
		expect(normalizeTiers({})).toEqual([]);
	});
});

describe("resolveTier", () => {
	const tiers = [
		{ minQuantity: 3, discountPercent: 5 },
		{ minQuantity: 10, discountPercent: 15 },
	];

	it("falls back to the base price below the first threshold", () => {
		expect(resolveTier(100, tiers, 2)).toEqual({
			unitPrice: 100,
			activeTier: null,
			savingsPerUnit: 0,
		});
	});

	it("applies the highest met threshold", () => {
		expect(resolveTier(100, tiers, 3).unitPrice).toBe(95);
		expect(resolveTier(100, tiers, 12).unitPrice).toBe(85);
	});

	it("prefers an explicit unitPrice over percent", () => {
		const withPrice = [{ minQuantity: 3, unitPrice: 88, discountPercent: 50 }];
		expect(resolveTier(100, withPrice, 3).unitPrice).toBe(88);
	});

	it("resolves the highest met threshold even for unsorted tiers", () => {
		const unsorted = [
			{ minQuantity: 10, discountPercent: 15 },
			{ minQuantity: 3, discountPercent: 5 },
		];
		expect(resolveTier(100, unsorted, 5).unitPrice).toBe(95);
		expect(resolveTier(100, unsorted, 10).unitPrice).toBe(85);
	});

	it("drops discounts above 100% instead of yielding a negative price", () => {
		expect(normalizeTiers([{ minQuantity: 2, discountPercent: 150 }])).toEqual([]);
		expect(resolveTier(100, [{ minQuantity: 1, discountPercent: 150 }], 5).unitPrice).toBeGreaterThanOrEqual(0);
	});
});
