import { describe, it, expect } from "vitest";
import { orderStatusOptions } from "./orderStatusOptions";

describe("orderStatusOptions", () => {
	it("offers the valid transitions for each active status", () => {
		expect(orderStatusOptions("pending")).toEqual(["pending", "processing", "cancelled"]);
		expect(orderStatusOptions("processing")).toEqual(["processing", "shipped", "cancelled"]);
		expect(orderStatusOptions("shipped")).toEqual(["shipped", "delivered", "cancelled"]);
	});

	it("always includes the current status so the select never renders blank", () => {
		for (const status of ["paid", "refunded", "delivered", "cancelled", ""]) {
			const opts = orderStatusOptions(status);
			expect(opts[0]).toBe(status.toLowerCase() || "pending");
		}
		expect(orderStatusOptions("paid")).toEqual(["paid"]);
		expect(orderStatusOptions("refunded")).toEqual(["refunded"]);
	});

	it("normalizes uppercase status input", () => {
		expect(orderStatusOptions("PENDING")).toEqual(["pending", "processing", "cancelled"]);
	});
});
