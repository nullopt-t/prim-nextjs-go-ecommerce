import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { addRecentSearch, clearRecentSearches, getRecentSearches, removeRecentSearch } from "./recentSearches";

const store = new Map<string, string>();

beforeEach(() => {
	store.clear();
	vi.stubGlobal("window", {
		localStorage: {
			getItem: (key: string) => store.get(key) ?? null,
			setItem: (key: string, value: string) => void store.set(key, value),
			removeItem: (key: string) => void store.delete(key),
		},
	});
});

afterEach(() => vi.unstubAllGlobals());

describe("recent searches", () => {
	it("keeps five unique terms in LIFO order", () => {
		for (const term of ["a", "b", "c", "a", "d", "e", "f"]) addRecentSearch(term);
		expect(getRecentSearches()).toEqual(["f", "e", "d", "a", "c"]);
	});

	it("ignores blank queries", () => {
		addRecentSearch("   ");
		expect(getRecentSearches()).toEqual([]);
	});

	it("removes and clears terms", () => {
		addRecentSearch("one");
		addRecentSearch("two");
		removeRecentSearch("one");
		expect(getRecentSearches()).toEqual(["two"]);
		clearRecentSearches();
		expect(getRecentSearches()).toEqual([]);
	});

	it("is SSR-safe without window", () => {
		vi.stubGlobal("window", undefined);
		expect(getRecentSearches()).toEqual([]);
		expect(() => addRecentSearch("test")).not.toThrow();
		expect(() => clearRecentSearches()).not.toThrow();
	});
});
