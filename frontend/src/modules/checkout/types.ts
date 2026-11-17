export interface SellerShippingBreakdown {
	sellerId?: string;
	sellerName: string;
	subtotal: number;
	shippingFee: number;
	isFree: boolean;
	leadTimeDays: number;
}

export interface CheckoutSummary {
	shippingBreakdown?: SellerShippingBreakdown[];
}

export type CheckoutPreview = CheckoutSummary;
