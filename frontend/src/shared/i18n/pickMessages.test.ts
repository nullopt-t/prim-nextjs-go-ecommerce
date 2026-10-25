import { describe, expect, it } from "vitest";
import { pickMessages, type MessageMap } from "./pickMessages";

const source: MessageMap = {
	common: { header: { cart: "Cart" }, footer: "Footer" },
	home: { product: { addToCart: "Add" }, hero: "Hero" },
	admin: { roles: { admin: "Admin" }, products: { title: "Products" } },
	checkout: "Checkout",
};

describe("pickMessages", () => {
	it("selects top-level and nested paths", () => {
		const picked = pickMessages(source, ["common", "home.product", "admin.roles"]);

		expect(picked).toEqual({
			common: { header: { cart: "Cart" }, footer: "Footer" },
			home: { product: { addToCart: "Add" } },
			admin: { roles: { admin: "Admin" } },
		});
	});

	it("skips missing paths", () => {
		const picked = pickMessages(source, ["missing", "admin.missing"]);

		expect(picked).toEqual({});
	});

	it("does not mutate the source tree", () => {
		pickMessages(source, ["home.product"]);

		expect(source.home).toEqual({
			product: { addToCart: "Add" },
			hero: "Hero",
		});
	});
});
