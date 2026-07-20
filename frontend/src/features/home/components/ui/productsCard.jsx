import Stars from "@/features/home/components/ui/stars";
import CustomButton from "@/components/ui/button";

export default function ProductsCard({ cardDetails }) {
	return (
		<div className="cursor-pointer hover:scale-95 border border-border rounded-md overflow-hidden">
			<div className="aspect-auto">
				<img
					src={cardDetails.img}
					alt=""
					className="object-center object-cover w-full h-full"
				/>
			</div>
			<div className="p-2">
				<p className="font-medium mb-1">{cardDetails.product}</p>
				<p className=""></p>
				<p className="flex items-center gap-2.5 mb-2.5">
					<span className="">
						<Stars starsNum={cardDetails.stars} />
					</span>
					<span className="text-muted-foreground">
						&#40;{cardDetails.reviews}&#41;
					</span>
				</p>
				<p className="flex gap-2.5 items-center">
					<span className="font-medium text-title-sm md:text-title-md">
						{cardDetails.price}
					</span>
					<span className="">
						<del className="text-muted-foreground">
							{cardDetails.oldPrice}
						</del>
					</span>
					<span className="bg-[#d4183d] text-white rounded pr-1 pl-1">
						{cardDetails.discountPercentage}
					</span>
				</p>
			</div>
			<div className="h-10 m-2.5 text-primary-foreground bg-primary rounded-md">
				<CustomButton text="Add to cart" />
			</div>
		</div>
	);
}
