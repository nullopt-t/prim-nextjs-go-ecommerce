import { describe, expect, it } from "vitest";
import { convertCurrency, isSupportedCurrency, normalizeCurrency } from "./currency";

describe("normalizeCurrency", () => {
	it("accepts supported codes case-insensitively", () => {
		expect(normalizeCurrency("usd")).toBe("USD");
		expect(normalizeCurrency(" EGP ")).toBe("EGP");
	});

	it("rejects unsupported codes", () => {
		expect(normalizeCurrency("GBP")).toBeNull();
		expect(normalizeCurrency("")).toBeNull();
		expect(normalizeCurrency(null)).toBeNull();
	});
});

describe("isSupportedCurrency", () => {
	it("only accepts the supported set", () => {
		expect(isSupportedCurrency("EUR")).toBe(true);
		expect(isSupportedCurrency("GBP")).toBe(false);
	});
});

describe("convertCurrency", () => {
	it("converts EGP → USD at ~48.5 and back without drift", () => {
		const usd = convertCurrency(970, "EGP", "USD");
		expect(usd).toBeCloseTo(20, 0);
		expect(convertCurrency(usd, "USD", "EGP")).toBeCloseTo(970, 5);
	});

	it("is identity for matching codes and passes through unknowns", () => {
		expect(convertCurrency(50, "USD", "USD")).toBe(50);
		expect(convertCurrency(50, "GBP", "USD")).toBe(50);
		expect(convertCurrency(50, "USD", null)).toBe(50);
		expect(convertCurrency(50, null, "USD")).toBe(50);
	});
});
