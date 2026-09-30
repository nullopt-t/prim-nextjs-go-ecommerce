"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { catalogService } from "@/services/catalog";

export function normalizeProduct(raw: any) {
  if (!raw) return null;

  // Handle both backend Go format and legacy mock format
  const id = raw.id || raw.slug || "";
  const slug = raw.slug || raw.id || "";
  const titleStr = typeof raw.title === "string" 
    ? raw.title 
    : (raw.product?.en || raw.product?.ar || raw.name || "Product");

  const product = {
    en: typeof raw.product?.en === "string" ? raw.product.en : titleStr,
    ar: typeof raw.product?.ar === "string" ? raw.product.ar : titleStr,
  };

  const img = raw.thumbnail || raw.img || raw.image || (raw.images && raw.images[0]) || "/placeholder-product.png";
  
  // Price formatting
  const priceNum = raw.extractedPrice !== undefined
    ? Number(raw.extractedPrice)
    : (typeof raw.price === "number" ? raw.price : parseFloat(String(raw.price || "0").replace(/[^0-9.-]+/g, "")) || 0);

  const oldPriceNum = raw.extractedOriginalPrice !== undefined
    ? Number(raw.extractedOriginalPrice)
    : (raw.originalPrice || raw.oldPrice ? parseFloat(String(raw.originalPrice || raw.oldPrice).replace(/[^0-9.-]+/g, "")) : undefined);

  let discountPercentage = raw.discountPercentage;
  if (!discountPercentage && oldPriceNum && oldPriceNum > priceNum) {
    discountPercentage = `${Math.round(((oldPriceNum - priceNum) / oldPriceNum) * 100)}%`;
  }

  const stars = raw.rating?.averageRating !== undefined 
    ? raw.rating.averageRating 
    : (raw.stars !== undefined ? Number(raw.stars) : 4.5);

  const reviews = raw.rating?.reviewCount !== undefined
    ? raw.rating.reviewCount
    : (raw.reviews !== undefined ? Number(raw.reviews) : 0);

  const variantMedia: string[] = [];
  if (Array.isArray(raw.variants)) {
    raw.variants.forEach((v: any) => {
      if (Array.isArray(v.media)) {
        v.media.forEach((m: any) => {
          if (m.url && !variantMedia.includes(m.url)) variantMedia.push(m.url);
        });
      } else if (v.thumbnail && !variantMedia.includes(v.thumbnail)) {
        variantMedia.push(v.thumbnail);
      }
    });
  }

  const images = variantMedia.length > 0 
    ? variantMedia 
    : (Array.isArray(raw.images) && raw.images.length > 0 ? raw.images : [img]);

  // Extract unique colors from variants
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
        return {
          ...v,
          price: vPrice,
          extractedPrice: vPrice,
          oldPrice: vOldPrice,
          extractedOriginalPrice: vOldPrice,
        };
      })
    : [];

  return {
    ...raw,
    id,
    slug,
    product,
    title: titleStr,
    img,
    images,
    colors,
    variants,
    price: priceNum,
    oldPrice: oldPriceNum,
    discountPercentage,
    stars,
    reviews,
    inStock,
    stockCount,
    brand: typeof raw.brand === "string" ? raw.brand : raw.brand?.name || "PRIM",
    category: categoryName,
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
