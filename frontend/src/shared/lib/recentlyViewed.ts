"use client";

const STORAGE_KEY = "prim:recently-viewed";
const MAX_ITEMS = 50;

export const RECENTLY_VIEWED_UPDATED_EVENT = "prim:recently-viewed-updated";

function notifyRecentlyViewedChanged() {
	if (typeof window === "undefined") return;
	window.dispatchEvent(new Event(RECENTLY_VIEWED_UPDATED_EVENT));
}

/** Moves a slug to the front, drops duplicates, and enforces the cap. */
export function mergeRecentSlugs(
	existing: string[],
	slug: string,
	maxItems = MAX_ITEMS
): string[] {
	const next = [slug, ...existing.filter((item) => item !== slug)];
	return next.slice(0, maxItems);
}

export function getRecentlyViewedSlugs(): string[] {
	if (typeof window === "undefined") return [];
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		const parsed = raw ? JSON.parse(raw) : [];
		return Array.isArray(parsed)
			? parsed.filter((value): value is string => typeof value === "string").slice(0, MAX_ITEMS)
			: [];
	} catch {
		return [];
	}
}

export function recordRecentlyViewed(slug?: string | null) {
	if (typeof window === "undefined" || !slug) return;
	try {
		const next = mergeRecentSlugs(getRecentlyViewedSlugs(), slug);
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
		notifyRecentlyViewedChanged();
	} catch {
		// Storage can be unavailable in private mode; the server history still works.
	}
}

export function clearRecentlyViewed(notify = true) {
	if (typeof window === "undefined") return;
	try {
		window.localStorage.removeItem(STORAGE_KEY);
	} catch {
		// Ignore unavailable storage.
	}
	if (notify) notifyRecentlyViewedChanged();
}
