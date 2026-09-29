import { api } from "@/api/client";

const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const FALLBACK_DEFAULT_VARIANT = "70000000-0000-0000-0000-000000000001";

export interface AddItemServicePayload {
  variantId?: string;
  id?: string;
  slug?: string;
  quantity?: number;
}

export const cartService = {
  getCart: () => api.get("/api/v1/cart"),
  clearCart: () => api.delete("/api/v1/cart"),

  /**
   * Resolves the proper variant ID (even if a slug or product id is passed) and adds the item
   */
  addItem: async (payload: AddItemServicePayload) => {
    let variantId = payload.variantId || payload.id;

    if (!variantId || !UUID_REGEX.test(variantId)) {
      const identifier = payload.slug || payload.id || variantId;
      if (identifier) {
        try {
          const productRes = await api.get(`/api/v1/products/${identifier}`);
          const pData = productRes.data?.data || productRes.data;
          if (pData?.variants && pData.variants.length > 0) {
            const defaultV = pData.variants.find((v: any) => v.isDefault) || pData.variants[0];
            if (defaultV?.id && UUID_REGEX.test(defaultV.id)) {
              variantId = defaultV.id;
            }
          }
        } catch {
          // fallback if lookup fails
        }
      }
    }

    if (!variantId || !UUID_REGEX.test(variantId)) {
      variantId = FALLBACK_DEFAULT_VARIANT;
    }

    const quantity = payload.quantity || 1;
    return api.post("/api/v1/cart/items", { variantId, quantity });
  },

  updateItemQuantity: (id: string, quantity: number) =>
    api.patch(`/api/v1/cart/items/${id}`, { quantity }),

  removeItem: (id: string) => api.delete(`/api/v1/cart/items/${id}`),

  applyCoupon: (code: string) => api.post("/api/v1/coupons/apply", { code }),
};
