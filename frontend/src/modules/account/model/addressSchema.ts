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
		label: z.string(),
		fullName: z.string().trim().min(1, t("validation.fullName")),
		phone: z
			.string()
			.refine((val) => isValidEgyptianPhone(val), {
				message: t("validation.phone"),
			}),
		line1: z.string().trim().min(1, t("validation.line1")),
		line2: z.string(),
		city: z.string().trim().min(1, t("validation.city")),
		state: z.string(),
		postalCode: z
			.string()
			.refine((val) => isValidPostalCode(val), {
				message: t("validation.postalCode"),
			}),
		country: z.string().trim().min(1, t("validation.country")),
		isDefault: z.boolean(),
	});
}
