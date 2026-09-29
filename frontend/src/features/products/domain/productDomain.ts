export interface ProductVariantStock {
  isInStock?: boolean;
  availableQuantity?: number;
}

export interface ProductVariant {
  id: string;
  sku?: string;
  price?: number;
  stock?: ProductVariantStock;
  attributes?: {
    color?: string;
    size?: string;
    [key: string]: any;
  };
  media?: Array<{ url: string }>;
  thumbnail?: string;
  color?: string;
}

export interface ProductColorOption {
  name: string;
  class: string;
  variantIndex: number;
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
  oldPrice?: number;
  discountPercentage?: string;
  stars?: number;
  reviews?: number;
  category: string;
  categoryId?: string;
  brand?: string | { id?: string; name?: string };
  img?: string;
  images?: string[];
  colors?: ProductColorOption[];
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

/**
 * Resolves localized display name for a product
 */
export function getProductDisplayName(product: ProductDetailsEntity, locale: string = "en"): string {
  if (typeof product.product === "object" && product.product !== null) {
    if (locale === "ar" && product.product.ar) return product.product.ar;
    return product.product.en || product.product.ar || "Product";
  }
  return (product.product as string) || product.title || product.name || "Product";
}

/**
 * Resolves active variant, stock, and availability state
 */
export function resolveProductStock(
  product: ProductDetailsEntity,
  selectedColorIndex: number
): {
  activeVariant: ProductVariant | null;
  availableStock: number;
  isVariantInStock: boolean;
} {
  const color = product.colors && product.colors[selectedColorIndex];
  const activeVariantIndex = color?.variantIndex !== undefined ? color.variantIndex : 0;
  
  const activeVariant =
    (product.variants && product.variants[activeVariantIndex]) ||
    (product.variants && product.variants[0]) ||
    null;

  const availableStock = activeVariant?.stock?.availableQuantity ?? product.stockCount ?? 10;
  const isVariantInStock = (activeVariant?.stock?.isInStock ?? product.inStock ?? true) && availableStock > 0;

  return {
    activeVariant,
    availableStock,
    isVariantInStock,
  };
}

/**
 * Resolves clean image array for product gallery, prioritizing the active variant's media
 */
export function resolveProductImages(
  product: ProductDetailsEntity,
  activeVariant?: ProductVariant | null
): string[] {
  // If active variant has media images, use them
  if (activeVariant?.media && activeVariant.media.length > 0) {
    const urls = activeVariant.media.map((m) => m.url).filter(Boolean);
    if (urls.length > 0) return urls;
  }

  // If active variant has a thumbnail
  if (activeVariant?.thumbnail) {
    return [activeVariant.thumbnail];
  }

  // Fallback to product images
  if (product.images && product.images.length > 0) {
    return product.images;
  }
  if (product.img) {
    return [product.img];
  }
  return ["/placeholder-product.png"];
}
