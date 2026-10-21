import {
  ProductDetailsEntity,
  ProductVariant,
} from "../types";
import { normalizeMediaUrl } from "@/utils/media";

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
 * Resolves the active child variant for a set of selected axis values.
 * Missing selections are ignored, and the product's default/first child is used
 * as a fallback when no exact match exists.
 */
export function resolveVariantByAttributes(
  product: ProductDetailsEntity,
  selectedValues: Record<string, string> = {}
): ProductVariant | null {
  const variants = product.variants || [];
  if (variants.length === 0) return null;

  const axes = product.variationAxes || [];
  if (axes.length === 0) {
    return variants.find((v) => (v as any).isDefault) || variants[0];
  }

  const hasAllSelections = axes.every((axis) => Boolean(selectedValues[axis.key]));
  if (hasAllSelections) {
    const match = variants.find((v) =>
      axes.every((axis) => {
        const value = v.attributes?.[axis.key];
        return value !== undefined && String(value) === selectedValues[axis.key];
      })
    );
    if (match) return match;
  }

  return variants.find((v) => (v as any).isDefault) || variants[0];
}

/**
 * Resolves stock for an explicitly selected variant.
 */
export function resolveVariantStock(
  product: ProductDetailsEntity,
  activeVariant: ProductVariant | null
): {
  availableStock: number;
  isVariantInStock: boolean;
} {
  const availableStock = activeVariant?.stock?.availableQuantity ?? product.stockCount ?? 0;
  const isVariantInStock =
    (activeVariant?.stock?.isInStock ?? (availableStock > 0)) && availableStock > 0;
  return { availableStock, isVariantInStock };
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

  const availableStock = activeVariant?.stock?.availableQuantity ?? product.stockCount ?? 0;
  const isVariantInStock = (activeVariant?.stock?.isInStock ?? (availableStock > 0)) && availableStock > 0;

  return {
    activeVariant,
    availableStock,
    isVariantInStock,
  };
}

/**
 * Resolves clean image array for product gallery:
 * - Prioritizes the active variant's thumbnail and preview media gallery at the beginning.
 * - Followed by the master product thumbnail and any product images.
 * - Followed by media from other variants of the product, fully deduplicated and normalized.
 */
export function resolveProductImages(
  product: ProductDetailsEntity,
  activeVariant?: ProductVariant | null
): string[] {
  const imageList: string[] = [];
  const seen = new Set<string>();

  const addImage = (rawUrl?: string | null) => {
    if (!rawUrl || typeof rawUrl !== "string") return;
    const trimmed = rawUrl.trim();
    if (!trimmed || trimmed === "/placeholder-product.png") return;
    const normalized = normalizeMediaUrl(trimmed);
    if (normalized && !seen.has(normalized)) {
      seen.add(normalized);
      imageList.push(normalized);
    }
  };

  // 1. Active variant's specific thumbnail first
  if (activeVariant) {
    const vThumb = typeof activeVariant.thumbnail === "string" 
      ? activeVariant.thumbnail 
      : (activeVariant.thumbnail as any)?.url || (activeVariant.thumbnail as any)?.publicUrl;
    addImage(vThumb);

    // 2. Active variant's preview media gallery
    if (Array.isArray(activeVariant.media)) {
      for (const m of activeVariant.media) {
        const mUrl = typeof m === "string" ? m : m?.url;
        addImage(mUrl);
      }
    }
  }

  // 3. Product master thumbnail & img
  addImage((product as any)?.thumbnail);
  addImage((product as any)?.thumbnailUrl);
  addImage(product.img);

  // 4. Product images list
  if (Array.isArray(product.images)) {
    for (const img of product.images) {
      addImage(img);
    }
  }

  // 5. Media from other variants of this product so the customer can browse all product pictures
  if (Array.isArray(product.variants)) {
    for (const v of product.variants) {
      if (activeVariant && v.id === activeVariant.id) continue;
      const otherThumb = typeof v.thumbnail === "string" 
        ? v.thumbnail 
        : (v.thumbnail as any)?.url || (v.thumbnail as any)?.publicUrl;
      addImage(otherThumb);

      if (Array.isArray(v.media)) {
        for (const m of v.media) {
          const mUrl = typeof m === "string" ? m : m?.url;
          addImage(mUrl);
        }
      }
    }
  }

  if (imageList.length === 0) {
    return ["/placeholder-product.png"];
  }

  return imageList;
}
