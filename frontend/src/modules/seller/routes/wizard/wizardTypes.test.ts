import { describe, expect, it } from "vitest";
import { bundlePayloadItems, isValidDownloadUrl } from "./wizardTypes";

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

	it("filters out invalid/deleted/zero-price products when predicate is provided", () => {
		const validProducts = new Set(["a"]);
		expect(
			bundlePayloadItems(
				[
					{ productId: "a", quantity: "2" },
					{ productId: "deleted-item", quantity: "1" },
				],
				(id) => validProducts.has(id),
			),
		).toEqual([{ productId: "a", quantity: 2 }]);
	});
});

describe("isValidDownloadUrl", () => {
	it("accepts valid http and https URLs", () => {
		expect(isValidDownloadUrl("https://example.com/file.zip")).toBe(true);
		expect(isValidDownloadUrl("http://cdn.store.com/download/pkg.tar.gz?token=abc")).toBe(true);
	});

	it("rejects non-http(s) protocols and invalid URLs", () => {
		expect(isValidDownloadUrl("ftp://example.com/file.zip")).toBe(false);
		expect(isValidDownloadUrl("javascript:alert(1)")).toBe(false);
		expect(isValidDownloadUrl("data:text/plain,hello")).toBe(false);
		expect(isValidDownloadUrl("file:///etc/passwd")).toBe(false);
		expect(isValidDownloadUrl("not-a-url")).toBe(false);
		expect(isValidDownloadUrl("")).toBe(false);
		expect(isValidDownloadUrl("   ")).toBe(false);
	});
});
