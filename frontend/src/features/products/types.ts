export interface ProductVariantStock {
  isInStock?: boolean;
  availableQuantity?: number;
}

export interface ProductVariant {
  id: string;
  sku?: string;
  title?: string;
  price?: number;
  extractedPrice?: number;
  originalPrice?: string;
  extractedOriginalPrice?: number;
  oldPrice?: number;
  currency?: string;
  isDefault?: boolean;
  stock?: ProductVariantStock;
  attributes?: {
    color?: string;
    size?: string;
    [key: string]: any;
  };
  media?: Array<{ url: string; id?: string; mediaType?: string; sortOrder?: number }>;
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

export interface ProductDetailsEntity {
  id: string | number;
  slug?: string;
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
}

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

export interface ProductRatingSummary {
  averageRating: number;
  reviewCount: number;
  ratingDistribution?: Record<number, number>;
}
