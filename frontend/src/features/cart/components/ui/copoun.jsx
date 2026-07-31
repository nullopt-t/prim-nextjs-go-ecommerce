import CustomInput from "@/components/ui/input";
import CustomButton from "@/components/ui/button";
import { useTranslation } from "react-i18next";

function Copoun() {
	const { t } = useTranslation("cart");

	return (
		<div className="flex gap-10">
			<div className="flex-1 p-2.5 pl-5 border border-border rounded-md">
				<CustomInput
					type={"text"}
					placeholder={t("coupon.placeholder")}
					handle={() => null}
				/>
			</div>
			<div className="w-24 md:w-32 bg-primary text-primary-foreground rounded-md hover:bg-accent hover:text-accent-foreground">
				<CustomButton text={t("coupon.apply")} />
			</div>
		</div>
	);
}

export default Copoun;
