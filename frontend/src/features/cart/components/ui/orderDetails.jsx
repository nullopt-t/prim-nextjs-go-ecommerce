import ProductImage from "@/assets/imgs/placeholders/product.jpeg";
import { useTranslation } from "react-i18next";

function OrderDetails({ details }) {
	const { i18n } = useTranslation();

	return (
		<div className="flex gap-2.5 md:gap-5">
			<img
				src={ProductImage}
				alt="Order Image"
				className="rounded-md object-cover object-center aspect-square w-20 md:w-24"
			/>
			<div className="">
				<p className="font-medium text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-0.5">
					{i18n.resolvedLanguage === "en"
						? details.productName.en
						: details.productName.ar}
				</p>
				<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg mb-2.5">
					{details.productBrand}
				</p>
				<p className="font-medium text-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
					{details.productPrice}
				</p>
			</div>
		</div>
	);
}

export default OrderDetails;
