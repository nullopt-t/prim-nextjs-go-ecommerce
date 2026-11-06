import { describe, expect, it } from "vitest";
import { mergeRecentSlugs } from "./recentlyViewed";

describe("mergeRecentSlugs", () => {
	it("moves an existing slug to the front without duplicating it", () => {
		expect(mergeRecentSlugs(["b", "a", "c"], "a")).toEqual(["a", "b", "c"]);
	});

	it("adds a new slug at the front", () => {
		expect(mergeRecentSlugs(["b", "a"], "c")).toEqual(["c", "b", "a"]);
	});

	it("enforces the rolling cap", () => {
		const existing = ["2", "3", "4"];
		expect(mergeRecentSlugs(existing, "1", 3)).toEqual(["1", "2", "3"]);
	});
});
