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
  /**
   * Guest (unauthenticated) order lookup by email + order number (UUID).
   * No session is required or sent by the backend contract; skipAuthRedirect
   * prevents the shared client from bouncing guests to /auth on a 401.
   */
  lookupOrder: (email: string, orderNumber: string) => {
    const query = new URLSearchParams({ email, orderNumber });
    return api.get<ApiResponse<Order>>(`/api/v1/orders/lookup?${query.toString()}`, {
      skipAuthRedirect: true,
    });
  },
  checkout: (payload: CheckoutPayload) =>
    api.post<ApiResponse<CheckoutResult>, CheckoutPayload>("/api/v1/checkout", payload),
  updateOrder: (id: string, payload: UpdateOrderPayload) =>
    api.patch<ApiResponse<Order>, UpdateOrderPayload>(`/api/v1/orders/${id}`, payload),
  cancelOrder: (id: string) => api.post<ApiResponse<Order>>(`/api/v1/orders/${id}/cancel`),
};
