import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type {
  CheckoutPayload,
  CheckoutResult,
  Order,
  UpdateOrderPayload,
} from "@/modules/orders/types";

export type { CheckoutPayload };

export const orderService = {
  getOrders: () => api.get<ApiResponse<Order[]>>("/api/v1/orders"),
  getOrderById: (id: string) => api.get<ApiResponse<Order>>(`/api/v1/orders/${id}`),
  checkout: (payload: CheckoutPayload) =>
    api.post<ApiResponse<CheckoutResult>, CheckoutPayload>("/api/v1/checkout", payload),
  updateOrder: (id: string, payload: UpdateOrderPayload) =>
    api.patch<ApiResponse<Order>, UpdateOrderPayload>(`/api/v1/orders/${id}`, payload),
  cancelOrder: (id: string) => api.post<ApiResponse<Order>>(`/api/v1/orders/${id}/cancel`),
};
