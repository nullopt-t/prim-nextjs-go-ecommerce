import Image from "@/assets/product.jpeg";
import Stars from "@/features/home/components/ui/stars";

export default function ProductsCard() {
	const cards = [
		{
			id: "pfd-1",
			img: "",
			product: "iphone 13 pro",
			stars: "3",
			reviews: "256",
			price: "$999",
			oldPrice: "$1200",
			discountPercentage: "37%",
		},
	];

	return (
		<div className="border border-border w-fit rounded-md overflow-hidden">
			<img
				src={Image}
				alt=""
				className="aspect-square object-center object-cover"
			/>
			<div className="p-2">
				<p className="font-medium mb-1">Iphone 13 pro</p>
				<p className=""></p>
				<p className="flex items-center gap-2.5 mb-2.5">
					<span className="">
						<Stars starsNum="2" />
					</span>
					<span className="text-muted-foreground">&#40;227&#41;</span>
				</p>
				<p className="flex gap-2.5 items-center">
					<span className="font-medium text-title-sm md:text-title-md">
						$999
					</span>
					<span className="">
						<del className="text-muted-foreground">$1200</del>
					</span>
					<span className="bg-[#d4183d] text-white rounded pr-1 pl-1">
						-37%
					</span>
				</p>
			</div>
		</div>
	);
}
