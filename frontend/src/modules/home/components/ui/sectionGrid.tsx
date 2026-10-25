"use client";

import SectionCard from "@/modules/home/components/ui/sectionCard";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";

export default function SectionGrid() {
  const { categories } = useCatalogContext();

  if (categories.length === 0) {
    return null;
  }

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(180px,1fr))] xl:grid-cols-[repeat(6,minmax(0,1fr))] gap-5">
      {categories.map((cat: any) => {
        const catSlug = cat.slug || cat.name?.toLowerCase().replace(/[^a-z0-9]+/g, "-");
        const catName = typeof cat.name === "object" ? (cat.name.en || cat.name.ar) : (cat.name || cat.title || "Category");
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
