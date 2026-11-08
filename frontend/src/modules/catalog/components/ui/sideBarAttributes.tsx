"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import SideBarTitle from "@/modules/catalog/components/ui/sideBarTitle";
import FilterCheckbox from "@/modules/catalog/components/ui/filterCheckbox";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { CategoryAttribute } from "@/modules/catalog/types";

/**
 * Layered-navigation facet rows for the selected category. Definitions come
 * from the backend (`GET /api/v1/categories/:id/attributes`); only
 * `isFilterable` attributes with allowed values render. Selection is
 * single-valued per key (backend matches one key:value per attribute) and
 * lives in the catalog context so Content can mirror it into the URL and
 * re-fetch the server-filtered list.
 */
export default function SideBarAttributes() {
  const { categories, filters, setFilter } = useCatalogContext();
  const searchParams = useSearchParams();
  const [attributes, setAttributes] = useState<CategoryAttribute[]>([]);

  // Resolve the URL category slug/id to a category id for the attributes API.
  const categoryParam = searchParams.get("category");

  useEffect(() => {
    let isMounted = true;
    if (!categoryParam || categoryParam.toLowerCase() === "all") {
      setAttributes([]);
      return;
    }
    const matched = (categories ?? []).find(
      (c) => String(c.id) === categoryParam || (c as { slug?: string }).slug === categoryParam,
    );
    const categoryId = matched ? String(matched.id) : categoryParam;
    catalogService
      .getCategoryAttributes(categoryId)
      .then((res) => {
        if (!isMounted) return;
        const list = (res?.data ?? []) as unknown as CategoryAttribute[];
        setAttributes(
          (Array.isArray(list) ? list : [])
            .filter((a) => a.isFilterable && Array.isArray(a.allowedValues) && a.allowedValues.length > 0)
            .sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0)),
        );
      })
      .catch(() => {
        if (isMounted) setAttributes([]);
      });
    return () => {
      isMounted = false;
    };
  }, [categoryParam, categories]);

  if (attributes.length === 0) return null;

  const selected = filters.attributes ?? {};

  const toggleValue = (key: string, value: string) => {
    const next = { ...selected };
    if (next[key] === value) delete next[key];
    else next[key] = value;
    setFilter("attributes", next);
  };

  return (
    <>
      {attributes.map((attr) => (
        <div key={attr.key} className="border-b border-border pb-5">
          <SideBarTitle title={attr.label || attr.key} />
          <div className="flex flex-col gap-2.5">
            {attr.allowedValues!.map((value) => (
              <FilterCheckbox
                key={value}
                labelTxt={value}
                checked={selected[attr.key] === value}
                onChange={(checked) => {
                  if (checked) toggleValue(attr.key, value);
                  else if (selected[attr.key] === value) toggleValue(attr.key, value);
                }}
              />
            ))}
          </div>
        </div>
      ))}
    </>
  );
}
