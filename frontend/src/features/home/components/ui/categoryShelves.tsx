"use client";

import { useMemo, useRef } from "react";
import { Link } from "@/i18n/navigation";
import { ChevronLeft, ChevronRight, ArrowRight } from "lucide-react";
import { useCatalogContext } from "@/context/CatalogContext";
import { ProductsCard } from "@/components/ui/productsCard";

interface CategoryShelfProps {
  category: {
    id: string;
    name: string;
    slug?: string;
  };
  products: any[];
}

function CategoryShelf({ category, products }: CategoryShelfProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const scroll = (direction: "left" | "right") => {
    if (scrollRef.current) {
      const scrollAmount = direction === "left" ? -400 : 400;
      scrollRef.current.scrollBy({ left: scrollAmount, behavior: "smooth" });
    }
  };

  return (
    <div className="bg-card border border-border/80 rounded-2xl p-6 relative group">
      {/* Header with Title and "Explore all" Link */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h2 className="text-xl md:text-2xl font-bold text-foreground">
            {category.name}
          </h2>
          <p className="text-xs md:text-sm text-muted-foreground mt-0.5">
            Top rated items and bestselling gear
          </p>
        </div>

        <Link
          href={`/products?category=${category.id}`}
          className="text-xs md:text-sm font-semibold text-accent-brand hover:underline hover:underline-offset-4 inline-flex items-center gap-1.5 transition-colors"
        >
          <span>See all</span>
          <ArrowRight className="size-4 rtl:rotate-180" />
        </Link>
      </div>

      {/* Navigation Arrows for Carousel */}
      <button
        type="button"
        onClick={() => scroll("left")}
        aria-label="Scroll left"
        className="hidden md:flex absolute left-2 top-1/2 -translate-y-1/2 z-10 size-10 rounded-full bg-background/90 backdrop-blur-sm border border-border shadow-md items-center justify-center text-foreground hover:bg-accent hover:text-accent-brand opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
      >
        <ChevronLeft className="size-5" />
      </button>

      <button
        type="button"
        onClick={() => scroll("right")}
        aria-label="Scroll right"
        className="hidden md:flex absolute right-2 top-1/2 -translate-y-1/2 z-10 size-10 rounded-full bg-background/90 backdrop-blur-sm border border-border shadow-md items-center justify-center text-foreground hover:bg-accent hover:text-accent-brand opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
      >
        <ChevronRight className="size-5" />
      </button>

      {/* Horizontal Carousel */}
      <div
        ref={scrollRef}
        className="flex gap-4 overflow-x-auto no-scrollbar scroll-smooth pb-2"
        style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
      >
        {products.map((product) => (
          <div
            key={product.id || product.slug}
            className="w-[200px] sm:w-[220px] md:w-[240px] shrink-0"
          >
            <ProductsCard cardDetails={product} />
          </div>
        ))}
      </div>
    </div>
  );
}

export default function CategoryShelves() {
  const { categories, products, loading } = useCatalogContext();

  // Create curated rows for each category that has items
  const shelves = useMemo(() => {
    if (!categories || categories.length === 0 || !products || products.length === 0) {
      return [];
    }

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
      .filter((shelf: any) => shelf.products.length > 0);
  }, [categories, products]);

  if (loading && shelves.length === 0) {
    return (
      <div className="flex flex-col gap-8">
        {[...Array(2)].map((_, i) => (
          <div key={i} className="h-72 bg-secondary/30 rounded-2xl animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-10">
      {shelves.map((shelf) => (
        <CategoryShelf
          key={shelf.id}
          category={{ id: shelf.id, name: shelf.name, slug: shelf.slug }}
          products={shelf.products}
        />
      ))}
    </div>
  );
}
