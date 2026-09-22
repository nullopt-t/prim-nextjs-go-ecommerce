"use client";

import SectionCard from "@/features/home/components/ui/sectionCard";
import { useCatalogContext } from "@/context/CatalogContext";

const defaultCategories = [
  { id: "cat-1", slug: "electronics", name: "Electronics" },
  { id: "cat-2", slug: "laptops", name: "Laptops" },
  { id: "cat-3", slug: "smartphones", name: "Smartphones" },
  { id: "cat-4", slug: "audio-headphones", name: "Audio & Headphones" },
  { id: "cat-5", slug: "tablets-wearables", name: "Tablets & Wearables" },
  { id: "cat-6", slug: "computer-accessories", name: "Computer Accessories" },
  { id: "cat-7", slug: "gaming-consoles", name: "Gaming Consoles" },
  { id: "cat-8", slug: "apparel", name: "Apparel" },
];

export default function SectionGrid() {
  const { categories, loading } = useCatalogContext();

  const list = categories.length > 0 ? categories : defaultCategories;

  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(100px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(150px,1fr))] gap-5">
      {list.map((cat: any) => {
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
