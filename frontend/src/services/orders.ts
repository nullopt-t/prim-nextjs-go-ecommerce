import { api } from "@/api/client";

export const orderService = {
	getOrders: () => api.get("/api/v1/orders"),
	getOrderById: (id) => api.get(`/api/v1/orders/${id}`),
	checkout: (payload) => api.post("/api/v1/checkout", payload),
	updateOrder: (id, payload) => api.patch(`/api/v1/orders/${id}`, payload),
	cancelOrder: (id) => api.post(`/api/v1/orders/${id}/cancel`),
};
