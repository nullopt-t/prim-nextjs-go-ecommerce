/**
 * Catalog domain and response models matching Go backend catalog handlers.
 */

export interface Category {
  id: string;
  name: string;
}

export interface AdminCategory {
  id: string;
  parentId?: string;
  parentName?: string;
  name: string;
  productCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface Brand {
  id: string;
  name: string;
  link?: string;
  logoUrl?: string;
  productCount?: number;
}

export interface Tag {
  id: string;
  name: string;
  slug: string;
}

export interface ProductVariantStock {
  isInStock?: boolean;
  availableQuantity?: number;
}

export interface ProductVariantMedia {
  url: string;
  id?: string;
  mediaType?: string;
  sortOrder?: number;
}

export interface ProductVariant {
  id: string;
  sku?: string;
  title?: string;
  price?: number | string;
  extractedPrice?: number;
  originalPrice?: string;
  extractedOriginalPrice?: number;
  oldPrice?: number;
  currency?: string;
  isDefault?: boolean;
  inStock?: boolean;
  availableStock?: number;
  stock?: ProductVariantStock;
  attributes?: Record<string, string | number | boolean | unknown>;
  media?: ProductVariantMedia[];
  thumbnail?: string;
  color?: string;
}

export interface ProductColorOption {
  name: string;
  class: string;
  variantIndex: number;
}

export interface ProductVariationAxis {
  key: string;
  label: string;
  values: string[];
}

export interface ProductRatingSummary {
  averageRating: number;
  reviewCount: number;
  ratingDistribution?: Record<number, number>;
  distribution?: Record<number, number>;
}

export interface Product {
  id: string;
  slug: string;
  title: string;
  name?: string;
  product?: { en?: string; ar?: string };
  sku?: string;
  variantId?: string;
  status?: string;
  description?: string;
  highlights?: string[];
  productType?: string;
  variationAxes?: string[];
  price?: string | number;
  extractedPrice?: number;
  originalPrice?: string;
  extractedOriginalPrice?: number;
  currency?: string;
  thumbnail?: string;
  img?: string;
  image?: string;
  inStock?: boolean;
  stockQuantity?: number;
  stockCount?: number;
  stock?: number;
  brand?: { id?: string; name?: string } | string;
  category?: { id?: string; name?: string } | string;
  categories?: Array<string | { id?: string; name?: string }>;
  variants?: ProductVariant[];
}

export interface ProductDetails extends Product {
  brand?: { id?: string; name?: string };
  category?: { id?: string; name?: string } | string;
  variants: ProductVariant[];
  tags?: Tag[];
  rating?: ProductRatingSummary;
}

export interface Review {
  id: string;
  productId: string;
  userId: string;
  orderItemId?: string;
  user?: { name?: string; avatar?: string } | null;
  rating: number;
  title?: string;
  body?: string;
  status?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ReviewEligibility {
  canReview: boolean;
}

/** Legacy / UI normalized product entity */
export interface ProductDetailsEntity {
  id: string | number;
  slug?: string;
  sku?: string;
  product?: {
    en?: string;
    ar?: string;
  } | string;
  name?: string;
  title?: string;
  price: number;
  extractedPrice?: number;
  oldPrice?: number;
  originalPrice?: string;
  extractedOriginalPrice?: number;
  currency?: string;
  discountPercentage?: string;
  stars?: number;
  reviews?: number;
  category: string;
  categoryId?: string;
  brand?: string | { id?: string; name?: string };
  img?: string;
  images?: string[];
  colors?: ProductColorOption[];
  variationAxes?: ProductVariationAxis[];
  variants?: ProductVariant[];
  inStock?: boolean;
  stockCount?: number;
  description?: string;
  highlights?: string[];
}

/** UI review model */
export interface ProductReview {
  id: string | number;
  author: string;
  avatar: string | null;
  rating: number;
  date: string;
  title: string;
  content: string;
  helpful: number;
  verified: boolean;
  userVoted?: boolean;
}
