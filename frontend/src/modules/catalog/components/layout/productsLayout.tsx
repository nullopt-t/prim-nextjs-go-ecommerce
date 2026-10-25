"use client";

import SideBar from "@/modules/catalog/components/ui/sideBar";
import Content from "@/modules/catalog/components/ui/content";
import { Title } from "@/shared/ui/title";
import { useCatalogContext } from "@/modules/catalog/model/catalogContext";
import { useTranslations } from "next-intl";

import { useSearchParams } from "next/navigation";

interface ProductsLayoutProps {
  category?: string;
  products?: any[];
}

export default function ProductsLayout({ category = "All", products }: ProductsLayoutProps) {
  const t = useTranslations("products");
  const { categories } = useCatalogContext();
  const searchParams = useSearchParams();
  const searchQuery = searchParams.get("q");

  let displayTitle = t("layout.allProducts");
  let displaySubtitle = t("layout.showingAll");

  if (searchQuery) {
    displayTitle = t("layout.searchTitle", { query: searchQuery });
    displaySubtitle = t("layout.matching");
  } else if (category && category.toLowerCase() !== "all") {
    const matchedCategory = categories.find((c: any) => c.id === category || c.slug === category);
    if (matchedCategory) {
      displayTitle = typeof matchedCategory.name === "object"
        ? (matchedCategory.name.en || matchedCategory.name.ar)
        : (matchedCategory.name || category);
    } else {
      displayTitle = category.replace(/-/g, " ");
    }
    displaySubtitle = t("layout.showingIn", { category: displayTitle });
  }
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <Title
          title={displayTitle}
          subtitle={displaySubtitle}
          className="mb-0"
        />
      </div>
      <div className="relative flex flex-col md:flex-row gap-5 lg:gap-10">
        <div className="shrink-0">
          <SideBar />
        </div>
        <div className="flex-1">
          <Content category={category} products={products} />
        </div>
      </div>
    </div>
  );
}
