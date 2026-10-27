/**
 * Wishlist domain and response models matching backend wishlist handlers.
 */

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
  [key: string]: unknown;
}

export interface WishlistItem {
  id: string;
  productId: string;
  variantId?: string;
  userId: string;
  createdAt: string;
  product: WishlistProduct;
  [key: string]: unknown;
}

export interface WishlistCheck {
  inWishlist: boolean;
  wishlistItemId: string | null;
}

export interface WishlistCount {
  count: number;
}
