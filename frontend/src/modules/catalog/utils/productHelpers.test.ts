import { describe, expect, it } from "vitest";
import { resolveVariantByAttributes } from "./productHelpers";
import type { ProductDetailsEntity } from "../types";

const product = {
	variationAxes: [
		{ key: "color", label: "Color", values: ["Red", "Blue"] },
		{ key: "size", label: "Size", values: ["M", "L"] },
	],
	variants: [
		{ id: "v1", attributes: { color: "Red", size: "M" }, isDefault: true },
		{ id: "v2", attributes: { color: "Red", size: "L" } },
		{ id: "v3", attributes: { color: "Blue", size: "M" } },
	],
} as unknown as ProductDetailsEntity;

describe("resolveVariantByAttributes", () => {
	it("matches the full selection", () => {
		expect(resolveVariantByAttributes(product, { color: "Red", size: "L" })?.id).toBe("v2");
	});

	it("matches a partial selection instead of falling back to default", () => {
		expect(resolveVariantByAttributes(product, { color: "Blue" })?.id).toBe("v3");
	});

	it("falls back to the default variant for an unknown combination", () => {
		expect(resolveVariantByAttributes(product, { color: "Blue", size: "L" })?.id).toBe("v1");
		expect(resolveVariantByAttributes(product, {})?.id).toBe("v1");
	});
});
