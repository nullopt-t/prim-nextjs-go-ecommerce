"use client";

const STORAGE_KEY = "prim:recently-viewed";
const MAX_ITEMS = 10;

export function getRecentlyViewedSlugs(): string[] {
	if (typeof window === "undefined") return [];
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		const parsed = raw ? JSON.parse(raw) : [];
		return Array.isArray(parsed) ? parsed.filter((v) => typeof v === "string") : [];
	} catch {
		return [];
	}
}

export function recordRecentlyViewed(slug?: string | null) {
	if (typeof window === "undefined" || !slug) return;
	try {
		const existing = getRecentlyViewedSlugs().filter((item) => item !== slug);
		const next = [slug, ...existing].slice(0, MAX_ITEMS);
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
	} catch {
		// Storage can be unavailable in private mode; ignore.
	}
}
