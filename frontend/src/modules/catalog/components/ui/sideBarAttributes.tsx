"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { ChevronUp } from "lucide-react";
import { useTranslations } from "next-intl";
import FilterCheckbox from "@/modules/catalog/components/ui/filterCheckbox";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { CategoryAttribute } from "@/modules/catalog/types";

interface CollapsibleAttributeGroupProps {
  attr: CategoryAttribute;
  selectedValue?: string;
  onToggle: (key: string, value: string) => void;
}

function CollapsibleAttributeGroup({
  attr,
  selectedValue,
  onToggle,
}: CollapsibleAttributeGroupProps) {
  const t = useTranslations("common.filters");
  const [isOpen, setIsOpen] = useState(true);
  const [isExpanded, setIsExpanded] = useState(false);

  const values = attr.allowedValues ?? [];
  const limit = 6;
  const hasMore = values.length > limit;
  const isSelectedBeyondLimit = Boolean(
    selectedValue && values.indexOf(selectedValue) >= limit,
  );
  const visibleValues = isExpanded || isSelectedBeyondLimit ? values : values.slice(0, limit);

  return (
    <div className="border-b border-border pb-5">
      <button
        type="button"
        onClick={() => setIsOpen((prev) => !prev)}
        aria-expanded={isOpen}
        className="w-full flex items-center justify-between text-start group cursor-pointer select-none mb-2.5"
      >
        <div className="flex items-center gap-2 min-w-0">
          <span className="block font-medium text-foreground uppercase text-xs tracking-wider truncate">
            {attr.label || attr.key}
          </span>
          {selectedValue && (
            <span
              className="shrink-0 size-1.5 rounded-full bg-accent-brand"
              aria-hidden="true"
            />
          )}
        </div>
        <ChevronUp
          className={`size-4 text-muted-foreground group-hover:text-foreground transition-transform duration-200 shrink-0 ${
            isOpen ? "rotate-0" : "rotate-180"
          }`}
        />
      </button>

      {isOpen && (
        <div className="flex flex-col gap-2.5 animate-in fade-in duration-150">
          {visibleValues.map((value) => (
            <FilterCheckbox
              key={value}
              labelTxt={value}
              checked={selectedValue === value}
              onChange={(checked) => {
                if (checked || selectedValue === value) {
                  onToggle(attr.key, value);
                }
              }}
            />
          ))}

          {hasMore && !isSelectedBeyondLimit && (
            <button
              type="button"
              onClick={() => setIsExpanded((prev) => !prev)}
              className="self-start text-xs font-medium text-accent-brand hover:underline pt-1 cursor-pointer"
            >
              {isExpanded ? t("showLess") : `${t("showMore")} (+${values.length - limit})`}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Layered-navigation facet rows for the selected category. Definitions come
 * from the backend (`GET /api/v1/categories/:id/attributes`); only
 * `isFilterable` attributes with allowed values render as collapsible facet groups.
 * Selection is single-valued per key and lives in catalog context.
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
    const param = categoryParam.trim().toLowerCase();
    const matched = (categories ?? []).find((c) => {
      const idMatch = String(c.id).toLowerCase() === param;
      const slugMatch = (c as { slug?: string }).slug?.toLowerCase() === param;
      const name = typeof c.name === "string" ? c.name : (c.name as { en?: string; ar?: string })?.en || "";
      const nameMatch = name.toLowerCase() === param;
      return idMatch || slugMatch || nameMatch;
    });
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
        <CollapsibleAttributeGroup
          key={attr.key}
          attr={attr}
          selectedValue={selected[attr.key]}
          onToggle={toggleValue}
        />
      ))}
    </>
  );
}
