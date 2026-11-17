import { describe, expect, it } from "vitest";
import { parseShippingSettings } from "./shippingSettings";

describe("parseShippingSettings", () => {
	it("parses valid values and treats an empty threshold as no free shipping", () => {
		expect(parseShippingSettings("5.50", "", "3")).toEqual({
			shippingFee: 5.5,
			freeShippingThreshold: null,
			shippingLeadTimeDays: 3,
		});
		expect(parseShippingSettings("0", "100", "1")?.freeShippingThreshold).toBe(100);
	});

	it("rejects negative or non-numeric fees and thresholds", () => {
		expect(parseShippingSettings("-1", "", "3")).toBeNull();
		expect(parseShippingSettings("abc", "", "3")).toBeNull();
		expect(parseShippingSettings("5", "-2", "3")).toBeNull();
	});

	it("accepts only whole lead times from 1 to 30 days", () => {
		expect(parseShippingSettings("0", "", "1")).not.toBeNull();
		expect(parseShippingSettings("0", "", "30")).not.toBeNull();
		expect(parseShippingSettings("0", "", "0")).toBeNull();
		expect(parseShippingSettings("0", "", "31")).toBeNull();
		expect(parseShippingSettings("0", "", "2.5")).toBeNull();
	});
});
