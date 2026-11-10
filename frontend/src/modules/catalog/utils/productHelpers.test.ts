import { describe, expect, it } from "vitest";
import { resolveDeliveryWindow, resolveVariantByAttributes } from "./productHelpers";
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

describe("resolveDeliveryWindow", () => {
	const monday = new Date("2025-11-10T00:00:00Z");

	it("turns a day range into localized arrival dates", () => {
		expect(resolveDeliveryWindow("2 - 4 business days", "en", monday)).toEqual({
			start: "Wed, Nov 12",
			end: "Fri, Nov 14",
		});
	});

	it("collapses a single-day window to the same date", () => {
		expect(resolveDeliveryWindow("3", "en", monday)).toEqual({
			start: "Thu, Nov 13",
			end: "Thu, Nov 13",
		});
	});

	it("returns null for settings it cannot read as a day range", () => {
		expect(resolveDeliveryWindow(undefined, "en", monday)).toBeNull();
		expect(resolveDeliveryWindow("soon", "en", monday)).toBeNull();
		expect(resolveDeliveryWindow("24 hours", "en", monday)).toBeNull();
	});
});
