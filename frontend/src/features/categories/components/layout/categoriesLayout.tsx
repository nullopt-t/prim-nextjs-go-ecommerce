"use client";

import SectionGrid from "@/features/home/components/ui/sectionGrid";
import { useTranslations } from "next-intl";

export default function CategoriesLayout() {
  const t = useTranslations("common");

  return (
    <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 flex flex-col gap-8">
      <div className="flex flex-col gap-2 mb-4">
        <h1 className="text-3xl md:text-4xl font-bold text-foreground">
          {t("nav.categories", { defaultMessage: "Categories" })}
        </h1>
        <p className="text-muted-foreground text-sm md:text-base max-w-2xl">
          {t("categories.description", { defaultMessage: "Explore our diverse range of product categories." })}
        </p>
      </div>

      <SectionGrid />
    </div>
  );
}
