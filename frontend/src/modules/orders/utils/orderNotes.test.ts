import { describe, expect, it } from "vitest";
import { customerFacingNote } from "./orderNotes";

describe("customerFacingNote", () => {
	it("hides backend debug and status-echo notes", () => {
		expect(
			customerFacingNote(
				"payment confirmed via client intent=pi_3UcLqY2FuPaeAQGE0Eg5SMwf",
			),
		).toBeNull();
		expect(customerFacingNote("payment webhook payment_intent.succeeded tx=evt_123")).toBeNull();
		expect(customerFacingNote("Order created with status pending")).toBeNull();
		expect(customerFacingNote("Auto-cancelled: checkout reservation expired")).toBeNull();
		expect(customerFacingNote("Cancelled by customer")).toBeNull();
		expect(customerFacingNote("Tracking updated: 1Z999")).toBeNull();
		expect(customerFacingNote("Seller fulfilled 2 item(s) (tracking: 1Z, carrier: DHL)")).toBeNull();
		expect(customerFacingNote("provider_transaction_id: abc")).toBeNull();
	});

	it("keeps free-form admin/seller notes", () => {
		expect(customerFacingNote("  Please call before delivery  ")).toBe(
			"Please call before delivery",
		);
		expect(customerFacingNote("سيتم التوصيل بعد الخامسة مساءً")).toBe(
			"سيتم التوصيل بعد الخامسة مساءً",
		);
	});

	it("returns null for empty notes", () => {
		expect(customerFacingNote(undefined)).toBeNull();
		expect(customerFacingNote("   ")).toBeNull();
	});
});
