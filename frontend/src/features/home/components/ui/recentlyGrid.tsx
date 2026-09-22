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
    <div className="flex flex-wrap gap-5 mt-5">
      {items.map((item: any) => (
        <RecentlyCard
          key={item.id || item.slug}
          title={item.title || item.product?.en}
          img={item.img}
          slug={item.slug || item.id}
        />
      ))}
    </div>
  );
}
