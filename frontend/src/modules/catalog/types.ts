/**
 * Catalog domain and response models matching Go backend catalog handlers.
 */

export interface Category {
  id: string;
  name: string;
  slug?: string;
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

export interface CategoryAttribute {
  key: string;
  label: string;
  valueType: string;
  allowedValues?: string[];
  isFilterable: boolean;
  sortOrder: number;
}

export interface Brand {
  id: string;
  name: string;
  link?: string;
  logoUrl?: string;
  logo?: string;
  productCount?: number;
}

export interface Tag {
  id: string;
  name: string;
  /** The backend tag summary carries no slug; it is optional on purpose. */
  slug?: string;
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
  [key: string]: unknown;
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
  /** Volume-discount thresholds; the PDP normalizes them before display. */
  tierPricing?: unknown;
  [key: string]: unknown;
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
  /** Backend sends the star breakdown keyed by score ("1".."5"). */
  distribution?: Record<string, number>;
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
  categoryId?: string;
  brandId?: string;
  categories?: Array<string | { id?: string; name?: string }>;
  variants?: ProductVariant[];
  [key: string]: unknown;
}

export interface ProductDetails extends Product {
  brand?: { id?: string; name?: string };
  category?: { id?: string; name?: string } | string;
  categoryId?: string;
  brandId?: string;
  updatedAt?: string;
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
  /** Live title of the reviewed product, used by the reviewer profile page. */
  productTitle?: string;
  rating: number;
  title?: string;
  body?: string;
  status?: string;
  helpfulCount?: number;
  userVoted?: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ReviewEligibility {
  canReview: boolean;
}

/** One product included in a bundle, with its own availability. */
export interface ProductBundleItem {
  id: string;
  title: string;
  slug: string;
  thumbnail?: string;
  quantity: number;
  variant?: string;
  availableQuantity: number;
  isInStock: boolean;
}

/**
 * A bundle's contents and the availability derived from them. A bundle keeps no
 * inventory of its own, so this is the only meaningful stock figure for a kit.
 */
export interface ProductBundle {
  itemCount: number;
  availableQuantity: number;
  isInStock: boolean;
  items: ProductBundleItem[];
}

/** Legacy / UI normalized product entity */
export interface ProductDetailsEntity {
  id: string | number;
  slug?: string;
  sku?: string;
  /** Variant ids the payload points at, kept for cart payloads. */
  variantId?: string;
  defaultVariantId?: string;
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
  bundle?: ProductBundle;
  /** Authoritative review aggregate from the backend (all approved reviews). */
  rating?: ProductRatingSummary;
  /** Taxonomy tags assigned to the product. */
  tags?: Tag[];
  /** Volume-discount tiers; the PDP normalizes them before display. */
  tierPricing?: unknown;
}

/** UI review model */
export interface ProductReview {
  id: string | number;
  author: string;
  avatar: string | null;
  /** Author's user UUID, used to link to their public review profile. */
  userId: string | null;
  /** Reviewed product UUID, used to link from the profile page to the PDP. */
  productId: string | null;
  /** Live title of the reviewed product on the reviewer profile page. */
  productTitle: string | null;
  rating: number;
  date: string;
  title: string;
  content: string;
  helpful: number;
  userVoted?: boolean;
}
