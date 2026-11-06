import Hero from "@/modules/home/components/ui/hero";
import CategoryShowcaseCards from "@/modules/home/components/ui/categoryShowcaseCards";
import CategoryShelves from "@/modules/home/components/ui/categoryShelves";
import { CatalogFeedLoader } from "@/modules/catalog/components/CatalogFeedLoader";

export default function HomeLayout() {
  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <CatalogFeedLoader />
      <Hero />

      {/* Category showcase cards */}
      <CategoryShowcaseCards />

      {/* Horizontal scrolling product shelves per category */}
      <CategoryShelves />
    </div>
  );
}
