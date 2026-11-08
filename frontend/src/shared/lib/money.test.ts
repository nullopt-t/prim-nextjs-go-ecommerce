import { describe, expect, it } from "vitest";
import { formatMoney, resolveCurrency } from "./money";

describe("formatMoney", () => {
	it("renders a dash for missing or invalid values", () => {
		expect(formatMoney(undefined)).toBe("—");
		expect(formatMoney(null)).toBe("—");
		expect(formatMoney(Number.NaN)).toBe("—");
	});

	it("formats with the currency code when provided", () => {
		expect(formatMoney(10.5, "EGP")).toBe("EGP 10.50");
		expect(formatMoney(1000, "USD")).toBe("USD 1000.00");
	});

	it("trims whitespace around the currency code", () => {
		expect(formatMoney(10.5, " EGP ")).toBe("EGP 10.50");
	});

	it("renders the bare amount when the currency is unknown", () => {
		expect(formatMoney(10.5)).toBe("10.50");
		expect(formatMoney(10.5, null)).toBe("10.50");
		expect(formatMoney(10.5, undefined)).toBe("10.50");
	});

	it("rounds to two decimals", () => {
		expect(formatMoney(10.556, "EGP")).toBe("EGP 10.56");
		expect(formatMoney(10.554, "EGP")).toBe("EGP 10.55");
	});
});

describe("resolveCurrency", () => {
	it("trims the code", () => {
		expect(resolveCurrency(" EGP ")).toBe("EGP");
	});

	it("returns an empty string for missing input", () => {
		expect(resolveCurrency(null)).toBe("");
		expect(resolveCurrency(undefined)).toBe("");
		expect(resolveCurrency("")).toBe("");
	});
});
