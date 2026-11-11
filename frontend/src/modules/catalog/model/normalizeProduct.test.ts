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

	it("keeps an explicit zero variant price instead of falling back to the product price", () => {
		const product = normalizeProduct({
			id: "prod-3",
			slug: "free-sample",
			title: "Free Sample",
			price: 99,
			variationAxes: ["color"],
			variants: [{ id: "var-free", price: 0, attributes: { color: "Red" } }],
		});
		expect(product?.variants[0]?.price).toBe(0);
		expect(product?.variants[0]?.extractedPrice).toBe(0);
	});

	it("falls back to the product price when variant price fields are null", () => {
		const product = normalizeProduct({
			id: "prod-4",
			slug: "null-price",
			title: "Null Price",
			price: 42,
			variants: [{ id: "var-null", price: null, extractedPrice: null }],
		});
		expect(product?.variants[0]?.price).toBe(42);
	});

	it("returns null for an empty payload", () => {
		expect(normalizeProduct(null)).toBeNull();
	});

	it("keeps the backend rating aggregate, its distribution and the tags", () => {
		const product = normalizeProduct({
			id: "prod-5",
			slug: "rated-product",
			title: "Rated Product",
			price: 100,
			rating: {
				averageRating: 4.5,
				reviewCount: 24,
				distribution: { "1": 1, "2": 0, "3": 2, "4": 8, "5": 13 },
			},
			tags: [
				{ id: "tag-1", name: "Featured" },
				{ id: "tag-2", name: "New Arrival" },
			],
		});

		expect(product?.rating?.averageRating).toBe(4.5);
		expect(product?.rating?.reviewCount).toBe(24);
		expect(product?.rating?.distribution?.["5"]).toBe(13);
		expect(product?.stars).toBe(4.5);
		expect(product?.reviews).toBe(24);
		expect(product?.tags).toEqual([
			{ id: "tag-1", name: "Featured" },
			{ id: "tag-2", name: "New Arrival" },
		]);
	});

	it("stays in stock when any variant is purchasable", () => {
		const product = normalizeProduct({
			id: "prod-6",
			slug: "multi-variant",
			title: "Multi Variant",
			price: 100,
			variants: [
				{ id: "v1", stock: { isInStock: false, availableQuantity: 0 } },
				{ id: "v2", isDefault: true, stock: { isInStock: true, availableQuantity: 4 } },
			],
		});

		expect(product?.inStock).toBe(true);
		expect(product?.stockCount).toBe(4);
	});

	it("is out of stock when no variant is purchasable", () => {
		const product = normalizeProduct({
			id: "prod-7",
			slug: "sold-out",
			title: "Sold Out",
			price: 100,
			variants: [{ id: "v1", stock: { isInStock: false, availableQuantity: 0 } }],
		});

		expect(product?.inStock).toBe(false);
	});
});
