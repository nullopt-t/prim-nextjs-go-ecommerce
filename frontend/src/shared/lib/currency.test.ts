import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
	convertCurrency,
	convertCurrencyRounded,
	getStoredCurrency,
	isSupportedCurrency,
	normalizeCurrency,
	rememberCurrencyChoice,
} from "./currency";

const originalWindow = globalThis.window as unknown;
const originalDocument = globalThis.document as unknown;
type GlobalWindow = Window & typeof globalThis;

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

describe("convertCurrencyRounded", () => {
	it("rounds to cents like the display formatter", () => {
		expect(convertCurrencyRounded(970, "EGP", "USD")).toBe(20);
		expect(convertCurrencyRounded(999.99, "EGP", "USD")).toBe(20.62);
	});

	it("matches the displayed price at filter thresholds", () => {
		// 970.23765 EGP converts to 20.0049... USD, which displays as 20.00.
		// The raw float would fail a maxPrice=20 filter; the rounded value passes.
		const raw = convertCurrency(970.23765, "EGP", "USD");
		expect(raw).toBeGreaterThan(20);
		expect(convertCurrencyRounded(970.23765, "EGP", "USD")).toBe(20);
		expect(convertCurrencyRounded(970.23765, "EGP", "USD") <= 20).toBe(true);
	});
});

describe("currency persistence", () => {
	let cookieStore = "";
	let storedItem: string | null = null;

	function installBrowser(protocol: string) {
		cookieStore = "";
		storedItem = null;
		globalThis.document = {
			get cookie(): string {
				return cookieStore;
			},
			set cookie(value: string) {
				cookieStore = String(value);
			},
		} as unknown as Document;
		globalThis.window = {
			location: { protocol },
			localStorage: {
				getItem: (key: string) => (key === "prim:currency" ? storedItem : null),
				setItem: (key: string, value: string) => {
					if (key === "prim:currency") storedItem = value;
				},
				removeItem: () => {
					storedItem = null;
				},
			} as unknown as Storage,
		} as unknown as GlobalWindow;
	}

	beforeEach(() => installBrowser("https:"));

	afterEach(() => {
		globalThis.window = originalWindow as unknown as GlobalWindow;
		globalThis.document = originalDocument as unknown as Document;
	});

	it("writes the cookie and localStorage, then reads the cookie back", () => {
		rememberCurrencyChoice("USD");
		expect(cookieStore).toContain("prim_currency=USD");
		expect(storedItem).toBe("USD");
		expect(getStoredCurrency()).toBe("USD");
	});

	it("prefers the cookie over localStorage", () => {
		storedItem = "EGP";
		cookieStore = "prim_currency=USD";
		expect(getStoredCurrency()).toBe("USD");
	});

	it("falls back to localStorage when the cookie is absent", () => {
		cookieStore = "";
		storedItem = "EUR";
		expect(getStoredCurrency()).toBe("EUR");
	});

	it("ignores unsupported codes", () => {
		rememberCurrencyChoice("GBP");
		expect(cookieStore).toBe("");
		expect(storedItem).toBeNull();
		expect(getStoredCurrency()).toBeNull();
	});

	it("is a no-op and reads null outside the browser", () => {
		globalThis.window = undefined as unknown as GlobalWindow;
		expect(() => rememberCurrencyChoice("USD")).not.toThrow();
		expect(getStoredCurrency()).toBeNull();
		globalThis.window = originalWindow as unknown as GlobalWindow;
	});

	it("drops the Secure flag on plain HTTP", () => {
		installBrowser("http:");
		rememberCurrencyChoice("AED");
		expect(cookieStore).toContain("prim_currency=AED");
		expect(cookieStore).not.toContain("Secure");
	});
});
