"use client";

import { useMemo } from "react";
import { Link } from "@/i18n/navigation";
import { ArrowRight } from "lucide-react";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations, useLocale } from "next-intl";

interface CategoryShowcaseCardProps {
  category: {
    id: string;
    name: string;
    slug?: string;
  };
  products: any[];
}

function CategoryShowcaseCard({ category, products }: CategoryShowcaseCardProps) {
  const t = useTranslations("home.categories");
  const locale = useLocale();
  // Take up to 4 items with images for this category
  const top4 = products.slice(0, 4);
  const categoryLinkTarget = category.slug || category.id;

  return (
    <div className="bg-card border border-border hover:border-accent-brand/50 rounded-2xl p-3.5 sm:p-4 flex flex-col justify-between shadow-2xs hover:shadow-sm transition-all">
      <div>
        <h3 className="text-base sm:text-lg font-bold text-foreground mb-3 line-clamp-1 text-left rtl:text-right">
          {category.name}
        </h3>

        {top4.length >= 4 ? (
          <div className="grid grid-cols-2 gap-2.5 mb-3">
            {top4.map((p) => {
              const productTarget = p.slug || p.id;
              const productTitle = typeof p.product === "object"
                ? (locale === "ar" ? p.product?.ar || p.product?.en : p.product?.en || p.product?.ar) || p.title
                : (p.title || "Product");

              return (
                <Link
                  key={productTarget}
                  href={`/products/${productTarget}`}
                  className="group flex flex-col text-left rtl:text-right"
                >
                  <div className="relative aspect-square w-full rounded-xl overflow-hidden bg-secondary/50 border border-border/40 group-hover:border-accent-brand/60 transition-colors">
                    <img
                      src={p.img || p.thumbnail || "/placeholder-product.png"}
                      alt={productTitle}
                      className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-300"
                    />
                  </div>
                  <p className="text-xs font-medium text-foreground group-hover:text-accent-brand mt-1.5 line-clamp-1 w-full transition-colors">
                    {productTitle}
                  </p>
                  {p.price !== undefined && (
                    <span className="text-[11px] font-semibold text-muted-foreground group-hover:text-foreground">
                      ${typeof p.price === "number" ? p.price.toFixed(2) : p.price}
                    </span>
                  )}
                </Link>
              );
            })}
          </div>
        ) : top4.length > 0 ? (
          (() => {
            const productTarget = top4[0].slug || top4[0].id;
            const productTitle = typeof top4[0].product === "object"
              ? (locale === "ar" ? top4[0].product?.ar || top4[0].product?.en : top4[0].product?.en || top4[0].product?.ar) || top4[0].title
              : (top4[0].title || "Product");

            return (
              <Link
                href={`/products/${productTarget}`}
                className="group block mb-3 text-left rtl:text-right"
              >
                <div className="relative aspect-4/3 w-full rounded-xl overflow-hidden bg-secondary/50 border border-border/40 group-hover:border-accent-brand/60 transition-colors">
                  <img
                    src={top4[0].img || top4[0].thumbnail || "/placeholder-product.png"}
                    alt={productTitle}
                    className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-300"
                  />
                </div>
                <div className="flex items-center justify-between mt-2">
                  <p className="text-xs sm:text-sm font-semibold text-foreground group-hover:text-accent-brand line-clamp-1 transition-colors">
                    {productTitle}
                  </p>
                  {top4[0].price !== undefined && (
                    <span className="text-xs font-bold text-foreground shrink-0 ltr:ml-2 rtl:mr-2">
                      ${typeof top4[0].price === "number" ? top4[0].price.toFixed(2) : top4[0].price}
                    </span>
                  )}
                </div>
              </Link>
            );
          })()
        ) : (
          <div className="aspect-square bg-secondary/40 rounded-xl mb-3 flex items-center justify-center text-xs text-muted-foreground">
            Explore Category
          </div>
        )}
      </div>

      <Link
        href={`/products?category=${encodeURIComponent(categoryLinkTarget)}`}
        className="text-xs font-semibold text-accent-brand hover:underline hover:underline-offset-4 inline-flex items-center gap-1.5 mt-1 transition-colors"
      >
        <span>{t("seeAll")}</span>
        <ArrowRight className="size-3.5 rtl:rotate-180" />
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
      .slice(0, 6);
  }, [categories, products]);

  if (loading && categoryGroups.length === 0) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-6 gap-4">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="aspect-square bg-secondary/40 rounded-2xl animate-pulse" />
        ))}
      </div>
    );
  }

  if (categoryGroups.length === 0) return null;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-6 gap-4">
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
