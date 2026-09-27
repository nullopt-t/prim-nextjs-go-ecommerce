"use client";

import { ChevronUp } from "lucide-react";
import { useState, useRef, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import SideBarTitle from "@/features/products/components/ui/sideBarTitle";
import { useCatalogContext } from "@/context/CatalogContext";
import { Link } from "@/i18n/navigation";
import { useTranslations } from "next-intl";

export default function SideBarCategories() {
  const t = useTranslations("common.filters");
  const catRef = useRef<HTMLUListElement>(null);
  const [height, setHeight] = useState<number | undefined>(undefined);
  const [isOpen, setIsOpen] = useState(true);
  const { categories } = useCatalogContext();
  const searchParams = useSearchParams();
  const currentCategory = searchParams.get("category");

  useEffect(() => {
    if (catRef.current) {
      setHeight(isOpen ? catRef.current.scrollHeight : 0);
    }
  }, [isOpen]);

  return (
    <div className="border-b border-border pb-5">
      <div className="flex justify-between items-center">
        <SideBarTitle title={t("category")} />
        <button
          type="button"
          onClick={() => setIsOpen((prev) => !prev)}
          className="cursor-pointer text-muted-foreground hover:text-foreground"
        >
          <ChevronUp
            className={`size-4 transition-transform duration-300 ${
              isOpen ? "rotate-0" : "rotate-180"
            }`}
          />
        </button>
      </div>

      <div
        style={{ height: height !== undefined ? `${height}px` : "auto" }}
        className="transition-[height] duration-300 ease-in-out overflow-hidden"
      >
        <ul ref={catRef} className="mt-3 space-y-2">
          <li>
            <Link
              href="/products"
              className={`text-txt-sm md:text-txt-md cursor-pointer transition-colors block ${
                !currentCategory ? "text-accent-brand font-semibold" : "text-muted-foreground hover:text-accent-brand"
              }`}
            >
              {t("allCategories")}
            </Link>
          </li>
          {categories.map((category: any) => {
            const catName = typeof category.name === "object"
              ? (category.name.en || category.name.ar)
              : (category.name || category.title);
            const catSlug = category.slug || catName.toLowerCase().replace(/[^a-z0-9]+/g, "-");

            const categoryId = category.id;
            const isActive = currentCategory === categoryId || currentCategory === catSlug;
            return (
              <li key={categoryId || catSlug}>
                <Link
                  href={`/products?category=${categoryId || catSlug}`}
                  className={`text-txt-sm md:text-txt-md cursor-pointer transition-colors block ${
                    isActive ? "text-accent-brand font-semibold" : "text-muted-foreground hover:text-accent-brand"
                  }`}
                >
                  {catName}
                </Link>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
