import SectionTitle from "@/features/home/components/ui/sectionTitle";
import SectionGrid from "@/features/home/components/ui/sectionGrid";
import ProductsSection from "@/features/home/components/ui/productsSection";
import Hero from "@/features/home/components/ui/hero";

export default function HomeLayout() {
	return (
		<div className="flex flex-col gap-10">
			<div className="bg-footer">
				<Hero />
			</div>
			<div className="">
				<div className="mb-2.5">
					<SectionTitle title="categories" />
				</div>
				<SectionGrid />
			</div>
			<ProductsSection />
		</div>
	);
}
