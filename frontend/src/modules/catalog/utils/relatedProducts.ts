/**
 * Explicit recommendation ranking for PDP "Recommended with this item" and
 * the cart drawer's "You may also like" rows.
 *
 * No seller-configured relations exist in the payload yet, so ranking is a
 * category/tag/brand heuristic (high signal first); when the backend ships
 * seller relation assets, prefer them here without touching consumers.
 */

export interface RelatedCandidate {
	id: string | number;
	slug?: string;
	category?: string | { id?: string; name?: string };
	categoryId?: string;
	brand?: string | { id?: string; name?: string };
	brandId?: string;
	tags?: Array<{ id?: string; name?: string; slug?: string } | string>;
	[key: string]: unknown;
}

function categoryKey(p: RelatedCandidate): string {
	const category = typeof p.category === "string" ? p.category : p.category?.name;
	return String(p.categoryId || category || "").toLowerCase();
}

function brandKey(p: RelatedCandidate): string {
	const brand = typeof p.brand === "string" ? p.brand : p.brand?.name;
	return String(p.brandId || brand || "").toLowerCase();
}

function tagKeys(p: RelatedCandidate): Set<string> {
	const out = new Set<string>();
	for (const tag of p.tags ?? []) {
		const key = typeof tag === "string" ? tag : tag.slug || tag.name || tag.id || "";
		if (key) out.add(String(key).toLowerCase());
	}
	return out;
}

function score(
	target: RelatedCandidate,
	candidate: RelatedCandidate,
	cartCategories: Set<string>,
	cartBrands: Set<string>,
): number {
	let value = 0;
	const targetTags = tagKeys(target);
	const candidateTags = tagKeys(candidate);
	for (const tag of candidateTags) {
		if (targetTags.has(tag)) value += 3;
	}
	const targetCategory = categoryKey(target);
	if (targetCategory && categoryKey(candidate) === targetCategory) value += 2;
	const targetBrand = brandKey(target);
	if (targetBrand && brandKey(candidate) === targetBrand) value += 1;
	if (cartCategories.has(categoryKey(candidate))) value += 2;
	if (cartBrands.has(brandKey(candidate))) value += 1;
	return value;
}

/**
 * Ranks candidates by category/tag/brand affinity to `target`, excluding the
 * products in `excludeIds` (self + cart). Deterministic: equal scores keep
 * the catalog's stable order (Array sort is stable).
 */
export function selectRelatedProducts<T extends RelatedCandidate>(
	target: RelatedCandidate,
	candidates: T[],
	excludeIds: Array<string | number> = [],
	limit = 5,
): T[] {
	const excluded = new Set(excludeIds.map(String));
	const cartCategories = new Set<string>();
	const cartBrands = new Set<string>();
	return candidates
		.filter((candidate) => candidate && !excluded.has(String(candidate.id)))
		.map((candidate) => ({
			candidate,
			value: score(target, candidate, cartCategories, cartBrands),
		}))
		.filter(({ value }) => value > 0)
		.sort((a, b) => b.value - a.value)
		.slice(0, Math.max(0, limit))
		.map(({ candidate }) => candidate);
}

/**
 * Cart "You may also like": same-category/brand affinity to anything already
 * in the cart, excluding the cart itself.
 */
export function selectCartRecommendations<T extends RelatedCandidate>(
	cartItems: RelatedCandidate[],
	candidates: T[],
	limit = 6,
): T[] {
	const cartIds = new Set(cartItems.map((item) => String(item.id)));
	const cartCategories = new Set(cartItems.map(categoryKey).filter(Boolean));
	const cartBrands = new Set(cartItems.map(brandKey).filter(Boolean));
	return candidates
		.filter((candidate) => candidate && !cartIds.has(String(candidate.id)))
		.map((candidate) => {
			let value = 0;
			if (cartCategories.has(categoryKey(candidate))) value += 2;
			if (cartBrands.has(brandKey(candidate))) value += 1;
			return { candidate, value };
		})
		.filter(({ value }) => value > 0)
		.sort((a, b) => b.value - a.value)
		.slice(0, Math.max(0, limit))
		.map(({ candidate }) => candidate);
}
