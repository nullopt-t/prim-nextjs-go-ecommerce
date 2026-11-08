import { describe, it, expect } from "vitest";
import { buildPromotionPayload } from "./promotionPayload";

const base = {
	code: "SUMMER25",
	type: "Percentage",
	value: "15",
	usageLimit: 200,
	startDate: "2026-08-01",
	endDate: "2026-12-31",
};

describe("buildPromotionPayload", () => {
	it("builds a percentage payload with minor-unit passthrough", () => {
		const p = buildPromotionPayload(base);
		expect(p).toMatchObject({
			code: "SUMMER25",
			discountType: "percentage",
			discountValue: 15,
			status: "active",
		});
		expect(p?.startsAt).toBe(new Date("2026-08-01").toISOString());
		expect(p?.expiresAt).toBe(new Date("2026-12-31").toISOString());
	});

	it("converts a fixed amount to minor units", () => {
		const p = buildPromotionPayload({ ...base, type: "Fixed Amount", value: "25" });
		expect(p?.discountType).toBe("fixed_amount");
		expect(p?.discountValue).toBe(2500);
	});

	it("rejects unknown types instead of silently making a fixed coupon", () => {
		expect(buildPromotionPayload({ ...base, type: "Free Shipping" })).toBeNull();
		expect(buildPromotionPayload({ ...base, type: "" })).toBeNull();
	});

	it("rejects empty code, non-numeric value, and inverted validity window", () => {
		expect(buildPromotionPayload({ ...base, code: "  " })).toBeNull();
		expect(buildPromotionPayload({ ...base, value: "abc" })).toBeNull();
		expect(
			buildPromotionPayload({ ...base, startDate: "2026-12-31", endDate: "2026-08-01" }),
		).toBeNull();
	});

	it("clamps percentage above 100 and omits empty dates", () => {
		const p = buildPromotionPayload({ ...base, type: "Percentage", value: "150", startDate: "", endDate: "" });
		expect(p?.discountValue).toBe(100);
		expect(p?.startsAt).toBeUndefined();
		expect(p?.expiresAt).toBeUndefined();
	});
});
