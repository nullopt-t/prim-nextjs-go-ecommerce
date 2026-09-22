"use client";

import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import FilterCheckbox from "@/features/products/components/ui/filterCheckbox";
import { useCatalogContext } from "@/context/CatalogContext";

export default function SideBarDiscount() {
  const { filters, setFilter } = useCatalogContext();

  const discounts = [
    { id: "discount-1", label: "10% or more", value: 10 },
    { id: "discount-2", label: "25% or more", value: 25 },
    { id: "discount-3", label: "50% or more", value: 50 },
  ];

  return (
    <div className="border-b border-border pb-5">
      <SideBarTitle title="Discount" />
      <div className="flex flex-col gap-2.5">
        {discounts.map((discount) => {
          const isChecked = filters.discount === discount.value;
          return (
            <FilterCheckbox
              key={discount.id}
              labelTxt={discount.label}
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
