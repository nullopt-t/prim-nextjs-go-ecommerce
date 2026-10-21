"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { catalogService } from "@/services/catalog";

export function normalizeProduct(raw: any) {
  if (!raw) return null;

  // Handle both backend Go format, WishlistItem envelope, and legacy mock format
  const isWishlistItem = Boolean(raw.product && typeof raw.product === "object" && (raw.product.title || raw.product.slug));
  const p = isWishlistItem ? raw.product : raw;
  const wishlistItemId = isWishlistItem ? raw.id : (raw.wishlistItemId || undefined);

  const id = p.id || raw.productId || p.slug || raw.id || "";
  const slug = p.slug || raw.slug || id;
  const titleStr = typeof p.title === "string" 
    ? p.title 
    : (p.product?.en || p.product?.ar || p.name || raw.title || "Product");

  const product = {
    en: typeof p.product?.en === "string" ? p.product.en : titleStr,
    ar: typeof p.product?.ar === "string" ? p.product.ar : titleStr,
  };

  const img = p.thumbnailUrl || p.thumbnail || p.img || p.image || (p.images && p.images[0]) || raw.thumbnail || raw.img || "/placeholder-product.png";
  
  // Price formatting - check p (inner product if wishlistItem, or raw) then fallback to raw
  const rawPrice = p.extractedPrice !== undefined ? p.extractedPrice : raw.extractedPrice;
  const priceNum = rawPrice !== undefined
    ? Number(rawPrice)
    : (typeof (p.price ?? raw.price) === "number"
        ? (p.price ?? raw.price)
        : parseFloat(String(p.price || raw.price || "0").replace(/[^0-9.-]+/g, "")) || 0);

  const rawOldPrice = p.extractedOriginalPrice !== undefined ? p.extractedOriginalPrice : raw.extractedOriginalPrice;
  const rawOrigPriceStr = p.originalPrice || raw.originalPrice || p.oldPrice || raw.oldPrice;
  const oldPriceNum = rawOldPrice !== undefined
    ? Number(rawOldPrice)
    : (rawOrigPriceStr ? parseFloat(String(rawOrigPriceStr).replace(/[^0-9.-]+/g, "")) : undefined);

  let discountPercentage = p.discountPercentage || raw.discountPercentage;
  if (!discountPercentage && oldPriceNum && oldPriceNum > priceNum) {
    discountPercentage = `${Math.round(((oldPriceNum - priceNum) / oldPriceNum) * 100)}%`;
  }

  const stars = (p.rating?.averageRating ?? raw.rating?.averageRating) !== undefined 
    ? (p.rating?.averageRating ?? raw.rating?.averageRating)
    : ((p.stars ?? raw.stars) !== undefined ? Number(p.stars ?? raw.stars) : 4.5);

  const reviews = (p.rating?.reviewCount ?? raw.rating?.reviewCount) !== undefined
    ? (p.rating?.reviewCount ?? raw.rating?.reviewCount)
    : ((p.reviews ?? raw.reviews) !== undefined ? Number(p.reviews ?? raw.reviews) : 0);

  const variantMedia: string[] = [];
  if (Array.isArray(raw.variants)) {
    raw.variants.forEach((v: any) => {
      const vThumb = typeof v.thumbnail === "string" ? v.thumbnail : (v.thumbnail?.url || v.thumbnail?.publicUrl);
      if (vThumb && !variantMedia.includes(vThumb)) {
        variantMedia.push(vThumb);
      }
      if (Array.isArray(v.media)) {
        v.media.forEach((m: any) => {
          const mUrl = typeof m === "string" ? m : m?.url;
          if (mUrl && !variantMedia.includes(mUrl)) variantMedia.push(mUrl);
        });
      }
    });
  }

  // Include master product thumbnail if available
  if (img && img !== "/placeholder-product.png" && !variantMedia.includes(img)) {
    variantMedia.push(img);
  }

  const images = variantMedia.length > 0 
    ? variantMedia 
    : (Array.isArray(raw.images) && raw.images.length > 0 ? raw.images : [img]);

  // Extract unique colors from variants (kept for legacy consumers)
  const colors: { name: string; class: string; variantIndex: number }[] = [];
  const seenColors = new Set<string>();
  if (Array.isArray(raw.variants)) {
    raw.variants.forEach((v: any, idx: number) => {
      const colorName = v.attributes?.color || v.color;
      if (colorName && !seenColors.has(colorName)) {
        seenColors.add(colorName);
        colors.push({
          name: colorName,
          class: `bg-gray-400`, // fallback swatch color
          variantIndex: idx,
        });
      }
    });
  }

  // Build generic variation axes. Products with declared axes use those keys;
  // legacy rows fall back to keys shared by every child.
  const axisLabels: Record<string, string> = {
    color: "Color",
    colour: "Color",
    size: "Size",
    storage: "Storage",
    ram: "RAM",
    memory: "Memory",
    capacity: "Capacity",
    material: "Material",
    style: "Style",
    band: "Band",
    finish: "Finish",
    model: "Model",
  };
  const axisLabel = (key: string) =>
    axisLabels[key] || key.replace(/[_-]+/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());

  const rawVariants = Array.isArray(raw.variants) ? raw.variants : [];
  let axisKeys: string[] = [];
  const declaredAxes = raw.variationAxes ?? p.variationAxes;
  if (Array.isArray(declaredAxes) && declaredAxes.length > 0) {
    axisKeys = declaredAxes.filter((k: any) => typeof k === "string" && k.trim() !== "");
  } else if (rawVariants.length > 1) {
    const keySets = rawVariants.map((v: any) =>
      v.attributes && typeof v.attributes === "object" ? Object.keys(v.attributes) : []
    );
    axisKeys = keySets[0].filter((key: string) => keySets.every((keys: string[]) => keys.includes(key)));
  }

  const variationAxes = axisKeys.map((key) => {
    const values: string[] = [];
    for (const v of rawVariants) {
      const value = v.attributes?.[key] ?? v[key];
      if (value === undefined || value === null || value === "") continue;
      const str = String(value);
      if (!values.includes(str)) values.push(str);
    }
    return { key, label: axisLabel(key), values };
  }).filter((axis) => axis.values.length > 0);

  // Extract stock info from first variant
  const firstVariant = Array.isArray(raw.variants) && raw.variants.length > 0 ? raw.variants[0] : null;
  const inStock = firstVariant?.stock?.isInStock ?? raw.inStock ?? true;
  const stockCount = firstVariant?.stock?.availableQuantity ?? raw.stockCount ?? 0;

  // Extract category — backend list/detail may provide category object { id, name } or string or categories array
  const categoryName = typeof raw.category === "string"
    ? raw.category
    : (raw.category?.name || raw.categories?.[0]?.name || "General");
  const categoryId = raw.category?.id || raw.categories?.[0]?.id || "";

  const variants = Array.isArray(raw.variants)
    ? raw.variants.map((v: any) => {
        const vPrice = v.extractedPrice !== undefined
          ? Number(v.extractedPrice)
          : (typeof v.price === "number"
              ? v.price
              : parseFloat(String(v.price || "0").replace(/[^0-9.-]+/g, "")) || priceNum);
        const vOldPrice = v.extractedOriginalPrice !== undefined
          ? Number(v.extractedOriginalPrice)
          : (v.originalPrice || v.oldPrice ? parseFloat(String(v.originalPrice || v.oldPrice).replace(/[^0-9.-]+/g, "")) : undefined);
        const vThumb = typeof v.thumbnail === "string" ? v.thumbnail : (v.thumbnail?.url || v.thumbnail?.publicUrl || undefined);
        return {
          ...v,
          thumbnail: vThumb,
          price: vPrice,
          extractedPrice: vPrice,
          oldPrice: vOldPrice,
          extractedOriginalPrice: vOldPrice,
          originalPrice: v.originalPrice || (vOldPrice !== undefined ? `E£${vOldPrice.toFixed(2)}` : undefined),
        };
      })
    : [];

  const variantId = raw.variantId || p.variantId || undefined;
  const defaultVariantId = raw.defaultVariantId || raw.variantId || p.variantId || undefined;

  return {
    ...raw,
    id,
    productId: id,
    wishlistItemId,
    variantId,
    defaultVariantId,
    slug,
    product,
    title: titleStr,
    img,
    images,
    colors,
    variationAxes,
    variants,
    price: priceNum,
    extractedPrice: priceNum,
    oldPrice: oldPriceNum,
    extractedOriginalPrice: oldPriceNum,
    originalPrice: p.originalPrice || raw.originalPrice || (oldPriceNum !== undefined ? `E£${oldPriceNum.toFixed(2)}` : undefined),
    discountPercentage,
    stars,
    reviews,
    inStock,
    stockCount,
    brand: typeof (p.brand ?? raw.brand) === "string" 
      ? (p.brand ?? raw.brand) 
      : (p.brandName || p.brand?.name || raw.brand?.name || "PRIM"),
    category: categoryName || p.categoryName || "General",
    categoryId,
  };
}

interface CatalogContextType {
  products: any[];
  categories: any[];
  brands: any[];
  tags: any[];
  loading: boolean;
  errorMsg: string | null;
  filters: {
    brand: string | null;
    inStockOnly: boolean;
    rating: number | null;
    discount: number | null;
  };
  setFilter: (key: string, value: any) => void;
  clearFilters: () => void;
  getProductBySlug: (slug: string) => any;
  refreshCatalog: () => Promise<void>;
}

const CatalogContext = createContext<CatalogContextType | null>(null);

export function CatalogProvider({ children }: { children: React.ReactNode }) {
  const [products, setProducts] = useState<any[]>([]);
  const [categories, setCategories] = useState<any[]>([]);
  const [brands, setBrands] = useState<any[]>([]);
  const [tags, setTags] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [filters, setFilters] = useState({
    brand: null as string | null,
    inStockOnly: false,
    rating: null as number | null,
    discount: null as number | null,
  });

  const setFilter = (key: string, value: any) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const clearFilters = () => {
    setFilters({
      brand: null,
      inStockOnly: false,
      rating: null,
      discount: null,
    });
  };

  const fetchCatalog = useCallback(async () => {
    setLoading(true);
    try {
      // Delegate to catalogService instead of raw URL fetching
      const [prodRes, catRes, brandRes] = await Promise.allSettled([
        catalogService.getProducts({ pageSize: 100 }),
        catalogService.getCategories(50),
        catalogService.getBrands(50),
      ]);

      if (prodRes.status === "fulfilled" && prodRes.value) {
        const rawList = Array.isArray(prodRes.value) 
          ? prodRes.value 
          : (Array.isArray(prodRes.value.data) ? prodRes.value.data : []);
        setProducts(rawList.map(normalizeProduct));
      }

      if (catRes.status === "fulfilled" && catRes.value) {
        const rawCats = Array.isArray(catRes.value) 
          ? catRes.value 
          : (Array.isArray(catRes.value.data) ? catRes.value.data : []);
        setCategories(rawCats);
      }

      if (brandRes.status === "fulfilled" && brandRes.value) {
        const rawBrands = Array.isArray(brandRes.value) 
          ? brandRes.value 
          : (Array.isArray(brandRes.value.data) ? brandRes.value.data : []);
        setBrands(rawBrands);
      }
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to load catalog");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCatalog();
  }, [fetchCatalog]);

  const getProductBySlug = (slug: string) => {
    return products.find((p) => p.slug === slug || p.id === slug) || null;
  };

  return (
    <CatalogContext.Provider
      value={{
        products,
        categories,
        brands,
        tags,
        loading,
        errorMsg,
        filters,
        setFilter,
        clearFilters,
        getProductBySlug,
        refreshCatalog: fetchCatalog,
      }}
    >
      {children}
    </CatalogContext.Provider>
  );
}

export function useCatalogContext() {
  const context = useContext(CatalogContext);
  if (!context) {
    throw new Error("useCatalogContext must be used within a CatalogProvider");
  }
  return context;
}
