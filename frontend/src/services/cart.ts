import { api } from "@/api/client";

export const cartService = {
	getCart: () => api.get("/api/v1/cart"),
	clearCart: () => api.delete("/api/v1/cart"),
	addItem: (payload) => api.post("/api/v1/cart/items", payload), // payload: { productId, variantId, quantity }
	updateItemQuantity: (id, quantity) => api.patch(`/api/v1/cart/items/${id}`, { quantity }),
	removeItem: (id) => api.delete(`/api/v1/cart/items/${id}`),
};
