"use client";

import CategoryShowcaseCards from "@/modules/home/components/ui/categoryShowcaseCards";
import { useTranslations } from "next-intl";
import type { Category, ProductDetailsEntity } from "@/modules/catalog/types";

export default function CategoriesLayout({
  categories,
  products,
}: {
  categories?: Category[];
  products?: ProductDetailsEntity[];
}) {
  const t = useTranslations("shop");

  return (
    <div className="w-full py-4 sm:py-8 flex flex-col gap-8">
      <div className="flex flex-col gap-2 mb-4">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("categories.title")}
        </h1>
        <p className="text-muted-foreground text-sm md:text-base max-w-2xl">
          {t("categories.description")}
        </p>
      </div>

      <CategoryShowcaseCards categories={categories} products={products} showAll />
    </div>
  );
}
