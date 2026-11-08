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
	productType: "simple" | "variable" | "bundle" | "downloadable";
	variationAxes: string;
	sku: string;
	price: string;
	crossesOutPrice: string;
	currency: string;
	initialStock: string;
	bundleItems: BundleItemDraft[];
	downloadUrl: string;
	deliveryInstructions: string;
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
	downloadUrl: "",
	deliveryInstructions: "",
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
	const minor = Math.round(parsed * 100);
	// Sub-cent input like "0.001" rounds to zero; a zero price is not a price.
	if (minor <= 0) return undefined;
	return minor;
}

// Stock must be a whole, non-negative unit count. Number() (not parseInt)
// rejects "3.5" and "3abc", which parseInt would silently truncate.
export function parseWholeStock(value: string): number | undefined {
	if (value.trim() === "") return undefined;
	const parsed = Number(value);
	return Number.isInteger(parsed) && parsed >= 0 ? parsed : undefined;
}

export function draftKeyFor(productId?: string) {
	return productId ? `prim:seller-product-draft:${productId}` : "prim:seller-product-draft:new";
}

// One composer row while building a single combination: an attribute plus
// the one value it takes in that combination. Saved combinations are built
// one at a time instead of expanding every value into a cartesian product.
// Downloadable products deliver a hosted file, so the seller link must be a
// real http(s) URL rather than free text.
export function isValidDownloadUrl(value: string): boolean {
	const trimmed = value.trim();
	if (!trimmed) return false;
	try {
		const parsed = new URL(trimmed);
		return parsed.protocol === "http:" || parsed.protocol === "https:";
	} catch {
		return false;
	}
}

// Bundle payload rows: only a chosen product with a positive whole quantity is
// sendable, so validation and submission share one rule. An optional catalog
// or predicate can be passed to exclude deleted or zero-priced items.
export function bundlePayloadItems(
	items: BundleItemDraft[],
	isValidProduct?: (productId: string) => boolean,
): { productId: string; quantity: number }[] {
	// Merge repeated rows for the same product so one kit never bills the
	// same line twice (the backend stores one row per product per bundle).
	const merged = new Map<string, number>();
	for (const item of items) {
		const quantity = Number(item.quantity);
		const productValid = isValidProduct ? isValidProduct(item.productId) : true;
		if (item.productId && productValid && Number.isInteger(quantity) && quantity > 0) {
			merged.set(item.productId, (merged.get(item.productId) ?? 0) + quantity);
		}
	}
	return [...merged].map(([productId, quantity]) => ({ productId, quantity }));
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
