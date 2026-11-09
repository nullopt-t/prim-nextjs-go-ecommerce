import Hero from "@/modules/home/components/ui/hero";
import ValuePropsStrip from "@/modules/home/components/ui/valuePropsStrip";
import CategoryShowcaseCards from "@/modules/home/components/ui/categoryShowcaseCards";
import NewArrivalsShelf from "@/modules/home/components/ui/newArrivalsShelf";
import CategoryShelves from "@/modules/home/components/ui/categoryShelves";
import OnSaleShelf from "@/modules/home/components/ui/onSaleShelf";
import BrandStrip from "@/modules/home/components/ui/brandStrip";
import Testimonials from "@/modules/home/components/ui/testimonials";
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

      {/* Trust / reassurance strip */}
      <ValuePropsStrip />

      {/* Category showcase cards */}
      <CategoryShowcaseCards products={products} categories={categories} />

      {/* Newest products */}
      <NewArrivalsShelf products={products} />

      {/* Horizontal scrolling product shelves per category */}
      <CategoryShelves products={products} categories={categories} />

      {/* Discounted products */}
      <OnSaleShelf products={products} />

      {/* Brands + social proof */}
      <BrandStrip products={products} />
      <Testimonials />
    </div>
  );
}
