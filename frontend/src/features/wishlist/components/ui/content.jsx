import { Title, ProductsGrid, CustomButton } from "@/components/ui";
import Image from "@/assets/imgs/placeholders/product.jpeg";
import { useTranslation } from "react-i18next";

export default function Content() {
	const { t } = useTranslation("wishlist");

	const cards = [
		{
			id: "sgsbsaegsgsdag-1",
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
			id: "fdgsdagweqtgs-2",
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
			id: "sfasgsahtewhrw-3",
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
			id: "assaghrqwhehfqsa-4",
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
			id: "afsvasgwer-5",
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
			<div className="flex items-center justify-between">
				<Title
					title={t("title")}
					subtitle={t("savedItems", { count: 24 })}
				/>
				<div className="w-44 h-10 bg-primary text-primary-foreground hover:bg-accent hover:text-accent-foreground font-medium rounded-md">
					<CustomButton text={t("addToCart")} onClick={() => {}} />
				</div>
			</div>
			<ProductsGrid
				products={cards}
				className="grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
				isWishlist={true}
			/>
		</div>
	);
}
