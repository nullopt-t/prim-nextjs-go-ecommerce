import Hero from "@/modules/home/components/ui/hero";
import DealsBanner from "@/modules/home/components/ui/dealsBanner";
import CategoryShowcaseCards from "@/modules/home/components/ui/categoryShowcaseCards";
import CategoryShelves from "@/modules/home/components/ui/categoryShelves";

export default function HomeLayout() {
  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <Hero />

      {/* Featured daily deals & trust banner */}
      <DealsBanner />

      {/* Category showcase cards */}
      <CategoryShowcaseCards />

      {/* Horizontal scrolling product shelves per category */}
      <CategoryShelves />
    </div>
  );
}
