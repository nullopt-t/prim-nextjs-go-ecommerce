import SectionTitle from "@/features/home/components/ui/sectionTitle";
import SectionGrid from "@/features/home/components/ui/sectionGrid";
import ProductsSection from "@/features/home/components/ui/productsSection";

export default function Section() {
	return (
		<>
			<div className="mb-5">
				<SectionTitle title="categories" />
			</div>
			<SectionGrid />
			<ProductsSection />
		</>
	);
}
