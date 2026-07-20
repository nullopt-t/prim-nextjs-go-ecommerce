import SectionTitle from "@/features/home/components/ui/sectionTitle";
import SectionGrid from "@/features/home/components/ui/sectionGrid";
import ProductsSection from "@/features/home/components/ui/productsSection";

export default function Section() {
	return (
		<div className="flex flex-col gap-10">
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
