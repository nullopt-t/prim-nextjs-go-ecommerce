"use client";

import { useEffect, useRef, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { ProductsGrid } from "@/shared/ui/productsGrid";
import { CardDetails } from "@/shared/ui/productsCard";
import { useCatalogContext, normalizeProduct } from "@/modules/catalog/model/catalogContext";
import type { ProductDetailsEntity } from "@/modules/catalog/types";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { useCurrency, useDisplayCurrency } from "@/modules/settings/model/publicSettingsContext";
import { convertCurrencyRounded } from "@/shared/lib/currency";

interface ContentProps {
  products?: ProductDetailsEntity[];
  category?: string;
}

export default function Content({ products, category }: ContentProps) {
  const { products: contextProducts, loading: contextLoading, filters, setFilter, clearFilters } = useCatalogContext();
  const t = useTranslations("products");
  const baseCurrency = useCurrency();
  const displayCurrency = useDisplayCurrency();
  const [categoryProducts, setCategoryProducts] = useState<ProductDetailsEntity[] | null>(products ?? null);
  const [catLoading, setCatLoading] = useState(false);
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const searchQuery = (searchParams.get("q") || "").trim().toLowerCase();
  const searchParamsString = searchParams.toString();

  // Initialize filter state from the URL once, so shared links restore the view.
  const initializedRef = useRef(false);
  useEffect(() => {
    if (initializedRef.current) return;
    initializedRef.current = true;
    const params = new URLSearchParams(searchParamsString);
    const number = (key: string) => {
      const value = params.get(key);
      if (!value) return null;
      const parsed = Number(value);
      return Number.isFinite(parsed) && parsed >= 0 ? parsed : null;
    };
    const brand = params.get("brand");
    if (brand) setFilter("brand", brand);
    if (params.get("inStock") === "true") setFilter("inStockOnly", true);
    const rating = number("rating");
    if (rating !== null) setFilter("rating", rating);
    const discount = number("discount");
    if (discount !== null) setFilter("discount", discount);
    const minPrice = number("minPrice");
    if (minPrice !== null) setFilter("minPrice", minPrice);
    const maxPrice = number("maxPrice");
    if (maxPrice !== null) setFilter("maxPrice", maxPrice);
    const sort = params.get("sort");
    if (sort) setFilter("sort", sort);
    // Repeated params (?attribute=color:red&attribute=size:xl); also accept the
    // legacy comma-joined single param (?attribute=color:red,size:xl).
    const attributes: Record<string, string> = {};
    for (const attributeParam of params.getAll("attribute")) {
      for (const part of attributeParam.split(",")) {
        const colon = part.indexOf(":");
        if (colon <= 0 || colon >= part.length - 1) continue;
        const k = part.slice(0, colon).trim();
        const v = part.slice(colon + 1).trim();
        if (k && v) attributes[k] = v;
      }
    }
    if (Object.keys(attributes).length > 0) setFilter("attributes", attributes);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Push filter state back to the URL (debounced) so views are shareable and
  // the server can apply the same filters to the next request.
  useEffect(() => {
    if (!initializedRef.current) return;
    const current = filters;
    const timer = setTimeout(() => {
      const next = new URLSearchParams(searchParamsString);
      const apply = (key: string, value: string | null) => {
        if (value) next.set(key, value);
        else next.delete(key);
      };
      apply("brand", current.brand);
      apply("inStock", current.inStockOnly ? "true" : null);
      apply("rating", current.rating !== null ? String(current.rating) : null);
      apply("discount", current.discount !== null ? String(current.discount) : null);
      apply("minPrice", current.minPrice !== null ? String(current.minPrice) : null);
      apply("maxPrice", current.maxPrice !== null ? String(current.maxPrice) : null);
      apply("sort", current.sort);
      next.delete("attribute");
      for (const [k, v] of Object.entries(current.attributes || {})) {
        if (k.trim() === "" || v.trim() === "") continue;
        next.append("attribute", `${k.trim()}:${v.trim()}`);
      }
      next.delete("page");
      const query = next.toString();
      if (query !== searchParamsString) {
        router.replace(query ? `${pathname}?${query}` : pathname);
      }
    }, 400);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters, pathname, router, searchParamsString]);

  useEffect(() => {
    // Server-provided products already reflect the URL category/search.
    if (products) {
      setCategoryProducts(products);
      return;
    }

    // If no category or "all", we use the full context products
    if (!category || category.toLowerCase() === "all" || category.toLowerCase() === "all categories") {
      setCategoryProducts(null);
      return;
    }

    let isMounted = true;
    setCatLoading(true);

    catalogService
      .getProducts({ pageSize: 100, category })
      .then((res) => {
        if (!isMounted) return;
        const rawList = res.data ?? [];
        setCategoryProducts(rawList.map(normalizeProduct).filter(Boolean) as unknown as ProductDetailsEntity[]);
      })
      .catch((err) => {
        console.error("Failed to fetch products by category:", err);
        if (isMounted) setCategoryProducts([]);
      })
      .finally(() => {
        if (isMounted) setCatLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [category, products]);

  const activeProducts = categoryProducts !== null ? categoryProducts : (products || contextProducts);
  const loading = catLoading || (categoryProducts === null && contextLoading);

  const filtered = activeProducts.filter((p: ProductDetailsEntity) => {
    // Search query filter (matches title, brand, description or tags)
    if (searchQuery) {
      const productName =
        typeof p.product === "object"
          ? p.product?.en || p.product?.ar || ""
          : (p.product || "");
      const title = (p.title || productName || "").toLowerCase();
      const brand = (typeof p.brand === "string" ? p.brand : p.brand?.name || "").toLowerCase();
      const desc = (p.description || "").toLowerCase();
      if (!title.includes(searchQuery) && !brand.includes(searchQuery) && !desc.includes(searchQuery)) {
        return false;
      }
    }

    // Brand filter
    if (filters.brand) {
      const pBrand = (typeof p.brand === "string" ? p.brand : p.brand?.name || "").toLowerCase();
      if (!pBrand.includes(filters.brand.toLowerCase())) return false;
    }

    // In-stock filter
    if (filters.inStockOnly && !(p.inStock ?? true)) {
      return false;
    }

    // Rating filter
    if (filters.rating !== null && (p.stars || 0) < filters.rating) {
      return false;
    }

    // Discount filter
    if (filters.discount !== null) {
      const disc = parseFloat(String(p.discountPercentage || "0").replace(/[^0-9.]/g, "")) || 0;
      if (disc < filters.discount) return false;
    }

    // Min/max price filters are typed in the *display* currency (what the
    // shopper sees and types), so convert the base-currency price first —
    // comparing raw base numbers would filter in the wrong currency.
    if (filters.minPrice !== null || filters.maxPrice !== null) {
      const rawPrice = typeof p.price === "number" ? p.price : (p.extractedPrice ?? 0);
      const priceVal = convertCurrencyRounded(rawPrice, baseCurrency, displayCurrency);
      if (filters.minPrice !== null && priceVal < filters.minPrice) return false;
      if (filters.maxPrice !== null && priceVal > filters.maxPrice) return false;
    }

    // Attribute facet filter: one variant must match every selected key:value
    // (mirrors the backend EXISTS clause on /api/v1/products?attribute=).
    const selectedFacets = Object.entries(filters.attributes || {});
    if (selectedFacets.length > 0) {
      const variants = (p.variants ?? []) as Array<{ attributes?: Record<string, unknown> }>;
      const matchesSomeVariant = variants.some((v) =>
        selectedFacets.every(([key, value]) => {
          const attr = v.attributes?.[key];
          return attr !== undefined && attr !== null && String(attr) === value;
        }),
      );
      if (!matchesSomeVariant) return false;
    }

    return true;
  });

  if (filters.sort) {
    filtered.sort((a, b) => {
      const priceA = typeof a.price === "number" ? a.price : (a.extractedPrice ?? 0);
      const priceB = typeof b.price === "number" ? b.price : (b.extractedPrice ?? 0);
      if (filters.sort === "price:asc" || filters.sort === "price") {
        return priceA - priceB;
      }
      if (filters.sort === "price:desc" || filters.sort === "-price") {
        return priceB - priceA;
      }
      return 0;
    });
  }

  if (loading && filtered.length === 0) {
    return (
      <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3 sm:gap-4">
        {[...Array(10)].map((_, i) => (
          <div key={i} className="aspect-square bg-secondary/50 rounded-xl animate-pulse" />
        ))}
      </div>
    );
  }

  if (filtered.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-16 text-center">
        <p className="text-sm font-medium text-foreground">{t("layout.emptyTitle")}</p>
        <p className="max-w-sm text-sm text-muted-foreground">{t("layout.emptyDescription")}</p>
        <button
          type="button"
          onClick={clearFilters}
          className="mt-2 rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
        >
          {t("layout.clearFilters")}
        </button>
      </div>
    );
  }

  return (
    <ProductsGrid
      products={filtered as unknown as CardDetails[]}
      className="grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5"
    />
  );
}
