import { describe, expect, it } from "vitest";
import {
	resolveTargetLocale,
	shouldAddLocalePrefix,
	type SupportedLocale,
} from "./locale";

describe("resolveTargetLocale", () => {
	it("prefers the explicit locale cookie over the store default", () => {
		expect(resolveTargetLocale("ar", "en")).toBe("ar");
		expect(resolveTargetLocale("en", "ar")).toBe("en");
	});

	it("falls back to the store default when no explicit choice", () => {
		expect(resolveTargetLocale(undefined, "en")).toBe("en");
		expect(resolveTargetLocale(undefined, "ar")).toBe("ar");
	});

	it("returns undefined when neither is a supported locale", () => {
		expect(resolveTargetLocale(undefined, undefined)).toBeUndefined();
		expect(resolveTargetLocale("fr", "de")).toBeUndefined();
		// Unsupported explicit + unsupported default => no valid target.
		expect(resolveTargetLocale("fr", "de")).toBeUndefined();
	});

	it("falls back to the store default when the explicit choice is unsupported", () => {
		expect(resolveTargetLocale("fr", "en")).toBe("en");
		expect(resolveTargetLocale("pt", "en")).toBe("en");
		expect(resolveTargetLocale("pt", undefined)).toBeUndefined();
	});

	it("prefers the explicit locale cookie over the store default", () => {
		expect(resolveTargetLocale("ar", "en")).toBe("ar");
		expect(resolveTargetLocale("en", "ar")).toBe("en");
	});

	it("falls back to the store default when no explicit choice", () => {
		expect(resolveTargetLocale(undefined, "en")).toBe("en");
		expect(resolveTargetLocale(undefined, "ar")).toBe("ar");
	});
});

describe("shouldAddLocalePrefix", () => {
	const cases: Array<{
		locale: SupportedLocale;
		askedPath: string;
		target: SupportedLocale | undefined;
		expected: boolean;
	}> = [
		{ locale: "en", askedPath: "/", target: "ar", expected: true },
		{ locale: "en", askedPath: "/products", target: "ar", expected: true },
		{ locale: "ar", askedPath: "/", target: "en", expected: true },
		{ locale: "en", askedPath: "/", target: "en", expected: false },
		{ locale: "en", askedPath: "/en", target: "ar", expected: false },
		{ locale: "ar", askedPath: "/ar/products", target: "en", expected: false },
		{ locale: "en", askedPath: "/", target: undefined, expected: false },
		{ locale: "en", askedPath: "/about?q=1", target: "ar", expected: true },
	];

	it.each(cases)(
		"locale=$locale askedPath=$askedPath target=$target => $expected",
		({ locale, askedPath, target, expected }) => {
			expect(shouldAddLocalePrefix(locale, askedPath, target)).toBe(expected);
		},
	);
});
