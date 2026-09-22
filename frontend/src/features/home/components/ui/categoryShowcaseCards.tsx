"use client";

import { useMemo } from "react";
import { Link } from "@/i18n/navigation";
import { useCatalogContext } from "@/context/CatalogContext";

interface CategoryShowcaseCardProps {
  category: {
    id: string;
    name: string;
    slug?: string;
  };
  products: any[];
}

function CategoryShowcaseCard({ category, products }: CategoryShowcaseCardProps) {
  // Take up to 4 items with images for this category
  const top4 = products.slice(0, 4);

  return (
    <div className="bg-card border border-border/80 rounded-2xl p-5 flex flex-col justify-between hover:shadow-lg transition-shadow">
      <div>
        <h3 className="text-lg font-bold text-foreground mb-4 line-clamp-1">
          {category.name}
        </h3>

        {top4.length >= 4 ? (
          <div className="grid grid-cols-2 gap-3 mb-4">
            {top4.map((p) => (
              <Link
                key={p.id || p.slug}
                href={`/products/${p.slug}`}
                className="group flex flex-col items-center"
              >
                <div className="relative aspect-square w-full rounded-xl overflow-hidden bg-secondary/50 border border-border/40 group-hover:border-accent-brand/60 transition-colors">
                  <img
                    src={p.img || p.thumbnail || "/placeholder-product.png"}
                    alt={p.title || "Product"}
                    className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-300"
                  />
                </div>
                <p className="text-xs text-muted-foreground group-hover:text-foreground mt-1.5 line-clamp-1 w-full text-center">
                  {p.title}
                </p>
              </Link>
            ))}
          </div>
        ) : top4.length > 0 ? (
          <Link
            href={`/products/${top4[0].slug}`}
            className="group block mb-4"
          >
            <div className="relative aspect-4/3 w-full rounded-xl overflow-hidden bg-secondary/50 border border-border/40 group-hover:border-accent-brand/60 transition-colors">
              <img
                src={top4[0].img || top4[0].thumbnail || "/placeholder-product.png"}
                alt={top4[0].title || "Product"}
                className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-300"
              />
            </div>
            <p className="text-sm font-medium text-foreground mt-2 line-clamp-1">
              {top4[0].title}
            </p>
          </Link>
        ) : (
          <div className="aspect-square bg-secondary/40 rounded-xl mb-4 flex items-center justify-center text-xs text-muted-foreground">
            Explore Category
          </div>
        )}
      </div>

      <Link
        href={`/products?category=${category.id}`}
        className="text-xs font-semibold text-accent-brand hover:underline hover:underline-offset-4 inline-flex items-center gap-1 mt-2"
      >
        <span>Shop all in {category.name}</span>
        <span aria-hidden="true">&rarr;</span>
      </Link>
    </div>
  );
}

export default function CategoryShowcaseCards() {
  const { categories, products, loading } = useCatalogContext();

  const categoryGroups = useMemo(() => {
    if (!categories || categories.length === 0 || !products || products.length === 0) {
      return [];
    }

    // Group products by category ID
    return categories
      .map((cat: any) => {
        const catName = typeof cat.name === "object" ? (cat.name.en || cat.name.ar) : (cat.name || "Category");
        const catProducts = products.filter(
          (p: any) => p.categoryId === cat.id || (p.category && p.category.toLowerCase() === catName.toLowerCase())
        );
        return {
          id: cat.id,
          name: catName,
          slug: cat.slug,
          products: catProducts,
        };
      })
      .filter((group: any) => group.products.length > 0)
      .slice(0, 4);
  }, [categories, products]);

  if (loading && categoryGroups.length === 0) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="aspect-square bg-secondary/40 rounded-2xl animate-pulse" />
        ))}
      </div>
    );
  }

  if (categoryGroups.length === 0) return null;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
      {categoryGroups.map((group) => (
        <CategoryShowcaseCard
          key={group.id}
          category={{ id: group.id, name: group.name, slug: group.slug }}
          products={group.products}
        />
      ))}
    </div>
  );
}
