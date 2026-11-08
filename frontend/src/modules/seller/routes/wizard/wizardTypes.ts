// Shared wizard types and pure helpers, extracted from the product wizard so
// the step sections can be maintained independently.

export interface BundleItemDraft {
	productId: string;
	quantity: string;
}

export interface ProductForm {
	title: string;
	slug: string;
	brandId: string;
	categoryId: string;
	description: string;
	highlights: string;
	productType: "simple" | "variable" | "bundle";
	variationAxes: string;
	sku: string;
	price: string;
	crossesOutPrice: string;
	currency: string;
	initialStock: string;
	bundleItems: BundleItemDraft[];
}

export const EMPTY_FORM: ProductForm = {
	title: "",
	slug: "",
	brandId: "",
	categoryId: "",
	description: "",
	highlights: "",
	productType: "simple",
	variationAxes: "",
	sku: "",
	price: "",
	crossesOutPrice: "",
	currency: "EGP",
	initialStock: "0",
	bundleItems: [{ productId: "", quantity: "1" }],
};

export type WizardStep = 0 | 1 | 2 | 3;

export function slugify(value: string) {
	return value
		.trim()
		.toLowerCase()
		.normalize("NFKD")
		.replace(/[^\w\s-]/g, "")
		.replace(/[\s_-]+/g, "-")
		.replace(/^-+|-+$/g, "");
}

export function parseMajorPrice(value: string): number | undefined {
	const parsed = Number(value);
	if (value.trim() === "" || !Number.isFinite(parsed) || parsed <= 0) return undefined;
	return Math.round(parsed * 100);
}

export function draftKeyFor(productId?: string) {
	return productId ? `prim:seller-product-draft:${productId}` : "prim:seller-product-draft:new";
}

// One composer row while building a single combination: an attribute plus
// the one value it takes in that combination. Saved combinations are built
// one at a time instead of expanding every value into a cartesian product.
// Bundle payload rows: only a chosen product with a positive whole quantity is
// sendable, so validation and submission share one rule.
export function bundlePayloadItems(items: BundleItemDraft[]): { productId: string; quantity: number }[] {
	return items.flatMap((item) => {
		const quantity = Number(item.quantity);
		return item.productId && Number.isInteger(quantity) && quantity > 0
			? [{ productId: item.productId, quantity }]
			: [];
	});
}

export interface VariationRow {
	key: string;
	value: string;
}

export function uniqueKeys(combinations: Record<string, string>[]): string[] {
	return [...new Set(combinations.flatMap((combination) => Object.keys(combination)))];
}

export function formatCombination(combination: Record<string, string>): string {
	return Object.values(combination).join(", ");
}

export function combinationKey(combination: Record<string, string>): string {
	return Object.keys(combination)
		.sort()
		.map((key) => `${key}=${combination[key]}`)
		.join("|");
}

export interface VariantOverride {
	price?: string;
	crossesOutPrice?: string;
	sku?: string;
	stock?: string;
	thumbnailUrl?: string;
}
