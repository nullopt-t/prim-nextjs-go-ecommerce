import { useTranslation } from "react-i18next";
import { Title } from "@/components/ui";

export default function BeliefsDescription() {
	const { t } = useTranslation("about");

	return (
		<div className="">
			<Title className="" title={t("belief.title")} />
			<p className="text-txt-sm md:text-txt-md lg:text-txt-lg text-muted-foreground">
				{t("belief.description")}
			</p>
		</div>
	);
}
