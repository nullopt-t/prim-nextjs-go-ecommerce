import { useTranslation } from "react-i18next";

function CartTitle() {
	const { t } = useTranslation("cart");

	return (
		<div className="">
			<p className="font-medium text-foreground text-title-sm md:text-title-md">
				{t("cart.title")}
			</p>
			<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
				{t("cart.items", { count: 3 })}
			</p>
		</div>
	);
}

export default CartTitle;
