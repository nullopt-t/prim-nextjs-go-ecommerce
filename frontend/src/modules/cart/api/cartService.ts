import { api } from "@/shared/api/client";
import type { ApiResponse } from "@/shared/api/types";
import type {
  AddItemPayload,
  Cart,
  CartItem,
  CouponApplication,
} from "@/modules/cart/types";
import type { ProductDetails } from "@/modules/catalog/types";

const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export type AddItemServicePayload = AddItemPayload;

export const cartService = {
  getCart: () => api.get<ApiResponse<Cart>>("/api/v1/cart"),
  clearCart: () => api.delete<void>("/api/v1/cart"),

  /**
   * Resolves the proper variant ID (even if a slug or product id is passed) and adds the item
   */
  addItem: async (payload: AddItemPayload): Promise<ApiResponse<CartItem>> => {
    let variantId = payload.variantId;

    // If no variantId provided, or if the provided variantId might be a product identifier, resolve it
    if (!variantId || !UUID_REGEX.test(variantId)) {
      const identifier = payload.slug || payload.id || variantId;
      if (identifier) {
        try {
          const productRes = await api.get<ApiResponse<ProductDetails>>(`/api/v1/products/${identifier}`);
          const pData = productRes?.data;
          if (pData?.variants && pData.variants.length > 0) {
            const defaultV = pData.variants.find((v) => v.isDefault) || pData.variants[0];
            if (defaultV?.id && UUID_REGEX.test(defaultV.id)) {
              variantId = defaultV.id;
            }
          }
        } catch {
          // fallback if lookup fails
        }
      }
    }

    // If still missing, check if payload.id is valid
    if (!variantId && payload.id && UUID_REGEX.test(payload.id)) {
      variantId = payload.id;
    }

    if (!variantId || !UUID_REGEX.test(variantId)) {
      // Never invent a variant id: posting a made-up UUID would attach the
      // wrong product (or fail server-side with a confusing error).
      throw new Error(
        "This product has no purchasable variant yet. Please choose an option."
      );
    }

    const quantity = payload.quantity || 1;
    return api.post<ApiResponse<CartItem>, { variantId: string; quantity: number }>("/api/v1/cart/items", {
      variantId,
      quantity,
    });
  },

  updateItemQuantity: (id: string, quantity: number) =>
    api.patch<ApiResponse<CartItem>, { quantity: number }>(`/api/v1/cart/items/${id}`, { quantity }),

  removeItem: (id: string) => api.delete<void>(`/api/v1/cart/items/${id}`),

  applyCoupon: (code: string, cartSubtotal?: number) =>
    api.post<ApiResponse<CouponApplication>, { code: string; cartSubtotal?: number }>("/api/v1/coupons/apply", {
      code,
      cartSubtotal: cartSubtotal !== undefined ? Math.round(cartSubtotal * 100) : undefined,
    }),
};
