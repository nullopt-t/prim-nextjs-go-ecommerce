"use client";

import { createContext, useContext, useState, useEffect, useCallback } from "react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { normalizeProduct } from "./normalizeProduct";
import type { Brand, Category, ProductDetailsEntity, Tag } from "../types";

export { normalizeProduct };

interface CatalogFilters {
  brand: string | null;
  inStockOnly: boolean;
  rating: number | null;
  discount: number | null;
  minPrice: number | null;
  maxPrice: number | null;
  sort: string | null;
  /** Selected variant attribute facets, one value per key: { color: "black" }. */
  attributes: Record<string, string>;
}

interface CatalogContextType {
  products: ProductDetailsEntity[];
  categories: Category[];
  brands: Brand[];
  tags: Tag[];
  loading: boolean;
  errorMsg: string | null;
  filters: CatalogFilters;
  setFilter: (key: keyof CatalogFilters, value: CatalogFilters[keyof CatalogFilters]) => void;
  clearFilters: () => void;
  getProductBySlug: (slug: string) => ProductDetailsEntity | null;
  refreshCatalog: () => Promise<void>;
}

const CatalogContext = createContext<CatalogContextType | null>(null);

export interface CatalogProviderProps {
  children: React.ReactNode;
  initialProducts?: ProductDetailsEntity[];
  initialCategories?: Category[];
  initialBrands?: Brand[];
}

export function CatalogProvider({
  children,
  initialProducts,
  initialCategories,
  initialBrands,
}: CatalogProviderProps) {
  const hasInitialData = Array.isArray(initialProducts);
  const [products, setProducts] = useState<ProductDetailsEntity[]>(initialProducts ?? []);
  const [categories, setCategories] = useState<Category[]>(initialCategories ?? []);
  const [brands, setBrands] = useState<Brand[]>(initialBrands ?? []);
  const [tags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(!hasInitialData);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [filters, setFilters] = useState<CatalogFilters>({
    brand: null,
    inStockOnly: false,
    rating: null,
    discount: null,
    minPrice: null,
    maxPrice: null,
    sort: null,
    attributes: {},
  });

  const setFilter = (
    key: keyof CatalogFilters,
    value: CatalogFilters[keyof CatalogFilters]
  ) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const clearFilters = () => {
    setFilters({
      brand: null,
      inStockOnly: false,
      rating: null,
      discount: null,
      minPrice: null,
      maxPrice: null,
      sort: null,
      attributes: {},
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
        setProducts(rawList.map(normalizeProduct) as unknown as ProductDetailsEntity[]);
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
        setBrands(
          rawBrands.map((b: Brand) => ({
            ...b,
            logoUrl: b.logoUrl || b.logo ? normalizeMediaUrl(b.logoUrl || b.logo) : undefined,
            logo: b.logoUrl || b.logo ? normalizeMediaUrl(b.logoUrl || b.logo) : undefined,
          }))
        );
      }
    } catch (err) {
      setErrorMsg((err as { message?: string }).message || "Failed to load catalog");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!hasInitialData) {
      fetchCatalog();
    }
  }, [fetchCatalog, hasInitialData]);

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

/** For consoles that may render outside the storefront catalog provider. */
export function useOptionalCatalogContext() {
  return useContext(CatalogContext);
}
