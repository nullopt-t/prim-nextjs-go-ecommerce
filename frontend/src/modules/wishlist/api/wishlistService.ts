import { api } from "@/shared/api/client";
import type { ApiResponse, ApiPaginatedResponse, ApiMessageResponse, OffsetMeta } from "@/shared/api/types";
import type {
  WishlistCheck,
  WishlistCount,
  WishlistItem,
  WishlistProduct,
} from "@/modules/wishlist/types";

export type { WishlistProduct, WishlistItem, WishlistCheck, WishlistCount };

// Backwards-compatible aliases
export type WishlistResponse = ApiPaginatedResponse<WishlistItem, OffsetMeta>;
export type WishlistCheckResponse = ApiResponse<WishlistCheck>;
export type WishlistCountResponse = ApiResponse<WishlistCount>;

export const wishlistService = {
  getItems: (page = 1, pageSize = 50) =>
    api.get<ApiPaginatedResponse<WishlistItem, OffsetMeta>>(
      `/api/v1/wishlist?page=${page}&pageSize=${pageSize}`
    ),

  getCount: () =>
    api.get<ApiResponse<WishlistCount>>("/api/v1/wishlist/count"),

  checkInWishlist: (productId: string, variantId?: string) =>
    api.get<ApiResponse<WishlistCheck>>(
      `/api/v1/wishlist/check/${productId}${variantId ? `?variantId=${encodeURIComponent(variantId)}` : ""}`
    ),

  checkItem: (productId: string) =>
    api.get<ApiResponse<WishlistCheck>>(`/api/v1/wishlist/check?productId=${productId}`),

  addItem: (productId: string, variantId?: string) =>
    api.post<ApiResponse<WishlistItem>, { productId: string; variantId?: string }>("/api/v1/wishlist", {
      productId,
      ...(variantId ? { variantId } : {}),
    }),

  removeItem: (itemId: string) =>
    api.delete<ApiMessageResponse>(`/api/v1/wishlist/items/${itemId}`),

  removeByProductId: (productId: string, variantId?: string) =>
    api.delete<ApiMessageResponse>(
      `/api/v1/wishlist/products/${productId}${variantId ? `?variantId=${encodeURIComponent(variantId)}` : ""}`
    ),

  clearWishlist: (params?: { productId?: string; variantId?: string; itemId?: string }) => {
    const query = new URLSearchParams();
    if (params?.productId) query.set("productId", params.productId);
    if (params?.variantId) query.set("variantId", params.variantId);
    if (params?.itemId) query.set("itemId", params.itemId);
    const qs = query.toString();
    return api.delete<ApiMessageResponse>(`/api/v1/wishlist${qs ? `?${qs}` : ""}`);
  },
};
