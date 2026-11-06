import Hero from "@/modules/home/components/ui/hero";
import CategoryShowcaseCards from "@/modules/home/components/ui/categoryShowcaseCards";
import CategoryShelves from "@/modules/home/components/ui/categoryShelves";
import { CatalogFeedLoader } from "@/modules/catalog/components/CatalogFeedLoader";

export default function HomeLayout({
  products = [],
  categories = [],
}: {
  products?: import("@/modules/catalog/types").ProductDetailsEntity[];
  categories?: import("@/modules/catalog/types").Category[];
}) {
  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <CatalogFeedLoader />
      <Hero products={products} />

      {/* Category showcase cards */}
      <CategoryShowcaseCards products={products} categories={categories} />

      {/* Horizontal scrolling product shelves per category */}
      <CategoryShelves products={products} categories={categories} />
    </div>
  );
}
