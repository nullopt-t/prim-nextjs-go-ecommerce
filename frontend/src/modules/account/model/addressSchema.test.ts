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
});
