import { describe, expect, it } from "vitest";
import { normalizeProduct } from "@/modules/catalog/model/normalizeProduct";

describe("normalizeProduct", () => {
	it("normalizes a backend product payload", () => {
		const product = normalizeProduct({
			id: "prod-1",
			slug: "classic-shoe",
			title: "Classic Shoe",
			price: "120.00",
			originalPrice: "150.00",
			thumbnail: "https://cdn.example.com/shoe.jpg",
			category: { id: "cat-1", name: "Shoes" },
			brand: { id: "brand-1", name: "Acme" },
			variationAxes: ["color"],
			variants: [
				{
					id: "var-1",
					attributes: { color: "Red" },
					price: "120.00",
					stock: { isInStock: true, availableQuantity: 7 },
					thumbnail: "https://cdn.example.com/red.jpg",
				},
				{
					id: "var-2",
					attributes: { color: "Blue" },
					price: "120.00",
					stock: { isInStock: false, availableQuantity: 0 },
				},
			],
		});

		expect(product).not.toBeNull();
		expect(product?.id).toBe("prod-1");
		expect(product?.slug).toBe("classic-shoe");
		expect(product?.title).toBe("Classic Shoe");
		expect(product?.price).toBe(120);
		expect(product?.oldPrice).toBe(150);
		expect(product?.discountPercentage).toBe("20%");
		expect(product?.category).toBe("Shoes");
		expect(product?.categoryId).toBe("cat-1");
		expect(product?.brand).toBe("Acme");
		expect(product?.inStock).toBe(true);
		expect(product?.stockCount).toBe(7);
		expect(product?.variationAxes).toEqual([
			{ key: "color", label: "Color", values: ["Red", "Blue"] },
		]);
		expect(
			product?.colors.map((color: { name: string }) => color.name),
		).toEqual(["Red", "Blue"]);
		expect(product?.images).toEqual([
			"https://cdn.example.com/red.jpg",
			"https://cdn.example.com/shoe.jpg",
		]);
	});

	it("unwraps wishlist items", () => {
		const product = normalizeProduct({
			id: "wish-1",
			product: {
				id: "prod-2",
				slug: "wishlisted-product",
				title: "Wishlisted Product",
				price: 50,
			},
		});

		expect(product?.id).toBe("prod-2");
		expect(product?.wishlistItemId).toBe("wish-1");
		expect(product?.title).toBe("Wishlisted Product");
		expect(product?.price).toBe(50);
	});

	it("returns null for an empty payload", () => {
		expect(normalizeProduct(null)).toBeNull();
	});
});
