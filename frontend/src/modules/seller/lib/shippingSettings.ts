export interface ShippingSettingsValues {
	shippingFee: number;
	freeShippingThreshold: number | null;
	shippingLeadTimeDays: number;
}

export function parseShippingSettings(
	feeText: string,
	thresholdText: string,
	leadDaysText: string,
): ShippingSettingsValues | null {
	const shippingFee = Number(feeText);
	const trimmedThreshold = thresholdText.trim();
	const freeShippingThreshold = trimmedThreshold === "" ? null : Number(trimmedThreshold);
	const shippingLeadTimeDays = Number(leadDaysText);

	if (
		!Number.isFinite(shippingFee) ||
		shippingFee < 0 ||
		(freeShippingThreshold !== null &&
			(!Number.isFinite(freeShippingThreshold) || freeShippingThreshold < 0)) ||
		!Number.isInteger(shippingLeadTimeDays) ||
		shippingLeadTimeDays < 1 ||
		shippingLeadTimeDays > 30
	) {
		return null;
	}

	return { shippingFee, freeShippingThreshold, shippingLeadTimeDays };
}
