"use client";

import { SlidersHorizontal } from "lucide-react";
import { useCatalogContext } from "@/context/CatalogContext";

export default function SideBarFilters() {
  const { clearFilters } = useCatalogContext();

  return (
    <div className="flex justify-between items-center text-txt-sm md:text-txt-md text-muted-foreground">
      <div className="flex items-center gap-2 font-medium text-foreground">
        <SlidersHorizontal className="size-4" />
        <span>Filters</span>
      </div>
      <button
        type="button"
        onClick={clearFilters}
        className="cursor-pointer text-accent-brand font-medium hover:underline text-xs"
      >
        Clear all
      </button>
    </div>
  );
}
