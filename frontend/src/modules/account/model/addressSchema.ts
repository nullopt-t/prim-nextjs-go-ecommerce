import { z } from "zod";

export interface AddressFormValues {
	label: string;
	fullName: string;
	phone: string;
	line1: string;
	line2: string;
	city: string;
	state: string;
	postalCode: string;
	country: string;
	isDefault: boolean;
}

export const INITIAL_ADDRESS: AddressFormValues = {
	label: "Home",
	fullName: "",
	phone: "",
	line1: "",
	line2: "",
	city: "",
	state: "",
	postalCode: "",
	country: "Egypt",
	isDefault: false,
};

// Egyptian mobile numbers: +20, 0020, or 0 followed by 1[0125] and 8 digits. Spaces and dashes ignored.
export const EGYPTIAN_PHONE_REGEX = /^(?:\+20|0020|0)1[0125]\d{8}$/;

// Egyptian postal codes: exactly 5 digits.
export const POSTAL_CODE_REGEX = /^\d{5}$/;

// Normalizes Egyptian phone input to international format (+201xxxxxxxxx).
export function normalizeEgyptianPhone(raw: string): string {
	const stripped = raw.trim().replace(/[\s-]/g, "");
	if (!stripped) return "";
	if (stripped.startsWith("0020")) return "+20" + stripped.slice(4);
	if (stripped.startsWith("01")) return "+20" + stripped.slice(1);
	return stripped;
}

export function isValidEgyptianPhone(raw: string): boolean {
	const stripped = raw.trim().replace(/[\s-]/g, "");
	return stripped === "" || EGYPTIAN_PHONE_REGEX.test(stripped);
}

export function isValidPostalCode(raw: string): boolean {
	const trimmed = raw.trim();
	return trimmed === "" || POSTAL_CODE_REGEX.test(trimmed);
}

/**
 * Shared address validation for account and checkout forms.
 * Messages come from the `addresses.validation.*` namespace.
 */
export function createAddressSchema(t: (key: string) => string) {
	return z.object({
		label: z.string().trim(),
		fullName: z
			.string()
			.trim()
			.min(1, t("validation.fullName"))
			.max(120),
		phone: z
			.string()
			.trim()
			.refine((val) => isValidEgyptianPhone(val), {
				message: t("validation.phone"),
			})
			.transform((val) => normalizeEgyptianPhone(val)),
		line1: z.string().trim().min(1, t("validation.line1")),
		line2: z.string().trim(),
		city: z.string().trim().min(1, t("validation.city")),
		state: z.string().trim(),
		postalCode: z
			.string()
			.trim()
			.refine((val) => isValidPostalCode(val), {
				message: t("validation.postalCode"),
			}),
		country: z.string().trim().min(1, t("validation.country")),
		isDefault: z.boolean(),
	});
}
