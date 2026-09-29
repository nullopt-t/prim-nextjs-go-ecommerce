"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { ProductsGrid } from "@/components/ui/productsGrid";
import { CardDetails } from "@/components/ui/productsCard";
import { useCatalogContext, normalizeProduct } from "@/context/CatalogContext";
import { catalogService } from "@/services/catalog";

interface ContentProps {
  products?: CardDetails[];
  category?: string;
}

export default function Content({ products, category }: ContentProps) {
  const { products: contextProducts, loading: contextLoading, filters } = useCatalogContext();
  const [categoryProducts, setCategoryProducts] = useState<any[] | null>(null);
  const [catLoading, setCatLoading] = useState(false);
  const searchParams = useSearchParams();
  const searchQuery = (searchParams.get("q") || "").trim().toLowerCase();

  useEffect(() => {
    // If no category or "all", we use the full context products
    if (!category || category.toLowerCase() === "all" || category.toLowerCase() === "all categories") {
      setCategoryProducts(null);
      return;
    }

    let isMounted = true;
    setCatLoading(true);

    catalogService
      .getProducts({ pageSize: 100, category })
      .then((res: any) => {
        if (!isMounted) return;
        const rawList = Array.isArray(res) ? res : Array.isArray(res?.data) ? res.data : [];
        setCategoryProducts(rawList.map(normalizeProduct));
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
  }, [category]);

  const activeProducts = categoryProducts !== null ? categoryProducts : (products || contextProducts);
  const loading = catLoading || (categoryProducts === null && contextLoading);

  const filtered = activeProducts.filter((p: any) => {
    // Search query filter (matches title, brand, description or tags)
    if (searchQuery) {
      const title = (p.title || p.product?.en || p.product?.ar || "").toLowerCase();
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

    return true;
  });

  if (loading && filtered.length === 0) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
        {[...Array(10)].map((_, i) => (
          <div key={i} className="aspect-square bg-secondary/50 rounded-xl animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <ProductsGrid
      products={filtered}
      className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5"
    />
  );
}
