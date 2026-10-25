import { api } from "@/shared/api/client";
import type { ApiResponse, ApiPaginatedResponse } from "@/shared/api/types";
import type { CreatePromotionPayload, Promotion } from "@/modules/promotions/types";

export const promotionService = {
  getPromotions: () => api.get<ApiPaginatedResponse<Promotion> | ApiResponse<Promotion[]>>("/api/v1/promotions"),
  createPromotion: (payload: CreatePromotionPayload) =>
    api.post<ApiResponse<Promotion>, CreatePromotionPayload>("/api/v1/promotions", payload),
  deletePromotion: (id: string) => api.delete<void>(`/api/v1/promotions/${id}`),
};
