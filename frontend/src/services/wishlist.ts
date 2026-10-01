import { api } from "@/api/client";

export interface WishlistProduct {
  id: string;
  slug: string;
  title: string;
  description?: string;
  productType: string;
  brandName?: string;
  categoryName?: string;
  thumbnailUrl?: string;
  price?: string;
  extractedPrice?: number;
  originalPrice?: string;
  extractedOriginalPrice?: number;
  currency?: string;
  inStock: boolean;
  variantId?: string;
  variantTitle?: string;
  variantSku?: string;
}

export interface WishlistItem {
  id: string;
  productId: string;
  variantId?: string;
  userId: string;
  createdAt: string;
  product: WishlistProduct;
}

export interface WishlistResponse {
  data: WishlistItem[];
  meta: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
    hasPrevious: boolean;
    hasNext: boolean;
  };
}

export interface WishlistCheckResponse {
  data: {
    inWishlist: boolean;
    wishlistItemId: string | null;
  };
}

export interface WishlistCountResponse {
  data: {
    count: number;
  };
}

export const wishlistService = {
  getItems: (page = 1, pageSize = 50) =>
    api.get<WishlistResponse>(`/api/v1/wishlist?page=${page}&pageSize=${pageSize}`),

  getCount: () =>
    api.get<WishlistCountResponse>("/api/v1/wishlist/count"),

  checkInWishlist: (productId: string, variantId?: string) =>
    api.get<WishlistCheckResponse>(
      `/api/v1/wishlist/check/${productId}${variantId ? `?variantId=${encodeURIComponent(variantId)}` : ""}`
    ),

  addItem: (productId: string, variantId?: string) =>
    api.post<{ data: WishlistItem }>("/api/v1/wishlist", {
      productId,
      ...(variantId ? { variantId } : {}),
    }),

  removeItem: (itemId: string) =>
    api.delete<{ message: string }>(`/api/v1/wishlist/items/${itemId}`),

  removeByProductId: (productId: string, variantId?: string) =>
    api.delete<{ message: string }>(
      `/api/v1/wishlist/products/${productId}${variantId ? `?variantId=${encodeURIComponent(variantId)}` : ""}`
    ),

  clearWishlist: (params?: { productId?: string; variantId?: string; itemId?: string }) => {
    const query = new URLSearchParams();
    if (params?.productId) query.set("productId", params.productId);
    if (params?.variantId) query.set("variantId", params.variantId);
    if (params?.itemId) query.set("itemId", params.itemId);
    const qs = query.toString();
    return api.delete<{ message: string }>(`/api/v1/wishlist${qs ? `?${qs}` : ""}`);
  },
};
