import {
  ProductDetailsEntity,
  ProductVariant,
} from "../types";
import { normalizeMediaUrl } from "@/shared/lib/media";

/** A product image URL as returned by the catalog backend: a plain string or an
 * object exposing `url` / `publicUrl` / `src`. */
type RawImage = string | { url?: string; publicUrl?: string; src?: string };

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
    return variants.find((v) => v.isDefault) || variants[0];
  }

  // Match on the axes the caller actually selected: a partial selection
  // (e.g. color picked, size momentarily empty) still resolves to a variant
  // carrying that color instead of silently falling back to the default.
  const selectedAxes = axes.filter((axis) => Boolean(selectedValues[axis.key]));
  if (selectedAxes.length > 0) {
    const match = variants.find((v) =>
      selectedAxes.every((axis) => {
        const value = v.attributes?.[axis.key];
        return value !== undefined && String(value) === selectedValues[axis.key];
      })
    );
    if (match) return match;
  }

  return variants.find((v) => v.isDefault) || variants[0];
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

export interface ProductImageEntry {
  url: string;
  variantId?: string;
  attributes?: Record<string, unknown>;
}

/**
 * Resolves clean image entries for the product gallery:
 * - Prioritizes the active variant's thumbnail and preview media gallery at the beginning.
 * - Followed by the master product thumbnail and any product images.
 * - Followed by media from other variants of the product, fully deduplicated and normalized.
 *
 * Each entry keeps the variant it belongs to so the gallery can drive option
 * selection in both directions.
 */
export function resolveProductImageEntries(
  product: ProductDetailsEntity,
  activeVariant?: ProductVariant | null
): ProductImageEntry[] {
  const imageList: ProductImageEntry[] = [];
  const seen = new Set<string>();
  // The same image URL can be reused by several variants. Pick the variant
  // where it appears earliest (thumbnail, then media order) as its owner so a
  // gallery click can still resolve to a concrete option.
  const variantUrlPriority = new Map<string, { variant: ProductVariant; priority: number }>();

  const normalizeImageUrl = (rawUrl?: RawImage): string | null => {
    if (!rawUrl) return null;
    let urlStr: RawImage | undefined = rawUrl;
    if (typeof rawUrl === "object") {
      urlStr = rawUrl.publicUrl || rawUrl.url || rawUrl.src;
    }
    if (typeof urlStr !== "string") return null;
    const trimmed = urlStr.trim();
    if (!trimmed || trimmed === "/placeholder-product.png") return null;
    const normalized = normalizeMediaUrl(trimmed);
    if (!normalized || normalized === "/placeholder-product.png") return null;
    return normalized;
  };

  const getVariantImageUrls = (variant: ProductVariant): string[] => {
    const urls: string[] = [];
    const thumbnail = variant.thumbnail as unknown as RawImage;
    const vThumb =
      typeof thumbnail === "string"
        ? thumbnail
        : thumbnail?.url || thumbnail?.publicUrl;
    if (vThumb) urls.push(vThumb);

    if (Array.isArray(variant.media)) {
      for (const m of variant.media) {
        const mUrl =
          typeof m === "string"
            ? m
            : (m?.url ||
              (m as { publicUrl?: string })?.publicUrl ||
              (m as { object?: { publicUrl?: string } })?.object?.publicUrl);
        if (mUrl) urls.push(mUrl);
      }
    }
    return urls;
  };

  // Index variant-owned URLs first. Some normalized catalog responses expose the
  // same URLs in the master `images` array as well; without this lookup those
  // generic entries would lose their variant association.
  for (const variant of product.variants || []) {
    const urls = getVariantImageUrls(variant);
    urls.forEach((rawUrl, priority) => {
      const normalized = normalizeImageUrl(rawUrl);
      if (!normalized) return;
      const existing = variantUrlPriority.get(normalized);
      if (!existing || priority < existing.priority) {
        variantUrlPriority.set(normalized, { variant, priority });
      }
    });
  }

  const addImage = (rawUrl?: RawImage) => {
    const normalized = normalizeImageUrl(rawUrl);
    if (!normalized || seen.has(normalized)) return;

    const owner = variantUrlPriority.get(normalized)?.variant ?? null;
    seen.add(normalized);
    imageList.push({
      url: normalized,
      variantId: owner?.id,
      attributes: owner?.attributes as Record<string, unknown> | undefined,
    });
  };

  const addVariantMedia = (variant: ProductVariant) => {
    for (const url of getVariantImageUrls(variant)) {
      addImage(url);
    }
  };

  // 1. Active variant's own media first
  if (activeVariant) {
    addVariantMedia(activeVariant);
  }

  // 2. Product master thumbnail & img
  addImage((product as { thumbnail?: string; thumbnailUrl?: string })?.thumbnail);
  addImage((product as { thumbnailUrl?: string })?.thumbnailUrl);
  addImage(product.img);

  // 3. Product images list
  if (Array.isArray(product.images)) {
    for (const img of product.images) {
      addImage(img);
    }
  }

  // 4. Media from other variants so the customer can browse every product picture
  if (Array.isArray(product.variants)) {
    for (const v of product.variants) {
      if (activeVariant && v.id === activeVariant.id) continue;
      addVariantMedia(v);
    }
  }

  if (imageList.length === 0) {
    return [{ url: "/placeholder-product.png" }];
  }

  return imageList;
}

/**
 * Turns the store's estimated delivery window setting (e.g. "2 - 4 business
 * days") into localized arrival dates. Returns null when the setting cannot be
 * read as a day range so callers can fall back to the raw string.
 */
export function resolveDeliveryWindow(
  estimatedDeliveryDays: string | undefined,
  locale: string,
  now: Date = new Date(),
): { start: string; end: string } | null {
  const raw = estimatedDeliveryDays?.trim();
  if (!raw) return null;
  // Only day-based windows become dates; hour/week wording keeps the raw value.
  if (!/day/i.test(raw) && !/^\s*\d+\s*(?:[-–—]|to)?\s*\d*\s*$/i.test(raw)) return null;

  const numbers = raw.match(/\d+/g);
  if (!numbers) return null;
  const min = Number(numbers[0]);
  const max = numbers.length > 1 ? Number(numbers[1]) : min;
  if (!Number.isFinite(min) || !Number.isFinite(max) || max < min) return null;

  const formatter = new Intl.DateTimeFormat(locale.startsWith("ar") ? "ar-EG" : locale, {
    weekday: "short",
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  });
  const addDays = (days: number) =>
    // shortcut: business days are added as calendar days; upgrade if exact working-day arrival matters.
    new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() + days));

  return { start: formatter.format(addDays(min)), end: formatter.format(addDays(max)) };
}

/**
 * Backwards-compatible URL-only view of the product gallery entries.
 */
export function resolveProductImages(
  product: ProductDetailsEntity,
  activeVariant?: ProductVariant | null
): string[] {
  return resolveProductImageEntries(product, activeVariant).map((image) => image.url);
}
