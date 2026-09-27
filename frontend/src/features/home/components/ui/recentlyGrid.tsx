"use client";

import RecentlyCard from "@/features/home/components/ui/recentlyCard";
import { useCatalogContext } from "@/context/CatalogContext";

export default function RecentlyGrid() {
  const { products } = useCatalogContext();

  // Show 5 items from the catalog as recently viewed
  const items = products.slice(5, 11);

  if (items.length === 0) {
    return null;
  }

  return (
    <div className="flex gap-4 sm:gap-6 mt-4 overflow-x-auto pb-2 scrollbar-none">
      {items.map((item: any) => (
        <div key={item.id || item.slug} className="shrink-0">
          <RecentlyCard
            title={item.title || item.product?.en}
            img={item.img}
            slug={item.slug || item.id}
          />
        </div>
      ))}
    </div>
  );
}
