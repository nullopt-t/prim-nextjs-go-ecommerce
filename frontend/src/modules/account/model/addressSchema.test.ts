import { describe, expect, it } from "vitest";
import {
	createAddressSchema,
	INITIAL_ADDRESS,
} from "@/modules/account/model/addressSchema";

const schema = createAddressSchema((key) => key);

describe("createAddressSchema", () => {
	it("accepts a complete address", () => {
		const result = schema.safeParse({
			...INITIAL_ADDRESS,
			fullName: "Ada Lovelace",
			line1: "12 Nile Street",
			city: "Cairo",
			country: "Egypt",
		});

		expect(result.success).toBe(true);
	});

	it("rejects missing required fields", () => {
		const result = schema.safeParse({
			...INITIAL_ADDRESS,
			fullName: " ",
			line1: "",
			city: "",
			country: "",
		});

		expect(result.success).toBe(false);
		if (result.success) return;
		expect(result.error.issues.map((issue) => issue.path.join("."))).toEqual(
			expect.arrayContaining(["fullName", "line1", "city", "country"]),
		);
	});

	it("trims required values", () => {
		const result = schema.parse({
			...INITIAL_ADDRESS,
			fullName: "  Ada  ",
			line1: "  12 Nile Street ",
			city: " Cairo ",
			country: " Egypt ",
		});

		expect(result.fullName).toBe("Ada");
		expect(result.line1).toBe("12 Nile Street");
		expect(result.city).toBe("Cairo");
		expect(result.country).toBe("Egypt");
	});

	it("validates Egyptian phone numbers", () => {
		const validPhones = ["", "01012345678", "+201012345678", "00201112345678", "+20 12 3456 7890", "015-1234-5678"];
		for (const phone of validPhones) {
			const res = schema.safeParse({
				...INITIAL_ADDRESS,
				fullName: "User",
				line1: "Street",
				city: "Cairo",
				country: "Egypt",
				phone,
			});
			expect(res.success).toBe(true);
		}

		const invalidPhones = ["abc", "+441234567890", "011234", "01312345678", "+2012345678901"];
		for (const phone of invalidPhones) {
			const res = schema.safeParse({
				...INITIAL_ADDRESS,
				fullName: "User",
				line1: "Street",
				city: "Cairo",
				country: "Egypt",
				phone,
			});
			expect(res.success).toBe(false);
		}
	});

	it("validates 5-digit postal codes", () => {
		const validCodes = ["", "  ", "11511", "12345"];
		for (const postalCode of validCodes) {
			const res = schema.safeParse({
				...INITIAL_ADDRESS,
				fullName: "User",
				line1: "Street",
				city: "Cairo",
				country: "Egypt",
				postalCode,
			});
			expect(res.success).toBe(true);
		}

		const invalidCodes = ["123", "123456", "12A45", "abcde"];
		for (const postalCode of invalidCodes) {
			const res = schema.safeParse({
				...INITIAL_ADDRESS,
				fullName: "User",
				line1: "Street",
				city: "Cairo",
				country: "Egypt",
				postalCode,
			});
			expect(res.success).toBe(false);
		}
	});
});
