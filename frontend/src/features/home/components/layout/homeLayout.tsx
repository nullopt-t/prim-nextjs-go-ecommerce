import Hero from "@/features/home/components/ui/hero";
import DealsBanner from "@/features/home/components/ui/dealsBanner";
import CategoryShowcaseCards from "@/features/home/components/ui/categoryShowcaseCards";
import CategoryShelves from "@/features/home/components/ui/categoryShelves";

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
