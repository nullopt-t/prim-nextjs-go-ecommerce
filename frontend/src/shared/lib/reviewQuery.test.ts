import { describe, expect, it } from "vitest";
import { buildReviewQuery } from "@/shared/lib/reviewQuery";

describe("buildReviewQuery", () => {
	it("sends the page size the backend pages by", () => {
		expect(buildReviewQuery({ pageSize: 20 }).get("pageSize")).toBe("20");
	});

	it("sends the cursor the backend returned for the next page", () => {
		expect(buildReviewQuery({ cursor: "abc123" }).get("cursor")).toBe("abc123");
	});

	it("omits the cursor for the first page", () => {
		expect(buildReviewQuery({ cursor: null }).has("cursor")).toBe(false);
		expect(buildReviewQuery({}).has("cursor")).toBe(false);
	});

	it("sends the star filter the backend applies", () => {
		expect(buildReviewQuery({ rating: 4 }).get("rating")).toBe("4");
	});

	it("omits the filter when no rating is selected", () => {
		expect(buildReviewQuery({ rating: null }).has("rating")).toBe(false);
	});
});
