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
}

export interface WishlistItem {
  id: string;
  productId: string;
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

  checkInWishlist: (productId: string) =>
    api.get<WishlistCheckResponse>(`/api/v1/wishlist/check/${productId}`),

  addItem: (productId: string) =>
    api.post<{ data: WishlistItem }>("/api/v1/wishlist", { productId }),

  removeItem: (itemId: string) =>
    api.delete<{ message: string }>(`/api/v1/wishlist/items/${itemId}`),

  removeByProductId: (productId: string) =>
    api.delete<{ message: string }>(`/api/v1/wishlist/products/${productId}`),

  clearWishlist: () =>
    api.delete<{ message: string }>("/api/v1/wishlist"),
};
