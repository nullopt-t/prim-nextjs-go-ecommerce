import { api } from "@/api/client";

export interface WishlistItemPayload {
  id: string;
  productName?: { ar?: string; en?: string } | string;
  productPrice?: string | number;
  img?: string;
}

export const wishlistService = {
  getItems: () => api.get("/api/v1/wishlist"),
  addItem: (payload: WishlistItemPayload) => api.post("/api/v1/wishlist", payload),
  removeItem: (id: string | number) => api.delete(`/api/v1/wishlist/${id}`),
};
