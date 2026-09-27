"use client";

import { SlidersHorizontal } from "lucide-react";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations } from "next-intl";

export default function SideBarFilters() {
  const { clearFilters } = useCatalogContext();
  const t = useTranslations("common.filters");

  return (
    <div className="flex justify-between items-center text-txt-sm md:text-txt-md text-muted-foreground">
      <div className="flex items-center gap-2 font-medium text-foreground">
        <SlidersHorizontal className="size-4" />
        <span>{t("title")}</span>
      </div>
      <button
        type="button"
        onClick={clearFilters}
        className="cursor-pointer text-accent-brand font-medium hover:underline text-xs"
      >
        {t("clearAll")}
      </button>
    </div>
  );
}
