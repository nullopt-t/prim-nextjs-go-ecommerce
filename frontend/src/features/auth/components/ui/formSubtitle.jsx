import { useTranslation } from "react-i18next";

function FormSubtitle({ type }) {
	const { t } = useTranslation("auth");
	const subTitle =
		type === "login"
			? t("sign.subtitle")
			: type === "verify"
				? t("sign.subtitle")
				: "";

	return (
		<p className="text-muted-foreground text-txt-sm md:text-txt-md lg:text-txt-lg">
			{subTitle}
		</p>
	);
}

export default FormSubtitle;
