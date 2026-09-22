"use client";

import { ProductsGrid } from "@/components/ui/productsGrid";
import SectionTitle from "@/features/home/components/ui/sectionTitle";
import { useCatalogContext } from "@/context/CatalogContext";

export default function ProductsSection() {
  const { products, loading } = useCatalogContext();

  // Take first 10 products from live backend catalog
  const displayProducts = products.length > 0 ? products.slice(0, 10) : [];

  return (
    <div>
      <div className="mb-4">
        <SectionTitle title="featuredProducts.title" />
      </div>
      {loading && displayProducts.length === 0 ? (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="aspect-square bg-secondary/50 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : (
        <ProductsGrid
          products={displayProducts}
          className="grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5"
          isWishlist={false}
        />
      )}
    </div>
  );
}
