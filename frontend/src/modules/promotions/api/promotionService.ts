import { api } from "@/shared/api/client";

export const promotionService = {
	getPromotions: () => api.get("/api/v1/promotions"),
	createPromotion: (payload: Record<string, unknown>) =>
		api.post("/api/v1/promotions", payload),
	deletePromotion: (id: string) => api.delete(`/api/v1/promotions/${id}`),
};
