/**
 * Volume/tier pricing math for the PDP. Tiers arrive (when the backend
 * provides them) as quantity thresholds with either an explicit per-unit
 * price or a discount percent; the display layer derives the active tier
 * for the selected quantity and the per-unit price to render.
 */

export interface TierPricing {
	minQuantity: number;
	/** Explicit per-unit price at this tier. Takes precedence over percent. */
	unitPrice?: number;
	/** Discount percent off the base unit price at this tier. */
	discountPercent?: number;
}

export interface TierResolution {
	/** Per-unit price that applies at the selected quantity. */
	unitPrice: number;
	/** The highest tier whose threshold the quantity meets, if any. */
	activeTier: TierPricing | null;
	/** base - unitPrice when discounted, else 0. */
	savingsPerUnit: number;
}

/** Normalizes raw tiers: drops invalid rows, sorts ascending by threshold. */
export function normalizeTiers(tiers: unknown): TierPricing[] {
	if (!Array.isArray(tiers)) return [];
	const out: TierPricing[] = [];
	for (const tier of tiers) {
		if (!tier || typeof tier !== "object") continue;
		const record = tier as Record<string, unknown>;
		const minQuantity = Number(record.minQuantity);
		if (!Number.isFinite(minQuantity) || minQuantity < 1) continue;
		const unitPrice = record.unitPrice !== undefined ? Number(record.unitPrice) : undefined;
		const discountPercent =
			record.discountPercent !== undefined ? Number(record.discountPercent) : undefined;
		const hasUnitPrice = unitPrice !== undefined && Number.isFinite(unitPrice) && unitPrice > 0;
		const hasPercent =
			discountPercent !== undefined &&
			Number.isFinite(discountPercent) &&
			discountPercent > 0 &&
			discountPercent <= 100;
		if (!hasUnitPrice && !hasPercent) continue;
		out.push({
			minQuantity: Math.floor(minQuantity),
			...(hasUnitPrice ? { unitPrice } : {}),
			...(hasPercent ? { discountPercent } : {}),
		});
	}
	return out.sort((a, b) => a.minQuantity - b.minQuantity);
}

/**
 * Resolves the applicable tier for `quantity`. Display-only: the charged
 * price stays server-side; this keeps the PDP honest about thresholds
 * without changing checkout math.
 */
export function resolveTier(
	baseUnitPrice: number,
	tiers: TierPricing[] | undefined | null,
	quantity: number,
): TierResolution {
	const base = Number(baseUnitPrice) || 0;
	const list = tiers ?? [];
	// Scan all tiers (callers may pass unsorted lists): keep the highest
	// threshold the quantity meets instead of stopping at the first miss.
	let activeTier: TierPricing | null = null;
	for (const tier of list) {
		if (tier.minQuantity <= quantity && (!activeTier || tier.minQuantity >= activeTier.minQuantity)) {
			activeTier = tier;
		}
	}
	if (!activeTier) return { unitPrice: base, activeTier: null, savingsPerUnit: 0 };
	const unitPrice = Math.max(
		0,
		activeTier.unitPrice ?? base * (1 - (activeTier.discountPercent || 0) / 100),
	);
	const rounded = Math.round(unitPrice * 100) / 100;
	return { unitPrice: rounded, activeTier, savingsPerUnit: Math.max(0, base - rounded) };
}
