"use client";

import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import { Toggle } from "@/components/ui/Toggle";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations } from "next-intl";

export default function SideBarAvailability() {
  const { filters, setFilter } = useCatalogContext();
  const t = useTranslations("common.filters");

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title={t("availability")} />
      <div className="flex justify-between items-center">
        <span className="text-muted-foreground text-txt-sm md:text-txt-md">
          {t("inStockOnly")}
        </span>
        <Toggle
          isEnabled={filters.inStockOnly}
          onChange={(e) => setFilter("inStockOnly", e.target.checked)}
        />
      </div>
    </div>
  );
}
