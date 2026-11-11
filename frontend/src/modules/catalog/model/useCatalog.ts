"use client";

import { useCatalogContext, normalizeProduct } from "@/modules/catalog/model/catalogContext";
import { useState, useEffect } from "react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { ProductDetailsEntity } from "../types";

export function useAllProducts() {
  const { products, loading, errorMsg, refreshCatalog } = useCatalogContext();
  return { products, loading, errorMsg, refreshCatalog };
}

/**
 * Resolves a product by slug/id.
 *
 * When `initialProduct` is provided by a Server Component, it is used for the
 * first render and no client fetch happens on mount. Navigating to a different
 * slug still fetches the new product.
 */
export function useProductBySlug(
  slug?: string | string[],
  initialProduct?: ProductDetailsEntity | null,
) {
  const cleanSlug = Array.isArray(slug) ? slug[0] : slug;

  const [product, setProduct] = useState<ProductDetailsEntity | null>(initialProduct ?? null);
  const [loading, setLoading] = useState(!initialProduct);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const { getProductBySlug: getFromContext } = useCatalogContext();

  useEffect(() => {
    const matchesInitial =
      initialProduct &&
      cleanSlug &&
      (initialProduct.id === cleanSlug || initialProduct.slug === cleanSlug);

    if (matchesInitial) {
      setProduct(initialProduct);
      setLoading(false);
      return;
    }

    let isMounted = true;

    const fetchDetailedProduct = async () => {
      setLoading(true);
      try {
        if (!cleanSlug) return;
        const res: any = await catalogService.getProductBySlug(cleanSlug);
        const raw = res?.data || res;
        if (isMounted && raw) {
          setProduct(normalizeProduct(raw) as unknown as ProductDetailsEntity);
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
  }, [cleanSlug, getFromContext, initialProduct]);

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
