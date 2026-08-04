import ProductsGrid from "@/components/ui/productsGrid";
import SectionTitle from "@/features/home/components/ui/sectionTitle";
import Image from "@/assets/imgs/placeholders/product.jpeg";

export default function ProductsSection() {
	const cards = [
		{
			id: "pfd-1",
			img: Image,
			product: {
				en: "iphone 13 pro",
				ar: "آيفون 13 برو",
			},
			stars: "3",
			reviews: "256",
			price: "999",
			oldPrice: "1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-2",
			img: Image,
			product: {
				en: "iphone 13 pro",
				ar: "آيفون 13 برو",
			},
			stars: "3",
			reviews: "256",
			price: "999",
			oldPrice: "1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-3",
			img: Image,
			product: {
				en: "iphone 13 pro",
				ar: "آيفون 13 برو",
			},
			stars: "3",
			reviews: "256",
			price: "999",
			oldPrice: "1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-4",
			img: Image,
			product: {
				en: "iphone 13 pro",
				ar: "آيفون 13 برو",
			},
			stars: "3",
			reviews: "256",
			price: "999",
			oldPrice: "1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-5",
			img: Image,
			product: {
				en: "iphone 13 pro",
				ar: "آيفون 13 برو",
			},
			stars: "3",
			reviews: "256",
			price: "999",
			oldPrice: "1200",
			discountPercentage: "37%",
		},
	];

	return (
		<div className="">
			<div className="mb-2.5">
				<SectionTitle title="featuredProducts.title" />
			</div>
			<ProductsGrid
				products={cards}
				className="grid-cols-2 md:grid-cols-4 lg:grid-cols-5"
			/>
		</div>
	);
}
