"use client";

import SideBar from "@/features/products/components/ui/sideBar";
import Content from "@/features/products/components/ui/content";
import { Title } from "@/components/ui/title";
import { useCatalogContext } from "@/context/CatalogContext";

import { useSearchParams } from "next/navigation";

interface ProductsLayoutProps {
  category?: string;
}

export default function ProductsLayout({ category = "All" }: ProductsLayoutProps) {
  const { categories } = useCatalogContext();
  const searchParams = useSearchParams();
  const searchQuery = searchParams.get("q");

  let displayTitle = "All Products";
  let displaySubtitle = "Showing all products";

  if (searchQuery) {
    displayTitle = `Search: "${searchQuery}"`;
    displaySubtitle = "Matching products";
  } else if (category && category.toLowerCase() !== "all") {
    const matchedCategory = categories.find((c: any) => c.id === category || c.slug === category);
    if (matchedCategory) {
      displayTitle = typeof matchedCategory.name === "object"
        ? (matchedCategory.name.en || matchedCategory.name.ar)
        : (matchedCategory.name || category);
    } else {
      displayTitle = category.replace(/-/g, " ");
    }
    displaySubtitle = `Showing products in ${displayTitle}`;
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
          <Content category={category} />
        </div>
      </div>
    </div>
  );
}
