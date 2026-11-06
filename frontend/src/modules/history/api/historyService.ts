import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type { SyncHistoryResult, ViewedProduct } from "@/modules/history/types";

export const historyService = {
  list: (limit = 10) =>
    api.get<ApiResponse<ViewedProduct[]>>(
      `/api/v1/user/history/products?limit=${encodeURIComponent(String(limit))}`
    ),

  recordView: (productId: string) =>
    api.post<void>(`/api/v1/user/history/products/${encodeURIComponent(productId)}`, undefined, {
      keepalive: true,
      skipAuthRedirect: true,
    }),

  syncGuest: (slugs: string[]) =>
    api.post<ApiResponse<SyncHistoryResult>, { slugs: string[] }>(
      "/api/v1/user/history/products/sync",
      { slugs }
    ),

  remove: (productId: string) =>
    api.delete<void>(`/api/v1/user/history/products/${encodeURIComponent(productId)}`),

  clear: () => api.delete<void>("/api/v1/user/history/products"),
};
