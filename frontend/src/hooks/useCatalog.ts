"use client";

import { useCatalogContext, normalizeProduct } from "@/context/CatalogContext";
import { useState, useEffect } from "react";
import { catalogService } from "@/services/catalog";

export function useAllProducts() {
  const { products, loading, errorMsg, refreshCatalog } = useCatalogContext();
  return { products, loading, errorMsg, refreshCatalog };
}

export function useProductBySlug(slug?: string | string[]) {
  const [product, setProduct] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const { getProductBySlug: getFromContext } = useCatalogContext();

  useEffect(() => {
    let isMounted = true;
    const cleanSlug = Array.isArray(slug) ? slug[0] : slug;

    const fetchDetailedProduct = async () => {
      setLoading(true);
      try {
        if (!cleanSlug) return;
        const res: any = await catalogService.getProductBySlug(cleanSlug);
        const raw = res?.data || res;
        if (isMounted && raw) {
          setProduct(normalizeProduct(raw));
        }
      } catch (error: any) {
        if (isMounted) {
          const fallback = cleanSlug ? getFromContext(cleanSlug) : null;
          if (fallback) {
            setProduct(fallback);
          } else {
            setErrorMsg(error?.message || "Failed to load product details");
          }
        }
      } finally {
        if (isMounted) setLoading(false);
      }
    };

    if (cleanSlug) {
      fetchDetailedProduct();
    }

    return () => {
      isMounted = false;
    };
  }, [slug, getFromContext]);

  return { product, loading, errorMsg };
}

export function useCategories() {
  const { categories, loading, errorMsg } = useCatalogContext();
  return { categories, loading, errorMsg };
}

export function useTags() {
  const { tags, loading, errorMsg } = useCatalogContext();
  return { tags, loading, errorMsg };
}
