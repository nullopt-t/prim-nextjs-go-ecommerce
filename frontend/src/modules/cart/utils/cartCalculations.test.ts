import { describe, it, expect } from "vitest";
import { calculateCartTotals } from "./cartCalculations";
import type { CartContextItem } from "@/modules/cart/model/cartContext";

const item = (unitPrice: number, quantity: number): CartContextItem => ({
	id: "1",
	productName: "Test",
	productPrice: "EGP 9999.99",
	unitPrice,
	quantity,
});

describe("calculateCartTotals", () => {
	it("uses exact unitPrice instead of parsing the formatted price string", () => {
		const r = calculateCartTotals([item(12.99, 3)], null, 150, 9.99);
		expect(r.subtotal).toBe(38.97);
		expect(r.total).toBe(48.96);
	});

	it("caps a fixed discount at the subtotal (backend parity)", () => {
		const r = calculateCartTotals(
			[item(10, 1)],
			{ code: "FIXED", discountType: "fixed_amount", fixedDiscount: 999 },
			150,
			9.99,
		);
		expect(r.discount).toBe(10);
		expect(r.total).toBe(9.99);
	});

	it("rounds float drift to cents", () => {
		const r = calculateCartTotals([item(0.1, 3)], null, 150, 0.2);
		expect(r.subtotal).toBe(0.3);
		expect(r.total).toBe(0.5);
	});
});
