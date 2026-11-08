export interface PromotionFormData {
	code: string;
	type: string;
	value: string;
	usageLimit: number;
	startDate: string;
	endDate: string;
}

export interface PromotionPayload {
	code: string;
	discountType: "percentage" | "fixed_amount";
	discountValue: number;
	currency: string;
	minCartAmount: number;
	usageLimit: number;
	status: "active";
	startsAt?: string;
	expiresAt?: string;
}

/**
 * Builds the create-promotion payload. The backend only knows percentage and
 * fixed_amount discount types, so anything else (e.g. a "free shipping" form
 * option) must be rejected here instead of silently becoming a fixed_amount
 * coupon. Returns null when the form is invalid.
 */
export function buildPromotionPayload(form: PromotionFormData): PromotionPayload | null {
	const code = form.code.toUpperCase().trim();
	if (!code) return null;

	const isPct = form.type === "Percentage";
	const isFixed = form.type === "Fixed Amount";
	if (!isPct && !isFixed) return null;

	const numVal = parseFloat(form.value.replace(/[^0-9.]/g, ""));
	if (!Number.isFinite(numVal)) return null;

	const discountValue = isPct
		? Math.min(100, Math.max(1, Math.round(numVal)))
		: Math.round(numVal * 100);
	if (discountValue <= 0) return null;

	const payload: PromotionPayload = {
		code,
		discountType: isPct ? "percentage" : "fixed_amount",
		discountValue,
		currency: "EGP",
		minCartAmount: 5000,
		usageLimit: Number(form.usageLimit) || 100,
		status: "active",
	};
	if (form.startDate) payload.startsAt = new Date(form.startDate).toISOString();
	if (form.endDate) payload.expiresAt = new Date(form.endDate).toISOString();
	if (payload.startsAt && payload.expiresAt && payload.expiresAt < payload.startsAt) return null;
	return payload;
}
