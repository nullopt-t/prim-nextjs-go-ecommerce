"use client";

import SideBar from "@/modules/catalog/components/ui/sideBar";
import Content from "@/modules/catalog/components/ui/content";
import { Title } from "@/shared/ui/title";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { useTranslations, useLocale } from "next-intl";
import { localizedName } from "@/shared/lib/localizedName";
import type { Category, ProductDetailsEntity } from "@/modules/catalog/types";

import { useSearchParams, useRouter, usePathname } from "next/navigation";

interface ProductsLayoutProps {
  category?: string;
  products?: ProductDetailsEntity[];
  /** Overrides the computed heading (e.g. brand pages). */
  heading?: string;
  page?: number;
  totalItems?: number;
  totalPages?: number;
}

export default function ProductsLayout({
  category = "All",
  products,
  heading,
  page = 1,
  totalItems,
  totalPages = 1,
}: ProductsLayoutProps) {
  const t = useTranslations("products");
  const locale = useLocale();
  const { categories } = useCatalogContext();
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const searchQuery = searchParams.get("q");

  const goToPage = (target: number) => {
    const next = new URLSearchParams(searchParams.toString());
    if (target <= 1) next.delete("page");
    else next.set("page", String(target));
    router.push(`${pathname}?${next.toString()}`);
  };

  let displayTitle = t("layout.allProducts");
  let displaySubtitle = t("layout.showingAll");

  if (searchQuery) {
    displayTitle = t("layout.searchTitle", { query: searchQuery });
    displaySubtitle = t("layout.matching");
  } else if (category && category.toLowerCase() !== "all") {
    const matchedCategory = categories.find((c: Category) => c.id === category || c.slug === category);
    if (matchedCategory) {
      const name = matchedCategory.name as string | { en?: string; ar?: string };
      displayTitle = localizedName(name, locale, category);
    } else {
      displayTitle = category.replace(/-/g, " ");
    }
    displaySubtitle = t("layout.showingIn", { category: displayTitle });
  }

  if (heading) {
    displayTitle = heading;
    displaySubtitle = t("layout.showingAll");
  }
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <Title
          title={displayTitle}
          subtitle={
            totalItems !== undefined
              ? t("layout.pagination.results", { count: totalItems })
              : displaySubtitle
          }
          className="mb-0"
        />
      </div>
      <div className="relative flex flex-col md:flex-row gap-5 lg:gap-10">
        <div className="shrink-0">
          <SideBar />
        </div>
        <div className="min-w-0 flex-1">
          <Content category={category} products={products} />
          {totalPages > 1 && (
            <div className="mt-6 flex flex-wrap items-center justify-between gap-3">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => goToPage(page - 1)}
                className="min-h-11 shrink-0 whitespace-nowrap rounded-md border border-border px-3 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-50"
              >
                {t("layout.pagination.previous")}
              </button>
              <span className="text-xs text-muted-foreground text-center min-w-0">
                {t("layout.pagination.pageOf", { page, total: totalPages })}
              </span>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => goToPage(page + 1)}
                className="min-h-11 shrink-0 whitespace-nowrap rounded-md border border-border px-3 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-50"
              >
                {t("layout.pagination.next")}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
