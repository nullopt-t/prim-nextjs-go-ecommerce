import { api } from "@/api/client";

export const promotionService = {
	getPromotions: () => api.get("/api/v1/promotions"),
	createPromotion: (payload) => api.post("/api/v1/promotions", payload),
	deletePromotion: (id) => api.delete(`/api/v1/promotions/${id}`),
};
