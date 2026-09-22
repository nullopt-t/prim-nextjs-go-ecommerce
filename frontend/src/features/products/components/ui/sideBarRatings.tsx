"use client";

import { Stars } from "@/components/ui/stars";
import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import { useCatalogContext } from "@/context/CatalogContext";

export default function SideBarRatings() {
  const { filters, setFilter } = useCatalogContext();

  const rates = [
    { id: "stars-5", stars: 5 },
    { id: "stars-4", stars: 4 },
    { id: "stars-3", stars: 3 },
  ];

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title="Ratings" />
      <div className="flex flex-col gap-2.5">
        {rates.map((value) => {
          const isSelected = filters.rating === value.stars;
          return (
            <button
              type="button"
              key={value.id}
              onClick={() => setFilter("rating", isSelected ? null : value.stars)}
              className={`flex gap-2.5 items-center group cursor-pointer text-left py-1 px-1.5 rounded-lg transition-colors ${
                isSelected ? "bg-accent-brand/10 text-accent-brand font-semibold" : "hover:bg-secondary"
              }`}
            >
              <Stars starsNum={value.stars} />
              <span className={`text-xs ${isSelected ? "text-accent-brand" : "text-muted-foreground group-hover:text-foreground"}`}>
                &amp; up
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
