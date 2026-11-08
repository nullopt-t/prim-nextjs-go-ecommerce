import { describe, expect, it } from "vitest";
import { bundlePayloadItems } from "./wizardTypes";

describe("bundle payload items", () => {
	it("keeps chosen products with a positive whole quantity", () => {
		expect(
			bundlePayloadItems([
				{ productId: "a", quantity: "2" },
				{ productId: "b", quantity: "1" },
			]),
		).toEqual([
			{ productId: "a", quantity: 2 },
			{ productId: "b", quantity: 1 },
		]);
	});

	it("drops empty rows and non-positive or fractional quantities", () => {
		expect(
			bundlePayloadItems([
				{ productId: "", quantity: "3" },
				{ productId: "a", quantity: "0" },
				{ productId: "b", quantity: "1.5" },
				{ productId: "c", quantity: "" },
			]),
		).toEqual([]);
	});
});
