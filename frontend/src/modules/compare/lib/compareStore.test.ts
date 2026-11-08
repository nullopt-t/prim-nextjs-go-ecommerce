import { describe, expect, it } from "vitest";
import { mergeCompareSlugs, removeCompareSlug } from "./compareStore";

describe("mergeCompareSlugs", () => {
	it("appends a new slug", () => {
		expect(mergeCompareSlugs(["a"], "b")).toEqual(["a", "b"]);
	});

	it("ignores a duplicate slug", () => {
		expect(mergeCompareSlugs(["a", "b"], "a")).toEqual(["a", "b"]);
	});

	it("enforces the cap", () => {
		expect(mergeCompareSlugs(["a", "b", "c", "d"], "e")).toEqual([
			"a",
			"b",
			"c",
			"d",
		]);
	});
});

describe("removeCompareSlug", () => {
	it("removes the slug", () => {
		expect(removeCompareSlug(["a", "b", "c"], "b")).toEqual(["a", "c"]);
	});
});
