"use client";

import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import FilterCheckbox from "@/features/products/components/ui/filterCheckbox";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations } from "next-intl";

export default function SideBarDiscount() {
  const { filters, setFilter } = useCatalogContext();
  const t = useTranslations("common.filters");

  const discounts = [
    { id: "discount-1", value: 10 },
    { id: "discount-2", value: 25 },
    { id: "discount-3", value: 50 },
  ];

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title={t("discount")} />
      <div className="flex flex-col gap-2.5">
        {discounts.map((discount) => {
          const isChecked = filters.discount === discount.value;
          return (
            <FilterCheckbox
              key={discount.id}
              labelTxt={t("discountMore", { percent: discount.value })}
              checked={isChecked}
              onChange={(checked) => {
                setFilter("discount", checked ? discount.value : null);
              }}
            />
          );
        })}
      </div>
    </div>
  );
}
