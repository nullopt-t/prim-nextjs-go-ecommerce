"use client";

const STORAGE_KEY = "prim_recent_searches";
const MAX_ITEMS = 5;

/** Most-recent-first, deduplicated snapshot. SSR-safe. */
export function getRecentSearches(): string[] {
	if (typeof window === "undefined") return [];
	try {
		const parsed: unknown = JSON.parse(window.localStorage.getItem(STORAGE_KEY) || "[]");
		return Array.isArray(parsed)
			? parsed.filter((item): item is string => typeof item === "string" && item.trim().length > 0).slice(0, MAX_ITEMS)
			: [];
	} catch {
		return [];
	}
}

export function addRecentSearch(query: string): void {
	const term = query.trim();
	if (typeof window === "undefined" || !term) return;
	try {
		const next = [term, ...getRecentSearches().filter((item) => item !== term)].slice(0, MAX_ITEMS);
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
	} catch {
		// Storage can be unavailable in private mode; recents are a convenience only.
	}
}

export function removeRecentSearch(query: string): void {
	if (typeof window === "undefined") return;
	try {
		window.localStorage.setItem(
			STORAGE_KEY,
			JSON.stringify(getRecentSearches().filter((item) => item !== query)),
		);
	} catch {
		// Ignore unavailable storage.
	}
}

export function clearRecentSearches(): void {
	if (typeof window === "undefined") return;
	try {
		window.localStorage.removeItem(STORAGE_KEY);
	} catch {
		// Ignore unavailable storage.
	}
}
