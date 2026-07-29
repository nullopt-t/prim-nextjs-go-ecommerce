import Stars from "@/components/ui/stars";
import CustomButton from "@/components/ui/button";
import { Heart } from "lucide-react";
import { useTranslation } from "react-i18next";

export default function ProductsCard({ cardDetails }) {
	const { t } = useTranslation(["home", "common"]);

	return (
		<div className="shadow-lg cursor-pointer hover:scale-95 hover:border-accent-brand border-2 border-border rounded-md overflow-hidden">
			<div className="relative aspect-auto">
				<img
					src={cardDetails.img}
					alt=""
					className="object-center object-cover w-full h-full"
				/>
				<div className="absolute flex items-center justify-center top-2.5 right-2.5 group rounded-full w-8 h-8 bg-background text-foreground hover:bg-accent-brand hover:text-white">
					<Heart className="" />
				</div>
			</div>
			<div className="p-2">
				<p className="font-medium mb-1 text-foreground">
					{cardDetails.product}
				</p>
				<p className=""></p>
				<p className="flex items-center gap-2.5 mb-2.5">
					<span className="">
						<Stars starsNum={cardDetails.stars} />
					</span>
					<span className="text-muted-foreground">
						&#40;{cardDetails.reviews}&#41;
					</span>
				</p>
				<p className="flex flex-wrap gap-2.5 items-center">
					<span className="font-medium text-title-sm md:text-title-md text-foreground">
						{t("common:currency")}&nbsp;{cardDetails.price}
					</span>
					<span className="">
						<del className="text-muted-foreground">
							{t("common:currency")}&nbsp;{cardDetails.oldPrice}
						</del>
					</span>
					<span className="bg-[#d4183d] text-white rounded pr-1 pl-1">
						{cardDetails.discountPercentage}
					</span>
				</p>
			</div>
			<div className="h-10 m-2.5 text-primary-foreground bg-primary rounded-md hover:bg-accent hover:text-accent-foreground">
				<CustomButton text={t("product.addToCart")} />
			</div>
		</div>
	);
}
