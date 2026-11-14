"use client";

import { useMemo } from "react";
import { Link } from "@/shared/i18n/navigation";
import { ArrowRight } from "lucide-react";

import { ErrorState } from "@/shared/ui/errorState";
import { useTranslations, useLocale } from "next-intl";
import { normalizeMediaUrl } from "@/shared/lib/media";
import { localizedName } from "@/shared/lib/localizedName";
import type { Brand, Category, ProductDetailsEntity } from "@/modules/catalog/types";

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
    <Link
      href={`/products?category=${encodeURIComponent(categoryLinkTarget)}`}
      className="group/card bg-card border border-border hover:border-accent-brand/50 rounded-2xl p-3.5 sm:p-4 flex flex-col justify-between shadow-2xs hover:shadow-sm transition-all text-start"
    >
      <div>
        <h3 className="text-base sm:text-lg font-bold text-foreground mb-3 line-clamp-1 text-start group-hover/card:text-accent-brand transition-colors">
          {category.name}
        </h3>

        <div className="grid grid-cols-2 gap-2.5 mb-3">
          {top4.map((p) => {
            const productTarget = p.slug || p.id;
            const productTitle = typeof p.product === "object"
              ? (locale === "ar" ? p.product?.ar || p.product?.en : p.product?.en || p.product?.ar) || p.title
              : (p.title || "—");

            return (
              <div
                key={productTarget}
                className="flex flex-col text-start"
                title={productTitle}
              >
                <div className="relative aspect-square w-full rounded-xl overflow-hidden bg-secondary/50 border border-border/40 transition-colors">
                  <img
                    src={normalizeMediaUrl(p.img || p.thumbnail)}
                    alt={productTitle}
                    className="w-full h-full object-cover object-center group-hover/card:scale-105 transition-transform duration-300"
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div className="text-xs font-semibold text-accent-brand inline-flex items-center gap-1.5 mt-1 transition-colors">
        <span>{t("seeAll")}</span>
        <ArrowRight className="size-3.5 rtl:rotate-180 group-hover/card:translate-x-1 rtl:group-hover/card:-translate-x-1 transition-transform" />
      </div>
    </Link>
  );
}

export default function CategoryShowcaseCards({
  categories = [],
  products = [],
  loading = false,
  showAll = false,
}: {
  categories?: Category[];
  products?: ProductDetailsEntity[];
  loading?: boolean;
  showAll?: boolean;
}) {
  const locale = useLocale();

  const categoryGroups = useMemo(() => {
    if (!categories || categories.length === 0) {
      return [];
    }

    // Group products by category ID
    const groups = categories
      .map((cat: any) => {
        const catName = localizedName(cat.name, locale);
        const enName = localizedName(cat.name, "en", "");
        const catSlug = (cat.slug || "").toLowerCase();
        const catProducts = (products || []).filter(
          (p: any) =>
            p.categoryId === cat.id ||
            (p.categorySlug && p.categorySlug.toLowerCase() === catSlug) ||
            (typeof p.category === "string"
              ? p.category.toLowerCase() === catName.toLowerCase() ||
                p.category.toLowerCase() === enName.toLowerCase() ||
                (catSlug && p.category.toLowerCase() === catSlug)
              : p.category?.name && (
                  String(p.category.name).toLowerCase() === catName.toLowerCase() ||
                  String(p.category.name).toLowerCase() === enName.toLowerCase()
                ))
        );
        return {
          id: cat.id,
          name: catName,
          slug: cat.slug,
          products: catProducts,
        };
      });

    const validGroups = groups.filter((group: any) => group.products.length >= 2);

    if (showAll) {
      return validGroups;
    }

    return validGroups.slice(0, 6);
  }, [categories, products, locale, showAll]);

  if (loading && categoryGroups.length === 0) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-6 gap-4">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="aspect-square bg-secondary/40 rounded-2xl animate-pulse" />
        ))}
      </div>
    );
  }

  if (!loading && categoryGroups.length === 0 && categories.length === 0) return <ErrorState />;

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
