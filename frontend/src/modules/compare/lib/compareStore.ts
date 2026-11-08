"use client";

import { MAX_COMPARE_ITEMS } from "../types";

const STORAGE_KEY = "prim:compare";

export const COMPARE_UPDATED_EVENT = "prim:compare-updated";

/** Appends a slug (no duplicates) and enforces the cap. Pure, testable. */
export function mergeCompareSlugs(
	existing: string[],
	slug: string,
	maxItems = MAX_COMPARE_ITEMS,
): string[] {
	if (!slug) return existing.slice(0, maxItems);
	const next = existing.includes(slug) ? existing : [...existing, slug];
	return next.slice(0, maxItems);
}

export function removeCompareSlug(existing: string[], slug: string): string[] {
	return existing.filter((item) => item !== slug);
}

function notifyCompareChanged() {
	if (typeof window === "undefined") return;
	window.dispatchEvent(new Event(COMPARE_UPDATED_EVENT));
}

export function getCompareSlugs(): string[] {
	if (typeof window === "undefined") return [];
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		const parsed = raw ? JSON.parse(raw) : [];
		if (!Array.isArray(parsed)) return [];
		return parsed
			.filter((value): value is string => typeof value === "string" && value.length > 0)
			.slice(0, MAX_COMPARE_ITEMS);
	} catch {
		return [];
	}
}

function persistCompareSlugs(slugs: string[]) {
	try {
		window.localStorage.setItem(
			STORAGE_KEY,
			JSON.stringify(slugs.slice(0, MAX_COMPARE_ITEMS)),
		);
	} catch {
		// Storage can be unavailable in private mode; compare degrades gracefully.
	}
	notifyCompareChanged();
}

export function addCompareSlug(slug?: string | null): string[] {
	if (typeof window === "undefined" || !slug) return getCompareSlugs();
	const next = mergeCompareSlugs(getCompareSlugs(), slug);
	persistCompareSlugs(next);
	return next;
}

export function removeCompareSlugStored(slug?: string | null): string[] {
	if (typeof window === "undefined" || !slug) return getCompareSlugs();
	const next = removeCompareSlug(getCompareSlugs(), slug);
	persistCompareSlugs(next);
	return next;
}

export function clearCompareSlugs(notify = true) {
	if (typeof window === "undefined") return;
	try {
		window.localStorage.removeItem(STORAGE_KEY);
	} catch {
		// Ignore unavailable storage.
	}
	if (notify) notifyCompareChanged();
}

export function isCompareFull(slugs: string[]): boolean {
	return slugs.length >= MAX_COMPARE_ITEMS;
}
