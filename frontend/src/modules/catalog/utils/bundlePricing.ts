/**
 * Bundle savings math. The kit payload carries no prices, so item prices are
 * resolved from each component product (default variant price, else product
 * price) and multiplied by the pinned quantity.
 */

export interface BundlePriceInput {
	slug: string;
	quantity: number;
	price: number;
}

export interface BundleSavings {
	itemsTotal: number;
	savings: number;
	percentOff: number;
}

export function computeBundleSavings(
	items: BundlePriceInput[],
	bundlePrice: number,
): BundleSavings | null {
	const itemsTotal = items.reduce((acc, item) => acc + (item.price || 0) * (item.quantity || 0), 0);
	if (itemsTotal <= 0 || bundlePrice <= 0) return null;
	const savings = itemsTotal - bundlePrice;
	if (savings <= 0) return null;
	return {
		itemsTotal,
		savings,
		percentOff: Math.round((savings / itemsTotal) * 100),
	};
}

export function resolveItemPrice(
	product: {
		price?: string | number;
		extractedPrice?: number;
		variants?: Array<{
			id?: string;
			sku?: string;
			extractedPrice?: number;
			price?: string | number;
			isDefault?: boolean;
		}>;
	},
	targetVariant?: string,
): number {
	const variants = product.variants ?? [];
	const matchedVariant = targetVariant
		? variants.find((v) => v.id === targetVariant || v.sku === targetVariant)
		: undefined;
	const variant =
		matchedVariant ?? variants.find((v) => v.isDefault) ?? variants[0];
	const raw = variant?.extractedPrice ?? variant?.price ?? product.extractedPrice ?? product.price;
	const num = Number(raw);
	return Number.isFinite(num) ? num : 0;
}
