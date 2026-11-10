"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Search } from "lucide-react";
import { Link } from "@/shared/i18n/navigation";
import { LazyImage } from "@/shared/ui/lazyImage";

export interface BrandDirectoryItem {
  id: string;
  name: string;
  slug: string;
  logoUrl?: string;
  productCount?: number;
  categories: string[];
}

export function BrandsDirectory({ brands }: { brands: BrandDirectoryItem[] }) {
  const t = useTranslations("shop");
  const [query, setQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState<string | null>(null);

  // Tabs come from the categories the brands actually sell, so the filter can
  // never offer an empty bucket.
  const categories = useMemo(() => {
    const unique = new Set<string>();
    for (const brand of brands) {
      for (const category of brand.categories) unique.add(category);
    }
    return [...unique].sort((a, b) => a.localeCompare(b));
  }, [brands]);

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return brands.filter((brand) => {
      const matchesCategory =
        activeCategory === null || brand.categories.includes(activeCategory);
      const matchesQuery =
        needle === "" || brand.name.toLowerCase().includes(needle);
      return matchesCategory && matchesQuery;
    });
  }, [activeCategory, brands, query]);

  return (
    <div className="mt-6">
      <div className="flex flex-col gap-4">
        <label className="relative block">
          <span className="sr-only">{t("brands.searchPlaceholder")}</span>
          <Search
            className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden="true"
          />
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={t("brands.searchPlaceholder")}
            className="min-h-11 w-full rounded-xl border border-border bg-background py-2.5 pe-3 ps-9 text-sm text-foreground transition-colors placeholder:text-muted-foreground focus:border-accent-brand focus:outline-none focus:ring-2 focus:ring-accent-brand/30"
          />
        </label>

        {categories.length > 1 && (
          <div
            role="group"
            aria-label={t("brands.filterLabel")}
            className="scrollbar-none flex gap-2 overflow-x-auto pb-1"
          >
            <button
              type="button"
              onClick={() => setActiveCategory(null)}
              aria-pressed={activeCategory === null}
              className={`shrink-0 cursor-pointer rounded-full border px-3.5 py-1.5 text-xs font-semibold transition-colors ${
                activeCategory === null
                  ? "border-foreground bg-foreground text-background"
                  : "border-border bg-background text-muted-foreground hover:border-foreground/40 hover:text-foreground"
              }`}
            >
              {t("brands.all")}
            </button>
            {categories.map((category) => (
              <button
                key={category}
                type="button"
                onClick={() => setActiveCategory(category)}
                aria-pressed={activeCategory === category}
                className={`shrink-0 cursor-pointer rounded-full border px-3.5 py-1.5 text-xs font-semibold transition-colors ${
                  activeCategory === category
                    ? "border-foreground bg-foreground text-background"
                    : "border-border bg-background text-muted-foreground hover:border-foreground/40 hover:text-foreground"
                }`}
              >
                {category}
              </button>
            ))}
          </div>
        )}
      </div>

      {filtered.length === 0 ? (
        <div className="mt-6 rounded-lg border border-border bg-background px-4 py-12 text-center">
          <p className="text-sm font-medium text-foreground">{t("brands.noResults")}</p>
          <p className="mt-1 text-sm text-muted-foreground">{t("brands.noResultsHint")}</p>
        </div>
      ) : (
        <ul className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4">
          {filtered.map((brand) => (
            <li key={brand.id}>
              <Link
                href={`/brands/${brand.slug}`}
                className="group flex h-full flex-col rounded-2xl border border-border bg-card p-3 shadow-sm transition-all hover:-translate-y-0.5 hover:border-accent-brand/50 hover:shadow-md sm:p-4"
              >
                <div className="flex h-16 w-full items-center justify-center overflow-hidden rounded-xl border border-border/40 bg-white p-3 dark:bg-white/95">
                  {brand.logoUrl ? (
                    <LazyImage
                      src={brand.logoUrl}
                      alt={brand.name}
                      containerClassName="flex h-full w-full items-center justify-center"
                      imageClassName="max-h-12 max-w-[130px] object-contain"
                    />
                  ) : (
                    <span className="text-2xl font-bold tracking-tight text-neutral-800">
                      {brand.name.slice(0, 1).toUpperCase()}
                    </span>
                  )}
                </div>

                <span dir="auto" className="mt-3 block truncate text-sm font-semibold text-foreground transition-colors group-hover:text-accent-brand">
                  {brand.name}
                </span>


                {brand.categories.length > 0 && (
                  <span className="mt-auto flex flex-wrap gap-1 pt-3">
                    {brand.categories.slice(0, 3).map((category) => (
                      <span
                        key={category}
                        className="max-w-full truncate rounded-full bg-secondary px-2 py-0.5 text-[11px] font-medium text-muted-foreground"
                      >
                        {category}
                      </span>
                    ))}
                    {brand.categories.length > 3 && (
                      <span className="rounded-full bg-secondary px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
                        +{brand.categories.length - 3}
                      </span>
                    )}
                  </span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
