import ProductsCard from "@/features/home/components/ui/productsCard";
import SectionTitle from "@/features/home/components/ui/sectionTitle";
import Image from "@/assets/product.jpeg";

export default function ProductsSection() {
	const cards = [
		{
			id: "pfd-1",
			img: Image,
			product: "iphone 13 pro",
			stars: "3",
			reviews: "256",
			price: "$999",
			oldPrice: "$1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-2",
			img: Image,
			product: "iphone 13 pro",
			stars: "5",
			reviews: "256",
			price: "$999",
			oldPrice: "$1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-3",
			img: Image,
			product: "iphone 13 pro",
			stars: "2",
			reviews: "256",
			price: "$999",
			oldPrice: "$1200",
			discountPercentage: "37%",
		},
		{
			id: "pfd-4",
			img: Image,
			product: "iphone 13 pro",
			stars: "1",
			reviews: "256",
			price: "$999",
			oldPrice: "$1200",
			discountPercentage: "37%",
		},
	];

	return (
		<div className="">
			<div className="mb-2.5">
				<SectionTitle title="Featured products" />
			</div>
			<div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-2.5 md:gap-5">
				{cards.map((value) => (
					<ProductsCard key={value.id} cardDetails={value} />
				))}
			</div>
		</div>
	);
}
