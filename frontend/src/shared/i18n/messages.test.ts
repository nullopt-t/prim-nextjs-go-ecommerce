import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

function load(locale: string): Record<string, unknown> {
	return JSON.parse(
		readFileSync(new URL(`../../../messages/${locale}.json`, import.meta.url), "utf8"),
	);
}

function flattenKeys(value: unknown, prefix = ""): string[] {
	const keys: string[] = [];
	if (!value || typeof value !== "object" || Array.isArray(value)) return keys;
	for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
		const full = prefix ? `${prefix}.${k}` : k;
		if (v && typeof v === "object" && !Array.isArray(v)) keys.push(...flattenKeys(v, full));
		else keys.push(full);
	}
	return keys;
}

function get(obj: Record<string, unknown>, dotted: string): unknown {
	let cur: unknown = obj;
	for (const part of dotted.split(".")) {
		if (!cur || typeof cur !== "object") return undefined;
		cur = (cur as Record<string, unknown>)[part];
	}
	return cur;
}

const en = load("en");
const ar = load("ar");
const enKeys = flattenKeys(en);
const arKeys = flattenKeys(ar);

describe("i18n message parity", () => {
	it("has identical message key sets in EN and AR", () => {
		expect(enKeys.length).toBeGreaterThan(0);
		expect([...enKeys].sort()).toEqual([...arKeys].sort());
	});

	it("covers every key used by shared shell components (header/footer/switcher)", () => {
		const used = [
			"common.currency.label",
			"common.currency.approximate",
			"common.currency.EGP",
			"common.currency.USD",
			"common.currency.EUR",
			"common.currency.SAR",
			"common.currency.AED",
			"common.footer.shop",
			"common.footer.support",
			"common.footer.company",
			"common.footer.allProducts",
			"common.footer.categories",
			"common.footer.wishlist",
			"common.footer.contactUs",
			"common.footer.faqs",
			"common.footer.shipping",
			"common.footer.returns",
			"common.footer.aboutUs",
			"common.footer.trackOrders",
			"common.footer.sellOnPrim",
			"common.footer.sellerPortal",
			"common.footer.careers",
			"common.footer.press",
			"common.footer.defaultStore",
			"common.footer.footerDescription",
			"common.footer.copyright",
		];
		for (const key of used) {
			expect(get(en, key), `EN missing ${key}`).toBeDefined();
			expect(get(ar, key), `AR missing ${key}`).toBeDefined();
			expect(typeof get(en, key)).toBe("string");
			expect(typeof get(ar, key)).toBe("string");
		}
	});
});
