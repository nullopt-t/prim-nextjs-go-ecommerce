import { describe, expect, it } from "vitest";
import { bundlePayloadItems, isValidDownloadUrl, parseMajorPrice, parseWholeStock } from "./wizardTypes";

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

	it("merges duplicate rows for the same product into one summed line", () => {
		expect(
			bundlePayloadItems([
				{ productId: "a", quantity: "1" },
				{ productId: "b", quantity: "1" },
				{ productId: "a", quantity: "2" },
			]),
		).toEqual([
			{ productId: "a", quantity: 3 },
			{ productId: "b", quantity: 1 },
		]);
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

describe("parseMajorPrice", () => {
	it("converts major units to rounded minor units", () => {
		expect(parseMajorPrice("19.99")).toBe(1999);
		expect(parseMajorPrice("10")).toBe(1000);
		expect(parseMajorPrice("0.01")).toBe(1);
	});

	it("rejects sub-cent prices that would round to zero", () => {
		expect(parseMajorPrice("0.001")).toBeUndefined();
		expect(parseMajorPrice("0.004")).toBeUndefined();
	});

	it("rejects empty, negative, and non-numeric prices", () => {
		expect(parseMajorPrice("")).toBeUndefined();
		expect(parseMajorPrice("0")).toBeUndefined();
		expect(parseMajorPrice("-5")).toBeUndefined();
		expect(parseMajorPrice("abc")).toBeUndefined();
	});
});

describe("parseWholeStock", () => {
	it("accepts whole non-negative counts", () => {
		expect(parseWholeStock("0")).toBe(0);
		expect(parseWholeStock("42")).toBe(42);
	});

	it("rejects fractional, suffixed, negative, and empty values", () => {
		expect(parseWholeStock("3.5")).toBeUndefined();
		expect(parseWholeStock("3abc")).toBeUndefined();
		expect(parseWholeStock("-1")).toBeUndefined();
		expect(parseWholeStock("")).toBeUndefined();
		expect(parseWholeStock("abc")).toBeUndefined();
	});
});
