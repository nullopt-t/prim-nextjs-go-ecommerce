/**
 * Compare domain models. The compare list stores product slugs only; product
 * details are fetched on demand via the existing catalog API.
 */

export type CompareSlug = string;

export interface CompareEntry {
	slug: string;
}

export interface CompareRow {
	key: string;
	label: string;
	values: string[];
}

export const MAX_COMPARE_ITEMS = 4;
