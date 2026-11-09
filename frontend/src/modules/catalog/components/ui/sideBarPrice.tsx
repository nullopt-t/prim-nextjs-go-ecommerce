"use client";

import SideBarTitle from "@/modules/catalog/components/ui/sideBarTitle";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { useTranslations } from "next-intl";
import { CustomSelect } from "@/shared/ui/customSelect";

export default function SideBarPrice() {
  const { filters, setFilter } = useCatalogContext();
  const t = useTranslations("common.filters");

  const sortOptions = [
    { value: "", label: t("sortDefault") },
    { value: "price:asc", label: t("sortPriceAsc") },
    { value: "price:desc", label: t("sortPriceDesc") },
  ];

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title={t("price")} />
      
      {/* Min & Max inputs */}
      <div className="grid grid-cols-2 gap-2 mb-3">
        <div className="min-w-0">
          <label className="block text-xs text-muted-foreground mb-1">{t("minPrice")}</label>
          <input
            type="number"
            min={0}
            placeholder="0"
            value={filters.minPrice ?? ""}
            onChange={(e) => {
              const val = e.target.value.trim();
              setFilter("minPrice", val === "" ? null : Math.max(0, Number(val)));
            }}
            className="w-full min-w-0 px-2 py-1.5 text-xs rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand"
          />
        </div>
        <div className="min-w-0">
          <label className="block text-xs text-muted-foreground mb-1">{t("maxPrice")}</label>
          <input
            type="number"
            min={0}
            placeholder="∞"
            value={filters.maxPrice ?? ""}
            onChange={(e) => {
              const val = e.target.value.trim();
              setFilter("maxPrice", val === "" ? null : Math.max(0, Number(val)));
            }}
            className="w-full min-w-0 px-2 py-1.5 text-xs rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand"
          />
        </div>
      </div>

      {/* Sort selection */}
      <div className="mt-3">
        <label className="block text-xs text-muted-foreground mb-1">{t("sort")}</label>
        <CustomSelect
          value={filters.sort ?? ""}
          onChange={(value) => setFilter("sort", value || null)}
          options={sortOptions}
          size="sm"
          className="w-full"
        />
      </div>
    </div>
  );
}
