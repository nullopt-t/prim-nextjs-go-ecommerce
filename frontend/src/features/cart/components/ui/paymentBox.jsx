import { CustomButton } from "@/components/ui";
import PaymentMethods from "@/features/cart/components/ui/paymentMethods";
import { useTranslation } from "react-i18next";

function PaymentBox() {
	const { t } = useTranslation("cart");

	return (
		<div className="border border-border p-2.5 md:p-5 rounded-md">
			<p className="mb-5 font-medium">{t("orderSummary.title")}</p>
			<div className="flex flex-col gap-2.5 mb-5">
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">
						{t("orderSummary.subtotal")}
					</span>
					<span className="text-foreground font-medium">$578.98</span>
				</p>
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">
						{t("orderSummary.shipping")}
					</span>
					<span className="text-foreground font-medium">
						{t("orderSummary.free")}
					</span>
				</p>
				<p className="flex justify-between text-txt-sm md:text-txt-sm lg:text-txt-sm">
					<span className="text-muted-foreground">
						{t("orderSummary.tax")}
					</span>
					<span className="text-foreground font-medium">$46.32</span>
				</p>
			</div>
			<hr className="text-border" />
			<div className="mt-5">
				<p className="flex justify-between text-title-sm md:text-title-sm">
					<span className="text-foreground font-medium">
						{t("orderSummary.total")}
					</span>
					<span className="text-foreground font-medium">$625.50</span>
				</p>
			</div>
			<div className="mt-5 h-12 bg-primary text-primary-foreground rounded-md hover:bg-accent hover:text-accent-foreground">
				<CustomButton text={t("orderSummary.checkout")} />
			</div>
			<div className="mt-5">
				<PaymentMethods />
			</div>
		</div>
	);
}

export default PaymentBox;
