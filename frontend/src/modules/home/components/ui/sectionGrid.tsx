"use client";

import SectionCard from "@/modules/home/components/ui/sectionCard";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { ErrorState } from "@/shared/ui/errorState";
import { localizedName } from "@/shared/lib/localizedName";
import { useLocale } from "next-intl";

export default function SectionGrid() {
  const { categories, errorMsg } = useCatalogContext();
  const locale = useLocale();

  if (categories.length === 0 && errorMsg) {
    return <ErrorState />;
  }

  if (categories.length === 0) {
    return null;
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(180px,1fr))] xl:grid-cols-[repeat(6,minmax(0,1fr))] gap-5">
      {categories.map((cat: any) => {
        const catSlug = cat.slug || cat.id;
        const catName = localizedName(cat.name ?? cat.title, locale);
        return (
          <SectionCard
            key={cat.id || catSlug}
            slug={cat.id || catSlug}
            category={catName}
          />
        );
      })}
    </div>
  );
}
