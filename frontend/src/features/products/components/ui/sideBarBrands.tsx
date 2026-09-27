"use client";

import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import FilterCheckbox from "@/features/products/components/ui/filterCheckbox";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations } from "next-intl";

export default function SideBarBrands() {
  const { brands, filters, setFilter } = useCatalogContext();
  const t = useTranslations("common.filters");

  const fallbackBrands = [
    { id: "brand-1", name: "Apple" },
    { id: "brand-2", name: "Sony" },
    { id: "brand-3", name: "Samsung" },
    { id: "brand-4", name: "Bose" },
    { id: "brand-5", name: "Logitech" },
  ];

  const brandList = brands && brands.length > 0 ? brands : fallbackBrands;

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title={t("brands")} />
      <div className="flex flex-col gap-2.5">
        {brandList.slice(0, 6).map((brand: any) => {
          const brandName = brand.name || brand.brand || "Brand";
          const isSelected = filters.brand?.toLowerCase() === brandName.toLowerCase();
          return (
            <FilterCheckbox
              key={brand.id || brandName}
              labelTxt={brandName}
              checked={isSelected}
              onChange={(checked) => {
                setFilter("brand", checked ? brandName : null);
              }}
            />
          );
        })}
      </div>
    </div>
  );
}
