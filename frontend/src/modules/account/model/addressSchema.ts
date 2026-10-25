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

/**
 * Shared address validation for account and checkout forms.
 * Messages come from the `addresses.validation.*` namespace.
 */
export function createAddressSchema(t: (key: string) => string) {
	return z.object({
		label: z.string(),
		fullName: z.string().trim().min(1, t("validation.fullName")),
		phone: z.string(),
		line1: z.string().trim().min(1, t("validation.line1")),
		line2: z.string(),
		city: z.string().trim().min(1, t("validation.city")),
		state: z.string(),
		postalCode: z.string(),
		country: z.string().trim().min(1, t("validation.country")),
		isDefault: z.boolean(),
	});
}
