"use client";

import RecentlyGrid from "@/features/home/components/ui/recentlyGrid";
import { useTranslations } from "next-intl";
import { useCatalogContext } from "@/context/CatalogContext";

export default function Recently() {
  const t = useTranslations("home");
  const { products, loading } = useCatalogContext();

  if (loading || !products || products.length < 6) {
    return null;
  }

  return (
    <div className="w-full px-3 sm:px-6 lg:px-8 mb-5 sm:mb-6">
      <div className="bg-card/40 border border-border/80 rounded-2xl p-3.5 sm:p-5 shadow-xs">
        <div className="flex items-center justify-between pb-2.5 border-b border-border/50">
          <p className="text-foreground font-bold text-sm sm:text-base">
            {t("recentlyViewed.title")}
          </p>
        </div>
        <RecentlyGrid />
      </div>
    </div>
  );
}
