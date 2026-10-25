import { api } from "@/shared/api/client";

export interface CheckoutPayload {
	shippingAddressId: string;
	couponId?: string;
	paymentMethod: string;
}

export const orderService = {
	getOrders: () => api.get("/api/v1/orders"),
	getOrderById: (id: string) => api.get(`/api/v1/orders/${id}`),
	checkout: (payload: CheckoutPayload) => api.post("/api/v1/checkout", payload),
	updateOrder: (id: string, payload: Record<string, unknown>) =>
		api.patch(`/api/v1/orders/${id}`, payload),
	cancelOrder: (id: string) => api.post(`/api/v1/orders/${id}/cancel`),
};
