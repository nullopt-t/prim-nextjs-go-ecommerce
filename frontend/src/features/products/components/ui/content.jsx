import ProductsGrid from "@/components/ui/productsGrid";
import Image from "@/assets/imgs/placeholders/product.jpeg";

export default function Content() {
	const products = [
		{
			id: "pfd-sf1",
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
			id: "pfd-asfd1",
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
			id: "asfdas-1",
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
			id: "safsafda-1",
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
			id: "safsdafc-1",
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
		<ProductsGrid
			products={products}
			className="grid-cols-2 lg:grid-cols-4"
		/>
	);
}
